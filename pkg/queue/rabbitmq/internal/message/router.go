/*
 * Copyright 2022-2023 Exactpro (Exactpro Systems Limited)
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
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	p_buff "github.com/th2-net/th2-common-go/pkg/common/grpc/th2_grpc_common"
	"github.com/th2-net/th2-common-go/pkg/log"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/common"
	"github.com/th2-net/th2-common-go/pkg/queue/filter"
	"github.com/th2-net/th2-common-go/pkg/queue/message"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/connection"
	"sync"
)

type CommonMessageRouter struct {
	connManager    *connection.Manager
	subscribers    map[string]internal.Subscriber
	senders        map[string]*CommonMessageSender
	filterStrategy filter.Strategy
	config         *queue.RouterConfig
	Logger         zerolog.Logger
	mutex          *sync.RWMutex
}

func NewRouter(
	manager *connection.Manager,
	config *queue.RouterConfig,
	logger zerolog.Logger,
) *CommonMessageRouter {
	return &CommonMessageRouter{
		connManager:    manager,
		subscribers:    make(map[string]internal.Subscriber),
		senders:        make(map[string]*CommonMessageSender),
		filterStrategy: filter.Default,
		Logger:         logger,
		config:         config,
		mutex:          &sync.RWMutex{},
	}
}

func (cmr *CommonMessageRouter) Close() error {
	return nil
}

func (cmr *CommonMessageRouter) SendAll(msgBatch *p_buff.MessageGroupBatch, attributes ...string) error {
	pinsFoundByAttrs := common.FindSendQueuesByAttr(cmr.config, attributes)
	if len(pinsFoundByAttrs) == 0 {
		cmr.Logger.Error().
			Strs("attributes", attributes).
			Msg("No such queue to send message")
		return fmt.Errorf("no pin found for specified attributes: %v", attributes)
	}
	for pin, config := range pinsFoundByAttrs {
		if !cmr.filterStrategy.Verify(msgBatch, config.Filters) {
			if e := cmr.Logger.Debug(); e.Enabled() {
				e.Str("Pin", pin).
					Interface("Metadata", filter.FirstIDFromMsgBatch(msgBatch)).
					Msg("First ID of message batch didn't match filter")
			}
			continue
		}
		if e := cmr.Logger.Debug(); e.Enabled() {
			e.Str("Pin", pin).
				Interface("Metadata", filter.FirstIDFromMsgBatch(msgBatch)).
				Msg("First ID of message batch matched filter")
		}
		sender := cmr.getSender(pin)
		err := sender.Send(msgBatch)
		if err != nil {
			cmr.Logger.Error().Err(err).Send()
			return err
		}
		if e := cmr.Logger.Debug(); e.Enabled() {
			e.Str("sending to pin", pin).
				Interface("Metadata", filter.FirstIDFromMsgBatch(msgBatch)).
				Msg("First ID of sent Message batch")
		}
	}
	return nil
}

func (cmr *CommonMessageRouter) SendRawAll(rawData []byte, attributes ...string) error {
	pinsFoundByAttrs := common.FindSendQueuesByAttr(cmr.config, attributes)
	if len(pinsFoundByAttrs) == 0 {
		return fmt.Errorf("no pin found for specified attributes: %v", attributes)
	}
	for pin, _ := range pinsFoundByAttrs {
		sender := cmr.getSender(pin)
		err := sender.SendRaw(rawData)
		if err != nil {
			return err
		}
	}
	return nil

}

func (cmr *CommonMessageRouter) SubscribeAllWithManualAck(listener message.ConformationListener, attributes ...string) (queue.Monitor, error) {
	pinFoundByAttrs := common.FindSubscribeQueuesByAttr(cmr.config, attributes)
	if len(pinFoundByAttrs) == 0 {
		return nil, fmt.Errorf("no pin found for attributes %v", attributes)
	}
	subscribers, err := cmr.subscribeAll(pinFoundByAttrs, func(router *CommonMessageRouter, pinName string) (internal.SubscriberMonitor, error) {
		return router.subByPinWithAck(listener, pinName)
	})
	if err != nil {
		return nil, err
	}

	if len(subscribers) == 0 {
		return nil, errors.New("no such subscriber")
	}

	err = cmr.startAll(subscribers)
	if err != nil {
		return nil, err
	}

	return internal.MultiplySubscribeMonitor{SubscriberMonitors: subscribers}, nil
}

func (cmr *CommonMessageRouter) SubscribeAll(listener message.Listener, attributes ...string) (queue.Monitor, error) {
	pinFoundByAttrs := common.FindSubscribeQueuesByAttr(cmr.config, attributes)
	if len(pinFoundByAttrs) == 0 {
		return nil, fmt.Errorf("no pin found for attributes %v", attributes)
	}
	subscribers, err := cmr.subscribeAll(pinFoundByAttrs, func(router *CommonMessageRouter, pinName string) (internal.SubscriberMonitor, error) {
		return router.subByPin(listener, pinName)
	})
	if err != nil {
		return nil, err
	}
	if len(subscribers) == 0 {
		return nil, errors.New("no such subscriber")
	}
	err = cmr.startAll(subscribers)
	if err != nil {
		return nil, err
	}
	return internal.MultiplySubscribeMonitor{SubscriberMonitors: subscribers}, nil
}

func (cmr *CommonMessageRouter) SubscribeRawAll(listener message.RawListener, attributes ...string) (queue.Monitor, error) {
	pinFoundByAttrs := common.FindSubscribeQueuesByAttr(cmr.config, attributes)
	if len(pinFoundByAttrs) == 0 {
		return nil, fmt.Errorf("no pin found for attributes %v", attributes)
	}
	subscribers, err := cmr.subscribeAll(pinFoundByAttrs, func(router *CommonMessageRouter, pinName string) (internal.SubscriberMonitor, error) {
		return router.subByPinRaw(listener, pinName)
	})
	if err != nil {
		return nil, err
	}
	if len(subscribers) == 0 {
		return nil, errors.New("no such subscriber")
	}
	err = cmr.startAll(subscribers)
	if err != nil {
		return nil, err
	}
	return internal.MultiplySubscribeMonitor{SubscriberMonitors: subscribers}, nil
}

func (cmr *CommonMessageRouter) subscribeAll(
	pinFoundByAttrs map[string]queue.DestinationConfig,
	subscribeFunc func(router *CommonMessageRouter, pinName string) (internal.SubscriberMonitor, error),
) ([]internal.SubscriberMonitor, error) {
	return internal.SubscribeAll(cmr, pinFoundByAttrs, &cmr.Logger, subscribeFunc)
}

func (cmr *CommonMessageRouter) startAll(subscribers []internal.SubscriberMonitor) error {
	return internal.StartAll(subscribers, &cmr.Logger)
}

func (cmr *CommonMessageRouter) subByPin(listener message.Listener, pin string) (internal.SubscriberMonitor, error) {
	subscriber, err := cmr.getSubscriber(pin, internal.AutoSubscriberType, parsedContentType)
	if err != nil {
		return nil, err
	}
	autoSubscriber, err := internal.AsAutoSubscriber(subscriber, pin)
	if err != nil {
		return nil, err
	}
	handler, ok := autoSubscriber.GetHandler().(*messageHandler)
	if !ok {
		return nil, fmt.Errorf("handler with different type %T is subscribed to pin %s",
			autoSubscriber.GetHandler(), pin)
	}
	handler.SetListener(listener)
	cmr.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber monitor")
	return internal.MonitorFor(subscriber), nil
}

func (cmr *CommonMessageRouter) subByPinRaw(listener message.RawListener, pin string) (internal.SubscriberMonitor, error) {
	subscriber, err := cmr.getSubscriber(pin, internal.AutoSubscriberType, rawContentType)
	if err != nil {
		return nil, err
	}
	autoSubscriber, err := internal.AsAutoSubscriber(subscriber, pin)
	if err != nil {
		return nil, err
	}
	handler, ok := autoSubscriber.GetHandler().(*rawMessageHandler)
	if !ok {
		return nil, fmt.Errorf("handler with different type %T is subscribed to pin %s",
			autoSubscriber.GetHandler(), pin)
	}
	handler.SetListener(listener)
	cmr.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber monitor")
	return internal.MonitorFor(subscriber), nil
}

func (cmr *CommonMessageRouter) subByPinWithAck(listener message.ConformationListener, pin string) (internal.SubscriberMonitor, error) {
	subscriber, err := cmr.getSubscriber(pin, internal.ManualSubscriberType, parsedContentType)
	if err != nil {
		return nil, err
	}
	manualSubscriber, err := internal.AsManualSubscriber(subscriber, pin)
	if err != nil {
		return nil, err
	}
	handler, ok := manualSubscriber.GetHandler().(*confirmationMessageHandler)
	if !ok {
		return nil, fmt.Errorf("handler with different type %T is subscribed to pin %s",
			manualSubscriber.GetHandler(), pin)
	}
	handler.SetListener(listener)
	cmr.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber monitor")
	return internal.MonitorFor(subscriber), nil
}

func (cmr *CommonMessageRouter) getSubscriber(pin string, subscriberType internal.SubscriberType, contentType contentType) (internal.Subscriber, error) {
	// TODO: probably, we should use lock here to make subscriber creation atomic
	queueConfig := cmr.config.Queues[pin] // get queue by pin
	var result internal.Subscriber
	result = cmr.findSubscriber(pin)
	if result != nil {
		return result, nil
	}
	cmr.mutex.Lock()
	defer cmr.mutex.Unlock()

	// check if someone already created the subscriber in a different goroutine
	if existing, ok := cmr.subscribers[pin]; ok {
		return existing, nil
	}

	result, err := newSubscriber(cmr.connManager, &queueConfig, pin, subscriberType, contentType)
	if err != nil {
		return nil, err
	}

	cmr.subscribers[pin] = result
	cmr.Logger.Trace().Str("Pin", pin).Msg("Created subscriber")
	return result, nil
}

func (cmr *CommonMessageRouter) findSubscriber(pin string) internal.Subscriber {
	cmr.mutex.RLock()
	defer cmr.mutex.RUnlock()

	if existing, ok := cmr.subscribers[pin]; ok {
		return existing
	}
	return nil
}

func (cmr *CommonMessageRouter) getSender(pin string) *CommonMessageSender {
	queueConfig := cmr.config.Queues[pin] // get queue by pin
	var result *CommonMessageSender

	result = cmr.findSender(pin)
	if result != nil {
		return result
	}

	cmr.mutex.Lock()
	defer cmr.mutex.Unlock()

	// check if someone already created the sender in a different goroutine
	if existing, ok := cmr.senders[pin]; ok {
		return existing
	}

	result = &CommonMessageSender{ConnManager: cmr.connManager, exchangeName: queueConfig.Exchange,
		sendQueue: queueConfig.RoutingKey, th2Pin: pin, Logger: log.ForComponent("rabbitmq_message_sender")}
	cmr.senders[pin] = result
	cmr.Logger.Trace().Str("Pin", pin).Msg("Created sender")
	return result
}

func (cmr *CommonMessageRouter) findSender(pin string) *CommonMessageSender {
	cmr.mutex.RLock()
	defer cmr.mutex.RUnlock()

	if existing, ok := cmr.senders[pin]; ok {
		return existing
	}
	return nil
}
