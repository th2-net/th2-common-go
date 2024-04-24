/*
 * Copyright 2024 Exactpro (Exactpro Systems Limited)
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

package message

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/connection"
	rabbitmqSupport "github.com/th2-net/th2-common-go/test/modules/rabbitmq"
	"io"
	"testing"
	"time"
)

func TestPublisherReconnects(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	containerName := fmt.Sprintf("reconnect-test-%d", time.Now().UTC().UnixNano())
	ctx := context.Background()
	port := fmt.Sprintf("%d:%s", 9900, rabbitmqSupport.MqPort)
	container := rabbitmqSupport.CreateMqContainer(ctx, t, containerName, port)
	err := container.Start(ctx)
	if err != nil {
		t.Fatal("cannot start container:", err)
	}
	t.Cleanup(func() {
		err := container.Terminate(ctx)
		if err != nil {
			t.Logf("cannot terminate container: %v", err)
		}
	})
	config := rabbitmqSupport.GetConfigForContainer(ctx, t, container, "test")

	routingKey := setupMq(t, config)

	router, _, manager, err := rabbitmq.NewRouters(config, &queue.RouterConfig{
		Queues: map[string]queue.DestinationConfig{
			"publish-pin1": {
				Exchange:   config.ExchangeName,
				RoutingKey: routingKey,
				Attributes: []string{"publish", "test", "unique"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func(manager io.Closer) {
		err := manager.Close()
		if err != nil {
			t.Logf("cannot close manager connection: %v", err)
		}
	}(manager)

	err = router.SendRawAll([]byte("hello"), "unique")
	if err != nil {
		t.Fatal("cannot send message:", err)
	}

	err = container.Terminate(ctx)
	if err != nil {
		t.Fatal("cannot stop container:", err)
	}

	// create new container with same port
	// cannot Stop and Start container because of wait strategy
	go func() {
		time.Sleep(7 * time.Second)
		container = rabbitmqSupport.CreateMqContainer(ctx, t, containerName, port)
		err = container.Start(ctx)
		if err != nil {
			t.Error("cannot start container:", err)
			return
		}
		_ = setupMq(t, config)
		t.Log("rabbitmq container restarted")
	}()
	t.Log("sending messages after container restart")
	err = router.SendRawAll([]byte("hello2"), "unique")
	if err != nil {
		t.Fatal("cannot send message:", err)
	}
	t.Log("message sent to queue after container restart")

	conn, err := rabbitmqSupport.RawAmqp(t, config, true)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	data := conn.Consume(conn.GetQueue(t, "test1"))
	rabbitmqSupport.CheckReceiveBytes(t, data, []byte("hello2"))
}

func TestConsumerReconnects(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	containerName := fmt.Sprintf("reconnect-test-%d", time.Now().UTC().UnixNano())
	ctx := context.Background()
	port := fmt.Sprintf("%d:%s", 9900, rabbitmqSupport.MqPort)
	container := rabbitmqSupport.CreateMqContainer(ctx, t, containerName, port)
	err := container.Start(ctx)
	if err != nil {
		t.Fatal("cannot start container:", err)
	}
	t.Cleanup(func() {
		err := container.Terminate(ctx)
		if err != nil {
			t.Logf("cannot terminate container: %v", err)
		}
	})
	config := rabbitmqSupport.GetConfigForContainer(ctx, t, container, "test")

	routingKey := setupMq(t, config)

	router, _, manager, err := rabbitmq.NewRouters(config, &queue.RouterConfig{
		Queues: map[string]queue.DestinationConfig{
			"sub-pin1": {
				Exchange:   config.ExchangeName,
				QueueName:  "test1",
				Attributes: []string{"subscribe", "test", "unique"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func(manager io.Closer) {
		err := manager.Close()
		if err != nil {
			t.Logf("cannot close manager connection: %v", err)
		}
	}(manager)

	channel := make(chan []byte, 1)
	monitor, err := router.SubscribeRawAll(rabbitmqSupport.TestRawListener{
		Channel: channel,
	})
	if err != nil {
		t.Fatal("cannot subscribe message:", err)
	}
	defer func(monitor queue.Monitor) {
		_ = monitor.Unsubscribe()
	}(monitor)
	rawConn, err := rabbitmqSupport.RawAmqp(t, config, false)
	if err != nil {
		t.Fatal("cannot get raw connection", err)
	}
	firstBytes := []byte("hello")
	rawConn.Publish(config, routingKey, firstBytes)
	actual := <-channel
	assert.Equal(t, firstBytes, actual)

	err = container.Terminate(ctx)
	if err != nil {
		t.Fatal("cannot stop container:", err)
	}

	// create new container with same port
	// cannot Stop and Start container because of wait strategy
	recovered := make(chan struct{})
	go func() {
		container = rabbitmqSupport.CreateMqContainer(ctx, t, containerName, port)
		err = container.Start(ctx)
		if err != nil {
			t.Error("cannot start container:", err)
			return
		}
		// delay is added to check that consumer will be recovered even if queue does not yet exist
		time.Sleep(5 * time.Second)
		_ = setupMq(t, config)
		t.Log("rabbitmq container restarted")
		recovered <- struct{}{}
		close(recovered)
	}()
	<-recovered
	rawConn, err = rabbitmqSupport.RawAmqp(t, config, false)
	if err != nil {
		t.Fatal("cannot get raw connection", err)
	}
	secondBytes := []byte("hello2")
	rawConn.Publish(config, routingKey, secondBytes)
	actual = <-channel
	assert.Equal(t, secondBytes, actual)
}

func setupMq(t *testing.T, config connection.Config) string {
	conn, err := rabbitmqSupport.RawAmqp(t, config, true)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	queue1 := conn.CreateQueue("test1")
	routingKey1 := "test-publish1"
	conn.BindQueue(config, queue1, routingKey1)
	return routingKey1
}
