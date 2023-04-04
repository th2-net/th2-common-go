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
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	configuration2 "github.com/th2-net/th2-common-go/schema/queue/configuration"
)

type ConnectionManager struct {
	QConfig      *configuration2.MessageRouterConfiguration
	MqConnConfig *configuration2.RabbitMQConfiguration
	Url          string
	Publisher    Publisher
	Consumer     Consumer

	Logger zerolog.Logger
}

func (manager *ConnectionManager) Construct() {
	manager.Publisher = Publisher{url: manager.Url, Logger: zerolog.New(os.Stdout).With().Str("component", "publisher").Timestamp().Logger()}
	manager.Publisher.connect()

	manager.Consumer = Consumer{url: manager.Url, channels: make(map[string]*amqp.Channel),
		Logger: zerolog.New(os.Stdout).With().Str("component", "subscriber").Timestamp().Logger()}
	manager.Consumer.connect()
}

func (manager *ConnectionManager) Close() error {
	err := manager.Publisher.conn.Close()
	if err != nil {
		manager.Logger.Error().Err(err).Send()
		return err
	}

	fail := manager.Consumer.conn.Close()
	if fail != nil {
		manager.Logger.Error().Err(err).Send()
		return fail
	}

	if len(manager.Consumer.channels) != 0 {
		for _, ch := range manager.Consumer.channels {
			ch.Close()
		}
		manager.Logger.Debug().Msg("Channels Closed")

	}
	manager.Logger.Info().Msg("Connection Closed gracefully")
	return nil
}

type DeliveryConfirmation struct {
	Delivery *amqp.Delivery
	Timer    *prometheus.Timer

	Logger zerolog.Logger
}

func (dc DeliveryConfirmation) Confirm() error {
	err := dc.Delivery.Ack(false)
	if err != nil {
		dc.Logger.Error().Err(err).Str("Method", "Confirm").Msg("Error during Acknowledgment")
		return err
	}
	dc.Logger.Info().Str("Method", "Confirm").Msg("Acknowledged")
	dc.Timer.ObserveDuration()
	return nil
}
func (dc DeliveryConfirmation) Reject() error {
	err := dc.Delivery.Reject(false)
	if err != nil {
		dc.Logger.Error().Err(err).Str("Method", "Reject").Msg("Error during Rejection")
		return err
	}
	dc.Logger.Info().Str("Method", "Reject").Msg("Rejected")
	dc.Timer.ObserveDuration()
	return nil
}
