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

type Publisher struct {
	url  string
	conn *amqp.Connection
}

func (pb *Publisher) connect() {
	conn, err := amqp.Dial(pb.url)
	if err != nil {
		log.Fatalln(err)
	}
	pb.conn = conn

}

func (pb *Publisher) Publish(body []byte, routingKey string, exchange string) error {
	ch, err := pb.conn.Channel()
	if err != nil {
		log.Println(err)
		return err
	}
	defer ch.Close()
	fail := ch.Publish(exchange, routingKey, false, false, amqp.Publishing{Body: body})
	if fail != nil {
		log.Println(err)
		return err
	}
	log.Printf(" [x] Sent %s", body)

	return nil
}
