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

package message

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/th2-net/th2-common-go/pkg/common"
	grpcCommon "github.com/th2-net/th2-common-go/pkg/common/grpc/th2_grpc_common"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq"
	rabbitmqSupport "github.com/th2-net/th2-common-go/test/modules/rabbitmq"
	"google.golang.org/protobuf/proto"
	"os"
	"testing"
)

var logger = zerolog.New(os.Stdout).With().Str("test_type", "message_router").Logger()

func TestMessageRouterSendAll(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	config := rabbitmqSupport.StartMq(t, "test")

	conn, err := rabbitmqSupport.RawAmqp(t, config, true)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	queue1 := conn.CreateQueue("test1")
	queue2 := conn.CreateQueue("test2")
	routingKey1 := "test-publish1"
	routingKey2 := "test-publish2"
	conn.BindQueue(config, queue1, routingKey1)
	conn.BindQueue(config, queue2, routingKey2)

	router, _, manager, err := rabbitmq.NewRouters(common.BoxConfig{}, config, &queue.RouterConfig{
		Queues: map[string]queue.DestinationConfig{
			"publish-pin1": {
				Exchange:   config.ExchangeName,
				RoutingKey: routingKey1,
				Attributes: []string{"publish", "test", "unique"},
			},
			"publish-pin2": {
				Exchange:   config.ExchangeName,
				RoutingKey: routingKey2,
				Attributes: []string{"publish", "test"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	originalBatch := createBatch()
	deliveries1 := conn.Consume(queue1)
	deliveries2 := conn.Consume(queue2)
	t.Run("one pin", func(t *testing.T) {
		err = router.SendAll(originalBatch, "unique")
		if err != nil {
			t.Fatal("cannot send batch", err)
		}

		rabbitmqSupport.CheckReceiveDelivery(t, deliveries1, originalBatch)
	})
	t.Run("all pins", func(t *testing.T) {
		err = router.SendAll(originalBatch, "test")
		if err != nil {
			t.Fatal("cannot send batch", err)
		}

		rabbitmqSupport.CheckReceiveDelivery(t, deliveries1, originalBatch)
		rabbitmqSupport.CheckReceiveDelivery(t, deliveries2, originalBatch)
	})
}

func TestMessageRouterSendAllReportErrorInNoPinMatch(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	config := rabbitmqSupport.StartMq(t, "test")

	router, _, manager, err := rabbitmq.NewRouters(common.BoxConfig{}, config, &queue.RouterConfig{
		Queues: map[string]queue.DestinationConfig{
			"publish-pin1": {
				Exchange:   config.ExchangeName,
				RoutingKey: "test",
				Attributes: []string{"publish", "test"},
			},
			"publish-pin2": {
				Exchange:   config.ExchangeName,
				QueueName:  "test",
				Attributes: []string{"subscribe", "test2"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	originalBatch := createBatch()

	testData := []struct {
		attribute   string
		description string
	}{
		{
			attribute:   "unknown",
			description: "not additional attribute",
		},
		{
			attribute:   "test2",
			description: "no publish attribute",
		},
	}
	for _, tt := range testData {
		t.Run(tt.description, func(t *testing.T) {
			err = router.SendAll(originalBatch, tt.attribute)
			assert.ErrorContains(t, err, "no pin found for specified attributes")
		})
	}
}

func TestMessageRouterSendRaw(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	config := rabbitmqSupport.StartMq(t, "test")

	conn, err := rabbitmqSupport.RawAmqp(t, config, true)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	queue1 := conn.CreateQueue("test1")
	queue2 := conn.CreateQueue("test2")
	routingKey1 := "test-publish1"
	routingKey2 := "test-publish2"
	conn.BindQueue(config, queue1, routingKey1)
	conn.BindQueue(config, queue2, routingKey2)

	router, _, manager, err := rabbitmq.NewRouters(common.BoxConfig{}, config, &queue.RouterConfig{
		Queues: map[string]queue.DestinationConfig{
			"publish-pin1": {
				Exchange:   config.ExchangeName,
				RoutingKey: routingKey1,
				Attributes: []string{"publish", "test", "unique"},
			},
			"publish-pin2": {
				Exchange:   config.ExchangeName,
				RoutingKey: routingKey2,
				Attributes: []string{"publish", "test"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	originalData := []byte("hello")
	deliveries1 := conn.Consume(queue1)
	deliveries2 := conn.Consume(queue2)
	t.Run("one pin", func(t *testing.T) {
		err = router.SendRawAll(originalData, "unique")
		if err != nil {
			t.Fatal("cannot send batch", err)
		}

		rabbitmqSupport.CheckReceiveBytes(t, deliveries1, originalData)
	})
	t.Run("all pins", func(t *testing.T) {
		err = router.SendRawAll(originalData, "test")
		if err != nil {
			t.Fatal("cannot send batch", err)
		}

		rabbitmqSupport.CheckReceiveBytes(t, deliveries1, originalData)
		rabbitmqSupport.CheckReceiveBytes(t, deliveries2, originalData)
	})
}

func TestMessageRouterSendRawReportErrorInNoPinMatch(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	config := rabbitmqSupport.StartMq(t, "test")

	router, _, manager, err := rabbitmq.NewRouters(common.BoxConfig{}, config, &queue.RouterConfig{
		Queues: map[string]queue.DestinationConfig{
			"publish-pin1": {
				Exchange:   config.ExchangeName,
				RoutingKey: "test",
				Attributes: []string{"publish", "test"},
			},
			"publish-pin2": {
				Exchange:   config.ExchangeName,
				QueueName:  "test",
				Attributes: []string{"subscribe", "test2"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	testData := []struct {
		attribute   string
		description string
	}{
		{
			attribute:   "unknown",
			description: "not additional attribute",
		},
		{
			attribute:   "test2",
			description: "no publish attribute",
		},
	}
	for _, tt := range testData {
		t.Run(tt.description, func(t *testing.T) {
			err = router.SendRawAll([]byte("hello"), tt.attribute)
			assert.ErrorContains(t, err, "no pin found for specified attributes")
		})
	}
}

func TestMessageRouterSubscribeAll(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	config := rabbitmqSupport.StartMq(t, "test")

	conn, err := rabbitmqSupport.RawAmqp(t, config, true)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	queue1 := conn.CreateQueue("test1")
	queue2 := conn.CreateQueue("test2")
	queue3 := conn.CreateQueue("test3")
	key1 := "test-publish1"
	key2 := "test-publish2"
	key3 := "test-publish3"

	conn.BindQueue(config, queue1, key1)
	conn.BindQueue(config, queue2, key2)
	conn.BindQueue(config, queue3, key3)

	router, _, manager, err := rabbitmq.NewRouters(common.BoxConfig{}, config, &queue.RouterConfig{
		Queues: map[string]queue.DestinationConfig{
			"sub-pin1": {
				Exchange:   config.ExchangeName,
				QueueName:  queue1.Name,
				Attributes: []string{"subscribe", "common"},
			},
			"sub-pin2": {
				Exchange:   config.ExchangeName,
				QueueName:  queue2.Name,
				Attributes: []string{"subscribe", "common"},
			},
			"sub-pin3": {
				Exchange:   config.ExchangeName,
				QueueName:  queue3.Name,
				Attributes: []string{"subscribe", "unique"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	originalBatch := createBatch()
	data, err := proto.Marshal(originalBatch)

	deliveries1 := make(chan *grpcCommon.MessageGroupBatch, 1)
	mon1, err := router.SubscribeAll(&rabbitmqSupport.GenericListener[grpcCommon.MessageGroupBatch]{
		Channel: deliveries1,
	}, "common")
	if err != nil {
		t.Fatal(err)
	}
	defer mon1.Unsubscribe()
	deliveries2 := make(chan *grpcCommon.MessageGroupBatch, 1)
	mon2, err := router.SubscribeAll(&rabbitmqSupport.GenericListener[grpcCommon.MessageGroupBatch]{
		Channel: deliveries2,
	}, "unique")
	if err != nil {
		t.Fatal(err)
	}
	defer mon2.Unsubscribe()

	t.Run("single_queue", func(t *testing.T) {
		conn.Publish(config, key3, data)
		rabbitmqSupport.CheckReceiveBatch(t, deliveries2, originalBatch)
	})
	t.Run("multiple_queue", func(t *testing.T) {
		conn.Publish(config, key1, data)
		rabbitmqSupport.CheckReceiveBatch(t, deliveries1, originalBatch)

		conn.Publish(config, key2, data)
		rabbitmqSupport.CheckReceiveBatch(t, deliveries1, originalBatch)
	})
}

func TestMessageRouterSubscribeAllWithAck(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	config := rabbitmqSupport.StartMq(t, "test")

	conn, err := rabbitmqSupport.RawAmqp(t, config, true)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	queue1 := conn.CreateQueue("test1")
	queue2 := conn.CreateQueue("test2")
	queue3 := conn.CreateQueue("test3")
	key1 := "test-publish1"
	key2 := "test-publish2"
	key3 := "test-publish3"

	conn.BindQueue(config, queue1, key1)
	conn.BindQueue(config, queue2, key2)
	conn.BindQueue(config, queue3, key3)

	router, _, manager, err := rabbitmq.NewRouters(common.BoxConfig{}, config, &queue.RouterConfig{
		Queues: map[string]queue.DestinationConfig{
			"sub-pin1": {
				Exchange:   config.ExchangeName,
				QueueName:  queue1.Name,
				Attributes: []string{"subscribe", "common"},
			},
			"sub-pin2": {
				Exchange:   config.ExchangeName,
				QueueName:  queue2.Name,
				Attributes: []string{"subscribe", "common"},
			},
			"sub-pin3": {
				Exchange:   config.ExchangeName,
				QueueName:  queue3.Name,
				Attributes: []string{"subscribe", "unique"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	originalBatch := createBatch()
	data, err := proto.Marshal(originalBatch)

	deliveries1 := make(chan *grpcCommon.MessageGroupBatch, 1)
	mon1, err := router.SubscribeAllWithManualAck(&rabbitmqSupport.GenericManualListener[grpcCommon.MessageGroupBatch]{
		Channel:        deliveries1,
		OnConfirmation: rabbitmqSupport.Confirm,
	}, "common")
	if err != nil {
		t.Fatal(err)
	}
	defer mon1.Unsubscribe()
	deliveries2 := make(chan *grpcCommon.MessageGroupBatch, 1)
	mon2, err := router.SubscribeAllWithManualAck(&rabbitmqSupport.GenericManualListener[grpcCommon.MessageGroupBatch]{
		Channel:        deliveries2,
		OnConfirmation: rabbitmqSupport.Confirm,
	}, "unique")
	if err != nil {
		t.Fatal(err)
	}
	defer mon2.Unsubscribe()

	t.Run("single_queue", func(t *testing.T) {
		conn.Publish(config, key3, data)
		rabbitmqSupport.CheckReceiveBatch(t, deliveries2, originalBatch)
	})
	t.Run("multiple_queue", func(t *testing.T) {
		conn.Publish(config, key1, data)
		rabbitmqSupport.CheckReceiveBatch(t, deliveries1, originalBatch)

		conn.Publish(config, key2, data)
		rabbitmqSupport.CheckReceiveBatch(t, deliveries1, originalBatch)
	})
}

func TestMessageRouterSubscribeAllReportErrorInNoPinMatch(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	config := rabbitmqSupport.StartMq(t, "test")

	router, _, manager, err := rabbitmq.NewRouters(common.BoxConfig{}, config, &queue.RouterConfig{
		Queues: map[string]queue.DestinationConfig{
			"sub-pin": {
				Exchange:   config.ExchangeName,
				QueueName:  "test_queue",
				Attributes: []string{"subscribe", "test"},
			},
			"pub-pin": {
				Exchange:   config.ExchangeName,
				RoutingKey: "test_queue",
				Attributes: []string{"publish", "test2"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	testData := []struct {
		attribute   string
		description string
	}{
		{
			attribute:   "unknown",
			description: "not additional attribute",
		},
		{
			attribute:   "test2",
			description: "no subscribe attribute",
		},
	}
	for _, tt := range testData {
		t.Run(tt.description, func(t *testing.T) {
			_, err := router.SubscribeAll(&rabbitmqSupport.GenericListener[grpcCommon.MessageGroupBatch]{}, tt.attribute)
			assert.ErrorContains(t, err, fmt.Sprintf("no pin found for attributes [%s]", tt.attribute))
		})
	}
}

func TestMessageRouterSubscribeAllWithManualAckReportErrorInNoPinMatch(t *testing.T) {
	if testing.Short() {
		t.Skip("do not run containers in short run")
		return
	}
	config := rabbitmqSupport.StartMq(t, "test")

	router, _, manager, err := rabbitmq.NewRouters(common.BoxConfig{}, config, &queue.RouterConfig{
		Queues: map[string]queue.DestinationConfig{
			"sub-pin": {
				Exchange:   config.ExchangeName,
				QueueName:  "test_queue",
				Attributes: []string{"subscribe", "test"},
			},
			"pub-pin": {
				Exchange:   config.ExchangeName,
				RoutingKey: "test_queue",
				Attributes: []string{"publish", "test2"},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.Close()

	testData := []struct {
		attribute   string
		description string
	}{
		{
			attribute:   "unknown",
			description: "not additional attribute",
		},
		{
			attribute:   "test2",
			description: "no subscribe attribute",
		},
	}
	for _, tt := range testData {
		t.Run(tt.description, func(t *testing.T) {
			_, err := router.SubscribeAllWithManualAck(&rabbitmqSupport.GenericManualListener[grpcCommon.MessageGroupBatch]{}, tt.attribute)
			assert.ErrorContains(t, err, fmt.Sprintf("no pin found for attributes [%s]", tt.attribute))
		})
	}
}
