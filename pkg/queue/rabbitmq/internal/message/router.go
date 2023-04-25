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
	"github.com/rs/zerolog/log"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/filter"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/connection"
	"os"
	"sync"
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/th2-net/th2-common-go/pkg/queue/common"
	"github.com/th2-net/th2-common-go/pkg/queue/message"
)

type CommonMessageRouter struct {
	connManager    *connection.Manager
	subscribers    map[string]*CommonMessageSubscriber
	senders        map[string]*CommonMessageSender
	filterStrategy filter.Strategy
	config         *queue.RouterConfig
	Logger         zerolog.Logger
}

func NewRouter(
	manager *connection.Manager,
	config *queue.RouterConfig,
	logger zerolog.Logger,
) *CommonMessageRouter {
	return &CommonMessageRouter{
		connManager:    manager,
		subscribers:    make(map[string]*CommonMessageSubscriber),
		senders:        make(map[string]*CommonMessageSender),
		filterStrategy: filter.Default,
		Logger:         logger,
		config:         config,
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
			if e := log.Debug(); e.Enabled() {
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
		if e := log.Debug(); e.Enabled() {
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
	var subscribers []SubscriberMonitor
	for queuePin, _ := range pinFoundByAttrs {
		cmr.Logger.Debug().Str("Pin", queuePin).Msg("Subscribing with manual ack")
		subscriber, err := cmr.subByPinWithAck(listener, queuePin)
		if err != nil {
			cmr.Logger.Error().Err(err).Send()
			return SubscriberMonitor{}, err
		}
		subscribers = append(subscribers, subscriber)
	}

	if len(subscribers) == 0 {
		return nil, errors.New("no such subscriber")
	}
	// TODO: mutex here does not make any sense
	var m sync.Mutex
	for _, s := range subscribers {
		m.Lock()
		cmr.Logger.Trace().Str("Pin", s.subscriber.th2Pin).Msg("Start confirmation subscribing of queue")
		err := s.subscriber.ConfirmationStart()
		if err != nil {
			return SubscriberMonitor{}, err
		}
		m.Unlock()
	}

	return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
}

func (cmr *CommonMessageRouter) SubscribeAll(listener message.Listener, attributes ...string) (queue.Monitor, error) {
	pinFoundByAttrs := common.FindSubscribeQueuesByAttr(cmr.config, attributes)
	var subscribers []SubscriberMonitor
	for queuePin, _ := range pinFoundByAttrs {
		cmr.Logger.Debug().Str("Pin", queuePin).Msg("Subscribing")
		subscriber, err := cmr.subByPin(listener, queuePin)
		if err != nil {
			cmr.Logger.Error().Err(err).Send()
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}
	if len(subscribers) == 0 {
		return nil, errors.New("no such subscriber")
	}
	for _, s := range subscribers {
		cmr.Logger.Trace().Str("Pin", s.subscriber.th2Pin).Msg("Start subscribing of queue")
		err := s.subscriber.Start()
		if err != nil {
			cmr.Logger.Error().Err(err).Send()
			return nil, err
		}
	}
	return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
}

func (cmr *CommonMessageRouter) SubscribeRawAll(listener message.RawListener, attributes ...string) (queue.Monitor, error) {
	pinFoundByAttrs := common.FindSubscribeQueuesByAttr(cmr.config, attributes)
	var subscribers []SubscriberMonitor
	for queuePin, _ := range pinFoundByAttrs {
		cmr.Logger.Debug().Str("Pin", queuePin).Msg("Subscribing")
		subscriber, err := cmr.subByPinRaw(listener, queuePin)
		if err != nil {
			cmr.Logger.Error().Err(err).Send()
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}
	if len(subscribers) == 0 {
		return nil, errors.New("no such subscriber")
	}
	for _, s := range subscribers {
		cmr.Logger.Trace().Str("Pin", s.subscriber.th2Pin).Msg("Start subscribing of queue")
		err := s.subscriber.StartRaw()
		if err != nil {
			cmr.Logger.Error().Err(err).Send()
			return nil, err
		}
	}
	return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
}

func (cmr *CommonMessageRouter) subByPin(listener message.Listener, pin string) (SubscriberMonitor, error) {
	subscriber := cmr.getSubscriber(pin)
	subscriber.SetListener(listener)
	cmr.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber monitor")
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) subByPinRaw(listener message.RawListener, pin string) (SubscriberMonitor, error) {
	subscriber := cmr.getSubscriber(pin)
	subscriber.SetRawListener(listener)
	cmr.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber monitor")
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) subByPinWithAck(listener message.ConformationListener, pin string) (SubscriberMonitor, error) {
	subscriber := cmr.getSubscriber(pin)
	subscriber.AddConfirmationListener(listener)
	cmr.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber monitor")
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) getSubscriber(pin string) *CommonMessageSubscriber {
	queueConfig := cmr.config.Queues[pin] // get queue by pin
	var result *CommonMessageSubscriber
	if _, ok := cmr.subscribers[pin]; ok {
		result = cmr.subscribers[pin]
		return result
	}
	result = &CommonMessageSubscriber{connManager: cmr.connManager, qConfig: &queueConfig,
		listener: nil, confirmationListener: nil, th2Pin: pin, Logger: zerolog.New(os.Stdout).With().Str("component", "rabbitmq_message_subscriber").Timestamp().Logger()}
	cmr.subscribers[pin] = result
	cmr.Logger.Trace().Str("Pin", pin).Msg("Created subscriber")
	return result
}

func (cmr *CommonMessageRouter) getSender(pin string) *CommonMessageSender {
	queueConfig := cmr.config.Queues[pin] // get queue by pin
	var result *CommonMessageSender
	if _, ok := cmr.senders[pin]; ok {
		result = cmr.senders[pin]
		return result
	}
	result = &CommonMessageSender{ConnManager: cmr.connManager, exchangeName: queueConfig.Exchange,
		sendQueue: queueConfig.RoutingKey, th2Pin: pin, Logger: zerolog.New(os.Stdout).With().Str("component", "rabbitmq_message_sender").Timestamp().Logger()}
	cmr.senders[pin] = result
	cmr.Logger.Trace().Str("Pin", pin).Msg("Created sender")
	return result
}
