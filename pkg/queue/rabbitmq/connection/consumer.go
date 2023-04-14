/*
 * Copyright 2022-2023 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"github.com/th2-net/th2-common-go/pkg/metrics"
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
	url      string
	conn     *amqp.Connection
	channels map[string]*amqp.Channel
	Logger   zerolog.Logger
}

func NewConsumer(url string, logger zerolog.Logger) (Consumer, error) {
	if url == "" {
		return Consumer{}, errors.New("url is empty")
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return Consumer{}, err
	}
	logger.Debug().Msg("Consumer connected")
	return Consumer{
		url:      url,
		conn:     conn,
		channels: make(map[string]*amqp.Channel),
		Logger:   logger,
	}, nil
}

func (cns *Consumer) Close() error {
	return cns.conn.Close()
}

func (cns *Consumer) Consume(queueName string, th2Pin string, th2Type string, handler func(delivery amqp.Delivery) error) error {
	ch, err := cns.conn.Channel()
	if err != nil {
		cns.Logger.Error().
			Err(err).
			Str("queue", queueName).
			Msg("cannot open channel")
		return err
	}
	cns.channels[queueName] = ch

	msgs, consErr := ch.Consume(
		queueName, // queue
		// TODO: we need to provide a name that will help to identify the component
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if consErr != nil {
		cns.Logger.Error().
			Err(err).
			Str("method", "consume").
			Str("queue", queueName).
			Msg("Consuming error")
		return consErr
	}

	go func() {
		cns.Logger.Debug().
			Str("method", "consume").
			Str("queue", queueName).
			Msg("start handling messages")
		for d := range msgs {
			timer := prometheus.NewTimer(th2RabbitmqMessageProcessDurationSeconds.WithLabelValues(th2Pin, th2Type, queueName))
			cns.Logger.Trace().
				Str("exchange", d.Exchange).
				Str("routing", d.RoutingKey).
				Int("bodySize", len(d.Body)).
				Msg("receive delivery")
			if err := handler(d); err != nil {
				cns.Logger.Error().
					Err(err).
					Str("exchange", d.Exchange).
					Str("routing", d.RoutingKey).
					Int("bodySize", len(d.Body)).
					Msg("Cannot handle delivery")
			}
			timer.ObserveDuration()
			th2RabbitmqMessageSizeSubscribeBytes.WithLabelValues(th2Pin, th2Type, queueName).Add(float64(len(d.Body)))
		}
		cns.Logger.Debug().
			Str("method", "consume").
			Str("queue", queueName).
			Msg("stop handling messages")
	}()

	return nil
}

func (cns *Consumer) ConsumeWithManualAck(queueName string, th2Pin string, th2Type string, handler func(msgDelivery amqp.Delivery, timer *prometheus.Timer) error) error {
	ch, err := cns.conn.Channel()
	if err != nil {
		cns.Logger.Error().
			Err(err).
			Str("queue", queueName).
			Msg("cannot open channel")
		return err
	}
	cns.channels[queueName] = ch
	msgs, consErr := ch.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if consErr != nil {
		cns.Logger.Error().
			Err(err).
			Str("method", "consumeWithManualAck").
			Str("queue", queueName).
			Msg("Consuming error")
		return consErr
	}
	go func() {
		cns.Logger.Debug().
			Str("method", "consumeWithManualAck").
			Str("queue", queueName).
			Msg("start handling messages")
		for d := range msgs {
			timer := prometheus.NewTimer(th2RabbitmqMessageProcessDurationSeconds.WithLabelValues(th2Pin, th2Type, queueName))
			if err := handler(d, timer); err != nil {
				cns.Logger.Error().
					Err(err).
					Str("exchange", d.Exchange).
					Str("routing", d.RoutingKey).
					Int("bodySize", len(d.Body)).
					Msg("cannot handle delivery")
			}
			th2RabbitmqMessageSizeSubscribeBytes.WithLabelValues(th2Pin, th2Type, queueName).Add(float64(len(d.Body)))
		}
		cns.Logger.Debug().
			Str("method", "consumeWithManualAck").
			Str("queue", queueName).
			Msg("stop handling messages")
	}()

	return nil
}
