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

var th2_rabbitmq_message_size_publish_bytes = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_rabbitmq_message_size_publish_bytes",
		Help: "Amount of bytes sent",
	},
	metrics.SENDER_LABELS,
)

var th2_rabbitmq_message_publish_total = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_rabbitmq_message_publish_total",
		Help: "Amount of batches sent",
	},
	metrics.SENDER_LABELS,
)

type Publisher struct {
	url  string
	conn *amqp.Connection

	Logger zerolog.Logger
}

func (pb *Publisher) connect() error {
	conn, err := amqp.Dial(pb.url)
	if err != nil {
		return err
	}
	pb.conn = conn
	pb.Logger.Debug().Msg("Publisher connected")
	return nil
}

func (pb *Publisher) Publish(body []byte, routingKey string, exchange string, th2Pin string, th2Type string) error {
	ch, err := pb.conn.Channel()
	if err != nil {
		pb.Logger.Error().Err(err).Send()
		return err
	}
	defer ch.Close()
	publError := ch.Publish(exchange, routingKey, false, false, amqp.Publishing{Body: body})
	if publError != nil {
		return err
	}
	pb.Logger.Trace().Msg(" [x] Sent ")
	th2_rabbitmq_message_size_publish_bytes.WithLabelValues(th2Pin, th2Type, exchange, routingKey).Add(float64(len(body)))
	th2_rabbitmq_message_publish_total.WithLabelValues(th2Pin, th2Type, exchange, routingKey).Inc()

	return nil
}
