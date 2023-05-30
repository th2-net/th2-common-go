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
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/th2-net/th2-common-go/pkg/queue/common"
	"github.com/th2-net/th2-common-go/pkg/queue/message"
)

type CommonMessageRouter struct {
	connManager    *connection.Manager
	subscribers    map[string]Subscriber
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
		subscribers:    make(map[string]Subscriber),
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
	if len(pinFoundByAttrs) == 0 {
		return nil, fmt.Errorf("no pin found for attributes %v", attributes)
	}
	subscribers, err := cmr.subscribeAll(pinFoundByAttrs, func(router *CommonMessageRouter, pinName string) (SubscriberMonitor, error) {
		return router.subByPinWithAck(listener, pinName)
	})

	if len(subscribers) == 0 {
		return nil, errors.New("no such subscriber")
	}

	err = cmr.startAll(subscribers)
	if err != nil {
		return nil, err
	}

	return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
}

func (cmr *CommonMessageRouter) SubscribeAll(listener message.Listener, attributes ...string) (queue.Monitor, error) {
	pinFoundByAttrs := common.FindSubscribeQueuesByAttr(cmr.config, attributes)
	if len(pinFoundByAttrs) == 0 {
		return nil, fmt.Errorf("no pin found for attributes %v", attributes)
	}
	subscribers, err := cmr.subscribeAll(pinFoundByAttrs, func(router *CommonMessageRouter, pinName string) (SubscriberMonitor, error) {
		return router.subByPin(listener, pinName)
	})
	if len(subscribers) == 0 {
		return nil, errors.New("no such subscriber")
	}
	err = cmr.startAll(subscribers)
	if err != nil {
		return nil, err
	}
	return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
}

func (cmr *CommonMessageRouter) SubscribeRawAll(listener message.RawListener, attributes ...string) (queue.Monitor, error) {
	pinFoundByAttrs := common.FindSubscribeQueuesByAttr(cmr.config, attributes)
	if len(pinFoundByAttrs) == 0 {
		return nil, fmt.Errorf("no pin found for attributes %v", attributes)
	}
	subscribers, err := cmr.subscribeAll(pinFoundByAttrs, func(router *CommonMessageRouter, pinName string) (SubscriberMonitor, error) {
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
	return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
}

func (cmr *CommonMessageRouter) subscribeAll(
	pinFoundByAttrs map[string]queue.DestinationConfig,
	subscribeFunc func(router *CommonMessageRouter, pinName string) (SubscriberMonitor, error),
) ([]SubscriberMonitor, error) {
	subscribers := make([]SubscriberMonitor, 0, len(pinFoundByAttrs))
	for queuePin, _ := range pinFoundByAttrs {
		cmr.Logger.Debug().Str("Pin", queuePin).Msg("Subscribing")
		subscriber, err := subscribeFunc(cmr, queuePin)
		if err != nil {
			cmr.Logger.Error().Err(err).Str("Pin", queuePin).Msg("cannot subscribe")
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}
	return subscribers, nil
}

func (cmr *CommonMessageRouter) startAll(subscribers []SubscriberMonitor) error {
	for _, s := range subscribers {
		cmr.Logger.Trace().Str("Pin", s.subscriber.Pin()).Msg("Start subscribing of queue")
		if s.subscriber.IsStarted() {
			cmr.Logger.Trace().Str("Pin", s.subscriber.Pin()).Msg("subscriber already started")
			continue
		}
		err := s.subscriber.Start()
		if err != nil {
			if err == DoubleStartError {
				cmr.Logger.Info().Str("Pin", s.subscriber.Pin()).Msg("already started")
				continue
			}
			cmr.Logger.Error().Err(err).Str("Pin", s.subscriber.Pin()).
				Msg("cannot start subscriber")
			return err
		}
	}
	return nil
}

func (cmr *CommonMessageRouter) subByPin(listener message.Listener, pin string) (SubscriberMonitor, error) {
	subscriber, err := cmr.getSubscriber(pin, autoSubscriber, parsedContentType)
	if err != nil {
		return SubscriberMonitor{}, err
	}
	autoSubscriber, err := asAutoSubscriber(subscriber, pin)
	if err != nil {
		return SubscriberMonitor{}, err
	}
	handler, ok := autoSubscriber.getHandler().(*messageHandler)
	if !ok {
		return SubscriberMonitor{}, fmt.Errorf("handler with different type %T is subscribed to pin %s",
			autoSubscriber.getHandler(), pin)
	}
	handler.SetListener(listener)
	cmr.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber monitor")
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) subByPinRaw(listener message.RawListener, pin string) (SubscriberMonitor, error) {
	subscriber, err := cmr.getSubscriber(pin, autoSubscriber, rawContentType)
	if err != nil {
		return SubscriberMonitor{}, err
	}
	autoSubscriber, err := asAutoSubscriber(subscriber, pin)
	if err != nil {
		return SubscriberMonitor{}, err
	}
	handler, ok := autoSubscriber.getHandler().(*rawMessageHandler)
	if !ok {
		return SubscriberMonitor{}, fmt.Errorf("handler with different type %T is subscribed to pin %s",
			autoSubscriber.getHandler(), pin)
	}
	handler.SetListener(listener)
	cmr.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber monitor")
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) subByPinWithAck(listener message.ConformationListener, pin string) (SubscriberMonitor, error) {
	subscriber, err := cmr.getSubscriber(pin, manualSubscriber, parsedContentType)
	if err != nil {
		return SubscriberMonitor{}, err
	}
	manualSubscriber, err := asManualSubscriber(subscriber, pin)
	if err != nil {
		return SubscriberMonitor{}, err
	}
	handler, ok := manualSubscriber.getHandler().(*confirmationMessageHandler)
	if !ok {
		return SubscriberMonitor{}, fmt.Errorf("handler with different type %T is subscribed to pin %s",
			manualSubscriber.getHandler(), pin)
	}
	handler.SetListener(listener)
	cmr.Logger.Trace().Str("Pin", pin).Msg("Getting subscriber monitor")
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func asAutoSubscriber(subscriber Subscriber, pin string) (AutoSubscriber, error) {
	autoSubscriber, ok := subscriber.(AutoSubscriber)
	if !ok {
		return nil, fmt.Errorf("subscriber with different type %T is subscribed to pin %s",
			subscriber, pin)
	}
	return autoSubscriber, nil
}

func asManualSubscriber(subscriber Subscriber, pin string) (ManualSubscriber, error) {
	manualSubscriber, ok := subscriber.(ManualSubscriber)
	if !ok {
		return nil, fmt.Errorf("subscriber with different type %T is subscribed to pin %s",
			subscriber, pin)
	}
	return manualSubscriber, nil
}

func (cmr *CommonMessageRouter) getSubscriber(pin string, subscriberType subscriberType, contentType contentType) (Subscriber, error) {
	// TODO: probably, we should use lock here to make subscriber creation atomic
	queueConfig := cmr.config.Queues[pin] // get queue by pin
	var result Subscriber
	if _, ok := cmr.subscribers[pin]; ok {
		result = cmr.subscribers[pin]
		return result, nil
	}
	result, err := newSubscriber(cmr.connManager, &queueConfig, pin, subscriberType, contentType)
	if err != nil {
		return nil, err
	}

	cmr.subscribers[pin] = result
	cmr.Logger.Trace().Str("Pin", pin).Msg("Created subscriber")
	return result, nil
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
