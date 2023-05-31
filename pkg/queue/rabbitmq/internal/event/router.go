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

package event

import (
	"errors"
	"fmt"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/connection"
	"os"
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/pkg/queue/common"
	"github.com/th2-net/th2-common-go/pkg/queue/event"
)

type CommonEventRouter struct {
	connManager *connection.Manager
	subscribers map[string]internal.Subscriber
	senders     map[string]*CommonEventSender
	config      *queue.RouterConfig
	Logger      zerolog.Logger
}

func NewRouter(
	manager *connection.Manager,
	config *queue.RouterConfig,
	logger zerolog.Logger,
) *CommonEventRouter {
	return &CommonEventRouter{
		connManager: manager,
		subscribers: make(map[string]internal.Subscriber),
		senders:     make(map[string]*CommonEventSender),
		config:      config,
		Logger:      logger,
	}
}

func (cer *CommonEventRouter) Close() error {
	return nil
}

func (cer *CommonEventRouter) SendAll(EventBatch *p_buff.EventBatch, attributes ...string) error {
	pinsFoundByAttrs := common.FindSendEventQueuesByAttr(cer.config, attributes)
	if len(pinsFoundByAttrs) == 0 {
		cer.Logger.Error().
			Any("attributes", attributes).
			Msg("No such queue to send message")
		return fmt.Errorf("no pin found for specified attributes: %v", attributes)
	}
	for pin, _ := range pinsFoundByAttrs {
		sender := cer.getSender(pin)
		err := sender.Send(EventBatch)
		if err != nil {
			cer.Logger.Error().Err(err).Send()
			return err
		}
	}
	return nil

}

func (cer *CommonEventRouter) SubscribeAll(listener event.Listener, attributes ...string) (queue.Monitor, error) {
	pinsFoundByAttrs := common.FindSubscribeEventQueuesByAttr(cer.config, attributes)
	if len(pinsFoundByAttrs) == 0 {
		return nil, fmt.Errorf("no pin found for attributes %v", attributes)
	}
	subscribers, err := internal.SubscribeAll(cer, pinsFoundByAttrs, &cer.Logger, func(router *CommonEventRouter, pinName string) (internal.SubscriberMonitor, error) {
		return router.subByPin(listener, pinName)
	})
	if err != nil {
		return nil, err
	}
	if len(subscribers) == 0 {
		return nil, errors.New("no such subscriber")
	}
	err = internal.StartAll(subscribers, &cer.Logger)
	if err != nil {
		return nil, err
	}
	return internal.MultiplySubscribeMonitor{SubscriberMonitors: subscribers}, nil
}

func (cer *CommonEventRouter) SubscribeAllWithManualAck(listener event.ConformationListener, attributes ...string) (queue.Monitor, error) {
	pinFoundByAttrs := common.FindSubscribeEventQueuesByAttr(cer.config, attributes)
	if len(pinFoundByAttrs) == 0 {
		return nil, fmt.Errorf("no pin found for attributes %v", attributes)
	}
	subscribers, err := internal.SubscribeAll(cer, pinFoundByAttrs, &cer.Logger, func(router *CommonEventRouter, pinName string) (internal.SubscriberMonitor, error) {
		return router.subByPinWithAck(listener, pinName)
	})
	if err != nil {
		return nil, err
	}
	if len(subscribers) == 0 {
		return nil, errors.New("no such subscriber")
	}
	err = internal.StartAll(subscribers, &cer.Logger)
	if err != nil {
		return nil, err
	}

	return internal.MultiplySubscribeMonitor{SubscriberMonitors: subscribers}, nil
}

func (cer *CommonEventRouter) subByPin(listener event.Listener, pin string) (internal.SubscriberMonitor, error) {
	subscriber, err := cer.getSubscriber(pin, internal.AutoSubscriberType)
	if err != nil {
		return nil, err
	}
	autoSubscriber, err := internal.AsAutoSubscriber(subscriber, pin)
	if err != nil {
		return nil, err
	}
	handler, ok := autoSubscriber.GetHandler().(*autoEventHandler)
	if !ok {
		return nil, fmt.Errorf("handler with different type %T is subscribed to pin %s",
			autoSubscriber.GetHandler(), pin)
	}
	handler.SetListener(listener)
	cer.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber monitor")
	return internal.MonitorFor(subscriber), nil
}

func (cer *CommonEventRouter) subByPinWithAck(listener event.ConformationListener, pin string) (internal.SubscriberMonitor, error) {
	subscriber, err := cer.getSubscriber(pin, internal.ManualSubscriberType)
	if err != nil {
		return nil, err
	}
	autoSubscriber, err := internal.AsManualSubscriber(subscriber, pin)
	if err != nil {
		return nil, err
	}
	handler, ok := autoSubscriber.GetHandler().(*confirmationEventHandler)
	if !ok {
		return nil, fmt.Errorf("handler with different type %T is subscribed to pin %s",
			autoSubscriber.GetHandler(), pin)
	}
	handler.SetListener(listener)
	cer.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber(with ack) monitor")
	return internal.MonitorFor(subscriber), nil
}

func (cer *CommonEventRouter) getSubscriber(pin string, subscriberType internal.SubscriberType) (internal.Subscriber, error) {
	queueConfig := cer.config.Queues[pin] // get queue by pin
	var result internal.Subscriber
	if _, ok := cer.subscribers[pin]; ok {
		result = cer.subscribers[pin]
		return result, nil
	}
	result, err := newSubscriber(cer.connManager, &queueConfig, pin, subscriberType)
	if err != nil {
		return nil, err
	}
	cer.subscribers[pin] = result
	cer.Logger.Trace().Str("Pin", pin).Msg("Created subscriber")
	return result, nil
}

func (cer *CommonEventRouter) getSender(pin string) *CommonEventSender {
	queueConfig := cer.config.Queues[pin] // get queue by pin
	var result *CommonEventSender
	if _, ok := cer.senders[pin]; ok {
		result = cer.senders[pin]
		return result
	}
	result = &CommonEventSender{ConnManager: cer.connManager, exchangeName: queueConfig.Exchange,
		sendQueue: queueConfig.RoutingKey, th2Pin: pin, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
	cer.senders[pin] = result
	cer.Logger.Trace().Str("Pin", pin).Msg("Created sender")
	return result
}
