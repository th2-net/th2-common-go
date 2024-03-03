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

package rabbitmq

import (
	"context"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/connection"
	"google.golang.org/protobuf/proto"
	"testing"
	"time"
)

const (
	containerName = "rabbitmq-connection-test"
	mqPort        = "5672"
	TestBook      = "test_book"
	TestScope     = "test_scope"
)

func StartMq(t *testing.T, exchange string) connection.Config {
	return StartMqWithContainerName(t, "", exchange)
}

func StartMqWithContainerName(t *testing.T, containerName, exchange string) connection.Config {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Name:         containerName,
		Image:        "rabbitmq:3.10",
		ExposedPorts: []string{mqPort},
		WaitingFor:   wait.ForLog("Server startup complete"),
	}
	rabbit, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
		Reuse:            containerName != "",
	})
	if err != nil {
		t.Fatal("cannot create container", err)
	}
	t.Cleanup(func() {
		err := rabbit.Terminate(ctx)
		if err != nil {
			t.Logf("cannot stop rabbitmq container: %v", err)
		}
	})
	host, err := rabbit.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := rabbit.MappedPort(ctx, mqPort)
	if err != nil {
		t.Fatal(err)
	}
	return connection.Config{
		Port:         port.Int(),
		Host:         host,
		ExchangeName: exchange,
		Username:     "guest",
		Password:     "guest",
	}
}

type RawAmqpHolder struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	t    *testing.T
}

func RawAmqp(t *testing.T, config connection.Config, createExchange bool) (*RawAmqpHolder, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		config.Username, config.Password, config.Host, config.Port, config.VHost))
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if createExchange {
		err = ch.ExchangeDeclare(
			config.ExchangeName,
			"direct",
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return nil, err
		}
	}
	return &RawAmqpHolder{
		conn: conn,
		ch:   ch,
		t:    t,
	}, nil
}

func (h RawAmqpHolder) Close() {
	if err := h.conn.Close(); err != nil {
		h.t.Error("cannot close connection", err)
	}
}

func (h RawAmqpHolder) CreateQueue(name string) amqp.Queue {
	queue, err := h.ch.QueueDeclare(
		name,
		false, // durable
		false, // auto delete
		false, // exclusive
		false,
		nil,
	)
	if err != nil {
		h.t.Fatal("cannot declare queue", err)
	}
	return queue
}

func (h RawAmqpHolder) BindQueue(connCfg connection.Config, queue amqp.Queue, bindings ...string) {

	for _, binding := range bindings {
		err := h.ch.QueueBind(
			queue.Name,
			binding,
			connCfg.ExchangeName,
			false,
			nil,
		)
		if err != nil {
			h.t.Fatal("cannot bind", err)
		}
	}
}

func (h RawAmqpHolder) Publish(connCfg connection.Config, routingKey string, data []byte) {

	err := h.ch.Publish(connCfg.ExchangeName, routingKey, true, false, amqp.Publishing{
		Body: data,
	})
	if err != nil {
		h.t.Fatal("cannot publish", err)
	}
}

func (h RawAmqpHolder) Consume(queue amqp.Queue) <-chan amqp.Delivery {

	deliveries, err := h.ch.Consume(
		queue.Name,
		fmt.Sprintf("test-consumer-%d", time.Now().UnixNano()),
		true, false,
		false, false, amqp.Table{})
	if err != nil {
		h.t.Fatal("cannot consume", err)
	}
	return deliveries
}

type GenericListener[T any] struct {
	Channel chan *T
}

func (l *GenericListener[T]) Handle(_ queue.Delivery, batch *T) error {
	l.Channel <- batch
	return nil
}

func (l *GenericListener[T]) OnClose() error {
	return nil
}

type GenericManualListener[T any] struct {
	Channel        chan *T
	OnConfirmation func(confirmation queue.Confirmation)
}

func (l *GenericManualListener[T]) Handle(_ queue.Delivery, batch *T, confirm queue.Confirmation) error {
	l.Channel <- batch
	l.OnConfirmation(confirm)
	return nil
}

func (l *GenericManualListener[T]) OnClose() error {
	return nil
}

func Reject(confirmation queue.Confirmation) {
	confirmation.Reject()
}

func Confirm(confirmation queue.Confirmation) {
	confirmation.Confirm()
}

func CheckReceiveDelivery[T proto.Message](t *testing.T, deliveries <-chan amqp.Delivery, originalBatch T) {
	select {
	case d := <-deliveries:
		receivedBatch := proto.Clone(originalBatch)
		proto.Reset(receivedBatch)
		if err := proto.Unmarshal(d.Body, receivedBatch); err != nil {
			t.Fatal("cannot deserialize data")
		}

		assert.True(t, proto.Equal(originalBatch, receivedBatch))
	case <-time.After(1 * time.Second):
		t.Fatal("didn't receive the delivery")

	}
}

func CheckReceiveBatch[T proto.Message](t *testing.T, deliveries <-chan T, originalBatch T) {
	select {
	case d := <-deliveries:
		assert.True(t, proto.Equal(originalBatch, d))
	case <-time.After(1 * time.Second):
		t.Fatal("didn't receive the delivery")
	}
}

func CheckReceiveBytes(t *testing.T, deliveries <-chan amqp.Delivery, originalBatch []byte) {
	select {
	case d := <-deliveries:
		assert.Equal(t, originalBatch, d.Body)
	case <-time.After(1 * time.Second):
		t.Fatal("didn't receive the delivery")
	}
}
