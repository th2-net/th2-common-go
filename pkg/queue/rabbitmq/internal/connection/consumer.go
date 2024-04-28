/*
 * Copyright 2022-2023 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package connection

import (
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/pkg/metrics"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/connection"
	"time"
)

var th2RabbitmqMessageSizeSubscribeBytes = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_rabbitmq_message_size_subscribe_bytes",
		Help: "Amount of bytes received",
	},
	metrics.SubscriberLabels,
)

var th2RabbitmqMessageProcessDurationSeconds = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "th2_rabbitmq_message_process_duration_seconds",
		Help:    "Subscriber's handling process duration",
		Buckets: metrics.DefaultBuckets,
	},
	metrics.SubscriberLabels,
)

type Consumer struct {
	*connectionHolder
	Logger                          zerolog.Logger
	maxMissingQueueRecoveryAttempts int
}

func NewConsumer(url string, configuration connection.Config, componentName string, logger zerolog.Logger) (Consumer, error) {
	if url == "" {
		return Consumer{}, errors.New("url is empty")
	}
	maxMissingQueueRecoveryAttempts := defaultMaxRecoveryAttempts
	if configuration.MaxRecoveryAttempts > 0 {
		maxMissingQueueRecoveryAttempts = configuration.MaxRecoveryAttempts
	}
	consumer := Consumer{
		Logger:                          logger,
		maxMissingQueueRecoveryAttempts: maxMissingQueueRecoveryAttempts,
	}
	conn, err := newConnection(url, fmt.Sprintf("%s_consumer", componentName),
		logger, configuration, nil, nil)
	if err != nil {
		return consumer, err
	}
	consumer.connectionHolder = conn
	logger.Debug().Msg("Consumer connected")
	return consumer, nil
}

func (cns *Consumer) Consume(queueName string, th2Pin string, th2Type string, handler func(delivery amqp.Delivery) error) error {
	return cns.consume(
		queueName, th2Pin, th2Type, cns.consumeWithAutoAck, "consume",
		func(delivery amqp.Delivery, timer *prometheus.Timer) error {
			defer timer.ObserveDuration()
			return handler(delivery)
		},
	)
}

func (cns *Consumer) ConsumeWithManualAck(queueName string, th2Pin string, th2Type string, handler func(msgDelivery amqp.Delivery, timer *prometheus.Timer) error) error {
	return cns.consume(
		queueName, th2Pin, th2Type, cns.consumeWithManualAck, "consumeWithManualAck",
		func(delivery amqp.Delivery, timer *prometheus.Timer) error {
			return handler(delivery, timer)
		},
	)
}

func (cns *Consumer) consume(queueName string, th2Pin string, th2Type string,
	subscriptionProducer func(queueName string, methodName string) (*amqp.Channel, <-chan amqp.Delivery, error),
	methodName string, handler func(delivery amqp.Delivery, timer *prometheus.Timer) error) error {
	ch, msgs, err := subscriptionProducer(queueName, methodName)
	if err != nil {
		return err
	}

	go func() {
		cns.Logger.Debug().
			Str("method", methodName).
			Str("queue", queueName).
			Msg("start handling messages")
		running := true
		durationObserver := th2RabbitmqMessageProcessDurationSeconds.WithLabelValues(th2Pin, th2Type, queueName)
		messageSizeObserver := th2RabbitmqMessageSizeSubscribeBytes.WithLabelValues(th2Pin, th2Type, queueName)
		handleDelivery := func(d amqp.Delivery) {
			timer := prometheus.NewTimer(durationObserver)
			cns.Logger.Trace().
				Str("exchange", d.Exchange).
				Str("routing", d.RoutingKey).
				Int("bodySize", len(d.Body)).
				Msg("receive delivery")
			if err := handler(d, timer); err != nil {
				cns.Logger.Error().
					Err(err).
					Str("exchange", d.Exchange).
					Str("routing", d.RoutingKey).
					Int("bodySize", len(d.Body)).
					Msg("Cannot handle delivery")
			}
			messageSizeObserver.Add(float64(len(d.Body)))
		}
		deliveries := msgs
		chErrors := ch.NotifyClose(make(chan *amqp.Error))
		for running {
			select {
			case _, ok := <-cns.done:
				if !ok {
					running = false
					// drain messages
					for d := range deliveries {
						handleDelivery(d)
					}
				}
			case chErr, ok := <-chErrors:
				if !ok {
					break
				}
				cns.Logger.Error().
					Err(chErr).
					Str("queue", queueName).
					Msg("consumer error")
				// drain messages
				for d := range deliveries {
					handleDelivery(d)
				}
				ch, deliveries, err = subscriptionProducer(queueName, methodName)
				if err != nil {
					if errors.Is(err, amqp.ErrClosed) {
						break
					}
					// TODO: decide what to do in this case
					panic(fmt.Sprintf("failure during consumer %s recovery: %v", methodName, err))
				}
				chErrors = ch.NotifyClose(make(chan *amqp.Error))
				cns.Logger.Info().
					Str("queue", queueName).
					Msg("consumer channel recovered")
			case d, ok := <-deliveries:
				if !ok {
					break
				}
				handleDelivery(d)
			}
		}
		cns.Logger.Debug().
			Str("method", methodName).
			Str("queue", queueName).
			Msg("stop handling messages")
	}()

	return nil
}

func (cns *Consumer) consumeWithManualAck(queueName string, methodName string) (*amqp.Channel, <-chan amqp.Delivery, error) {
	return cns.consumeFromQueue(queueName, methodName, false)
}

func (cns *Consumer) consumeWithAutoAck(queueName string, methodName string) (*amqp.Channel, <-chan amqp.Delivery, error) {
	return cns.consumeFromQueue(queueName, methodName, true)
}

func (cns *Consumer) consumeFromQueue(queueName string, methodName string, autoAck bool) (*amqp.Channel, <-chan amqp.Delivery, error) {
	attempts := 0
	timeout := cns.minRecoveryTimeout
	for {
		ch, err := cns.getChannel(queueName)
		if err != nil {
			return nil, nil, err
		}

		msgs, err := cns.startConsuming(ch, queueName, autoAck)
		if err != nil {
			var amqpErr *amqp.Error
			isAmqpErr := errors.As(err, &amqpErr)
			if !isAmqpErr || amqpErr.Code != amqp.NotFound {
				cns.Logger.Error().
					Err(err).
					Str("method", methodName).
					Str("queue", queueName).
					Msg("consuming error")
				return nil, nil, err
			}
			if attempts >= cns.maxMissingQueueRecoveryAttempts {
				return nil, nil, err
			}
			cns.Logger.Warn().
				Str("method", methodName).
				Str("queue", queueName).
				Int("attempts", attempts).
				Dur("timeout", timeout).
				Msg("queue is not found. Retry after timeout")
			time.Sleep(timeout)
			timeout *= 2
			if timeout > cns.maxRecoveryTimeout {
				timeout = cns.maxRecoveryTimeout
			}
			attempts += 1
			continue
		}
		return ch, msgs, nil
	}
}

func (cns *Consumer) startConsuming(ch *amqp.Channel, queueName string, autoAck bool) (<-chan amqp.Delivery, error) {
	msgs, err := ch.Consume(
		queueName, // queue
		// TODO: we need to provide a name that will help to identify the component
		"",      // consumer
		autoAck, // auto-ack
		false,   // exclusive
		false,   // no-local
		false,   // no-wait
		nil,     // args
	)
	if err != nil {
		return nil, err
	}
	return msgs, nil
}
