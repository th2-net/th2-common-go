/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
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
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/th2-net/th2-common-go/test/modules/rabbitmq"
	"os"
	"testing"
	"time"
)

var consumerLogger = zerolog.New(os.Stdout).With().Str("test_type", "consumer").Logger()

func TestConsumer_Consume(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	config := rabbitmq.StartMq(t, "test")

	manager, err := NewConnectionManager(config, consumerLogger)
	if err != nil {
		t.Fatal(err)
	}
	go manager.ListenForBlockingNotifications()
	defer manager.Close()
	conn, err := rabbitmq.RawAmqp(t, config, true)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	queue := conn.CreateQueue("test")
	routingKey := "test-publish"
	conn.BindQueue(config, queue, routingKey)

	deliveries := make(chan []byte, 1)
	err = manager.Consumer.Consume(queue.Name, "pin", "test", func(delivery amqp.Delivery) error {
		deliveries <- delivery.Body
		close(deliveries)
		return nil
	})
	if err != nil {
		t.Fatal("cannot start consuming")
	}

	conn.Publish(config, routingKey, []byte("hello"))

	select {
	case d := <-deliveries:
		assert.Equal(t, "hello", string(d))
	case <-time.After(1 * time.Second):
		t.Fatal("didn't receive the data withing 1 second")
	}
}
