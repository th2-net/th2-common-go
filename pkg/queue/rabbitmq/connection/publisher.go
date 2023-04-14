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
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"github.com/th2-net/th2-common-go/pkg/metrics"
)

var th2RabbitmqMessageSizePublishBytes = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_rabbitmq_message_size_publish_bytes",
		Help: "Amount of bytes sent",
	},
	metrics.SenderLabels,
)

var th2RabbitmqMessagePublishTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_rabbitmq_message_publish_total",
		Help: "Amount of batches sent",
	},
	metrics.SenderLabels,
)

type Publisher struct {
	url      string
	conn     *amqp.Connection
	channels map[string]*amqp.Channel
	mutex    *sync.Mutex

	Logger zerolog.Logger
}

func NewPublisher(url string, logger zerolog.Logger) (Publisher, error) {
	if url == "" {
		return Publisher{}, errors.New("url is not set")
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return Publisher{}, err
	}
	logger.Debug().Msg("Publisher connected")
	return Publisher{
		url:      url,
		conn:     conn,
		channels: make(map[string]*amqp.Channel),
		mutex:    &sync.Mutex{},
		Logger:   logger,
	}, nil
}

func (pb *Publisher) Publish(body []byte, routingKey string, exchange string, th2Pin string, th2Type string) error {

	ch, err := pb.getChannel(routingKey)

	publError := ch.Publish(exchange, routingKey, false, false, amqp.Publishing{Body: body})
	if publError != nil {
		pb.Logger.Error().Err(publError).Send()
		return err
	}
	bodySize := len(body)
	pb.Logger.Trace().Int("size", bodySize).Msg("data published")
	th2RabbitmqMessageSizePublishBytes.WithLabelValues(th2Pin, th2Type, exchange, routingKey).Add(float64(bodySize))
	th2RabbitmqMessagePublishTotal.WithLabelValues(th2Pin, th2Type, exchange, routingKey).Inc()

	return nil
}

func (pb *Publisher) Close() error {
	return pb.conn.Close()
}

func (pb *Publisher) getChannel(routingKey string) (*amqp.Channel, error) {
	var ch *amqp.Channel
	var err error
	var exists bool
	ch, exists = pb.channels[routingKey]
	if !exists {
		ch, err = pb.getOrCreateChannel(routingKey)
	}

	return ch, err
}

func (pb *Publisher) getOrCreateChannel(routingKey string) (*amqp.Channel, error) {
	pb.mutex.Lock()
	defer pb.mutex.Unlock()
	var ch *amqp.Channel
	var err error
	var exists bool
	ch, exists = pb.channels[routingKey]
	if !exists {
		ch, err = pb.conn.Channel()
		if err != nil {
			return nil, err
		}
		pb.channels[routingKey] = ch
	}
	return ch, nil
}
