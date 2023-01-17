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
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

type Consumer struct {
	url      string
	conn     *amqp.Connection
	channels map[string]*amqp.Channel
	Logger   zerolog.Logger
}

func (cns *Consumer) connect() {
	conn, err := amqp.Dial(cns.url)
	if err != nil {
		cns.Logger.Fatal().Err(err).Send()
	}
	cns.conn = conn
	cns.Logger.Debug().Msg("Consumer connected")
}

func (cns *Consumer) Consume(queueName string, handler func(delivery amqp.Delivery)) error {
	ch, err := cns.conn.Channel()
	if err != nil {
		cns.Logger.Fatal().Err(err).Send()
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
			handler(d)
		}
	}()

	return nil
}

func (cns *Consumer) ConsumeWithManualAck(queueName string, handler func(msgDelivery amqp.Delivery)) error {
	ch, err := cns.conn.Channel()
	if err != nil {
		cns.Logger.Fatal().Err(err).Send()
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
			handler(d)
		}
	}()

	return nil
}
