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
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/connection"
	"os"
	"sync"
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/pkg/queue/common"
	"github.com/th2-net/th2-common-go/pkg/queue/event"
)

type CommonEventRouter struct {
	connManager *connection.Manager
	subscribers map[string]*CommonEventSubscriber
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
		subscribers: make(map[string]*CommonEventSubscriber),
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
	var subscribers []SubscriberMonitor
	for queuePin, _ := range pinsFoundByAttrs {
		cer.Logger.Debug().Str("Pin", queuePin).Msg("Subscribing")
		subscriber, err := cer.subByPin(listener, queuePin)
		if err != nil {
			cer.Logger.Error().Err(err).Send()
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}
	if len(subscribers) == 0 {
		cer.Logger.Error().Msg("No such subscriber")
		return nil, errors.New("no such subscriber")
	}
	// TODO: mutex here does not make any sense
	var m sync.Mutex
	for _, s := range subscribers {
		m.Lock()
		err := s.subscriber.Start()
		if err != nil {
			return SubscriberMonitor{}, err
		}
		m.Unlock()
	}
	return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
}

func (cer *CommonEventRouter) SubscribeAllWithManualAck(listener event.ConformationListener, attributes ...string) (queue.Monitor, error) {
	pinFoundByAttrs := common.FindSubscribeEventQueuesByAttr(cer.config, attributes)
	var subscribers []SubscriberMonitor
	for queuePin, _ := range pinFoundByAttrs {
		cer.Logger.Debug().Str("Pin", queuePin).Msg("Subscribing with manual ack")
		subscriber, err := cer.subByPinWithAck(listener, queuePin)
		if err != nil {
			cer.Logger.Error().Err(err).Send()
			return SubscriberMonitor{}, err
		}
		subscribers = append(subscribers, subscriber)
	}
	if len(subscribers) == 0 {
		cer.Logger.Error().Msg("No such subscriber")
		return SubscriberMonitor{}, errors.New("no such subscriber")
	}
	// TODO: mutex here does not make any sense
	var m sync.Mutex
	for _, s := range subscribers {
		m.Lock()
		err := s.subscriber.ConfirmationStart()
		if err != nil {
			return SubscriberMonitor{}, err
		}
		m.Unlock()
	}

	return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
}

func (cer *CommonEventRouter) subByPin(listener event.Listener, pin string) (SubscriberMonitor, error) {
	subscriber := cer.getSubscriber(pin)
	subscriber.AddListener(listener)
	cer.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber monitor")
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cer *CommonEventRouter) subByPinWithAck(listener event.ConformationListener, pin string) (SubscriberMonitor, error) {
	subscriber := cer.getSubscriber(pin)
	subscriber.AddConfirmationListener(listener)
	cer.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber(with ack) monitor")
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cer *CommonEventRouter) getSubscriber(pin string) *CommonEventSubscriber {
	queueConfig := cer.config.Queues[pin] // get queue by pin
	var result *CommonEventSubscriber
	if _, ok := cer.subscribers[pin]; ok {
		result = cer.subscribers[pin]
		return result
	}
	result = &CommonEventSubscriber{connManager: cer.connManager, qConfig: &queueConfig,
		listener: nil, confirmationListener: nil, th2Pin: pin, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
	cer.subscribers[pin] = result
	cer.Logger.Trace().Str("Pin", pin).Msg("Created subscriber")
	return result
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
