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
	"github.com/streadway/amqp"
	"log"
)

type Consumer struct {
	url      string
	conn     *amqp.Connection
	channels map[string]*amqp.Channel
}

func (manager *Consumer) connect() {
	conn, err := amqp.Dial(manager.url)
	if err != nil {
		log.Fatalln(err)
	}
	manager.conn = conn
}

func (manager *Consumer) Consume(queueName string, handler func(delivery amqp.Delivery)) error {
	ch, err := manager.conn.Channel()
	if err != nil {
		log.Fatalln(err)
	}
	manager.channels[queueName] = ch

	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		log.Println(err)
		return err
	}

	go func() {
		log.Printf("in queue %v \n", queueName)
		for d := range msgs {
			handler(d)
		}
	}()

	return nil
}

func (manager *Consumer) ConsumeWithManualAck(queueName string, handler func(msgDelivery amqp.Delivery)) error {
	ch, err := manager.conn.Channel()
	if err != nil {
		log.Fatalln(err)
	}
	manager.channels[queueName] = ch
	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		log.Println(err)
		return err
	}
	go func() {
		log.Printf("in queue %v \n", queueName)
		for d := range msgs {
			handler(d)
		}
	}()

	return nil
}
