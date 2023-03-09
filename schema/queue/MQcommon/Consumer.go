/*
 * Copyright 2022 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package MQcommon

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"github.com/th2-net/th2-common-go/schema/metrics"
)

var th2_rabbitmq_message_size_subscribe_bytes = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_rabbitmq_message_size_subscribe_bytes",
		Help: "Amount of bytes received",
	},
	metrics.SUBSCRIBER_LABELS,
)

var th2_rabbitmq_message_process_duration_seconds = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "th2_rabbitmq_message_process_duration_seconds",
		Help:    "Subscriber's handling process duration",
		Buckets: metrics.DEFAULT_BUCKETS,
	},
	metrics.SUBSCRIBER_LABELS,
)

type Consumer struct {
	url      string
	conn     *amqp.Connection
	channels map[string]*amqp.Channel
	Logger   zerolog.Logger
}

func (cns *Consumer) connect() error {
	conn, err := amqp.Dial(cns.url)
	if err != nil {
		return err
	}
	cns.conn = conn
	cns.Logger.Debug().Msg("Consumer connected")
	return nil
}

func (cns *Consumer) Consume(queueName string, th2Pin string, th2Type string, handler func(delivery amqp.Delivery) error) error {
	ch, err := cns.conn.Channel()
	if err != nil {
		return err
	}
	cns.channels[queueName] = ch

	msgs, consErr := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if consErr != nil {
		cns.Logger.Error().Err(err).Msg("Consuming error")
		return consErr
	}

	go func() {
		cns.Logger.Debug().Msgf("Consumed messages will handled from queue %s", queueName)
		for d := range msgs {
			timer := prometheus.NewTimer(th2_rabbitmq_message_process_duration_seconds.WithLabelValues(th2Pin, th2Type, queueName))
			if err := handler(d); err != nil {
				cns.Logger.Error().Err(err).
					Str("exchange", d.Exchange).
					Str("routing", d.RoutingKey).
					Int("bodySize", len(d.Body)).
					Msg("Cannot handle delivery")
			}
			timer.ObserveDuration()
			th2_rabbitmq_message_size_subscribe_bytes.WithLabelValues(th2Pin, th2Type, queueName).Add(float64(len(d.Body)))
		}
	}()

	return nil
}

func (cns *Consumer) ConsumeWithManualAck(queueName string, th2Pin string, th2Type string, handler func(msgDelivery amqp.Delivery, timer *prometheus.Timer) error) error {
	ch, err := cns.conn.Channel()
	if err != nil {
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
		cns.Logger.Error().Err(err).Msg("Consuming error")
		return consErr
	}
	go func() {
		cns.Logger.Debug().Msgf("Consumed messages will handled from queue %s", queueName)
		for d := range msgs {
			timer := prometheus.NewTimer(th2_rabbitmq_message_process_duration_seconds.WithLabelValues(th2Pin, th2Type, queueName))
			if err := handler(d, timer); err != nil {
				cns.Logger.Error().Err(err).
					Str("exchange", d.Exchange).
					Str("routing", d.RoutingKey).
					Int("bodySize", len(d.Body)).
					Msg("cannot handle delivery")
			}
			th2_rabbitmq_message_size_subscribe_bytes.WithLabelValues(th2Pin, th2Type, queueName).Add(float64(len(d.Body)))
		}
	}()

	return nil
}
