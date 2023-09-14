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
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/th2-net/th2-common-go/test/modules/rabbitmq"
	"os"
	"testing"
	"time"
)

var publisherLogger = zerolog.New(os.Stdout).With().Str("test_type", "publisher").Logger()

func TestPublisher(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	config := rabbitmq.StartMq(t, "test")

	manager, err := NewConnectionManager(config, publisherLogger)
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()
	conn, err := rabbitmq.RawAmqp(t, config, true)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	queue := conn.CreateQueue("test")
	routingKey := "test-publish"
	conn.BindQueue(config, queue, routingKey)

	err = manager.Publisher.Publish([]byte("hello"), routingKey, config.ExchangeName, "test", "msg")
	if err != nil {
		t.Fatal("cannot publish message")
	}

	select {
	case d := <-conn.Consume(queue):
		assert.Equal(t, "hello", string(d.Body))
	case <-time.After(1 * time.Second):
		t.Fatal("didn't receive any delivery during 1 second")
	}
}
