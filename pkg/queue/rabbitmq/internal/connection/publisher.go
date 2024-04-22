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
	"context"
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/pkg/metrics"
	connCfg "github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/connection"
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
	*connectionHolder
	Logger zerolog.Logger
}

func NewPublisher(url string, configuration connCfg.Config, logger zerolog.Logger) (Publisher, error) {
	if url == "" {
		return Publisher{}, errors.New("url is not set")
	}
	publisher := Publisher{
		Logger: logger,
	}
	c, err := newConnection(url, "publisher", logger, nil, nil)
	if err != nil {
		return publisher, err
	}
	publisher.connectionHolder = c
	logger.Debug().Msg("Publisher connected")
	return publisher, nil
}

func (pb *Publisher) Publish(body []byte, routingKey string, exchange string, th2Pin string, th2Type string) error {

	ch, err := pb.getChannel(routingKey)
	if err != nil {
		return err
	}

	// Ideally, the context should be passed from outside
	// but this is breaking API change and we cannot do that
	publError := ch.PublishWithContext(context.Background(), exchange, routingKey, false, false, amqp.Publishing{Body: body})
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
