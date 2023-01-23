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
	configuration2 "github.com/th2-net/th2-common-go/schema/queue/configuration"
	"log"
)

type ConnectionManager struct {
	QConfig      *configuration2.MessageRouterConfiguration
	MqConnConfig *configuration2.RabbitMQConfiguration
	Url          string
	Publisher    Publisher
	Consumer     Consumer
}

func (manager *ConnectionManager) Construct() {
	manager.Publisher = Publisher{url: manager.Url}
	manager.Publisher.connect()

	manager.Consumer = Consumer{url: manager.Url, channels: make(map[string]*amqp.Channel)}
	manager.Consumer.connect()
}

func (manager *ConnectionManager) Close() error {
	err := manager.Publisher.conn.Close()
	if err != nil {
		return err
	}

	fail := manager.Consumer.conn.Close()
	if fail != nil {
		return fail
	}

	if len(manager.Consumer.channels) != 0 {
		for _, ch := range manager.Consumer.channels {
			ch.Close()
		}
	}

	log.Println("Connection Closed gracefully")
	return nil
}

type DeliveryConfirmation struct {
	Delivery *amqp.Delivery
}

func (dc DeliveryConfirmation) Confirm() error {
	err := dc.Delivery.Ack(false)
	log.Println("Acknowledged")
	if err != nil {
		log.Fatalf("error during confirmation: %v \n", err)
		return err
	}
	return nil
}
func (dc DeliveryConfirmation) Reject() error {
	err := dc.Delivery.Reject(false)
	if err != nil {
		return err
	}
	return nil
}
