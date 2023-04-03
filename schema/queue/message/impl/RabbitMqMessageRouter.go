/*
 * Copyright 2022 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package message

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/th2-net/th2-common-go/schema/filter"
	"os"
	"sync"
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/message"
)

type CommonMessageRouter struct {
	connManager    *MQcommon.ConnectionManager
	subscribers    map[string]CommonMessageSubscriber
	senders        map[string]CommonMessageSender
	filterStrategy filter.FilterStrategy

	Logger zerolog.Logger
}

func (cmr *CommonMessageRouter) Construct(manager *MQcommon.ConnectionManager) {
	cmr.connManager = manager
	cmr.subscribers = map[string]CommonMessageSubscriber{}
	cmr.senders = map[string]CommonMessageSender{}
	cmr.Logger.Debug().Msg("CommonMessageRouter was initialized")
	cmr.filterStrategy = filter.Default
}

func (cmr *CommonMessageRouter) Close() {
	_ = cmr.connManager.Close()
}

func (cmr *CommonMessageRouter) SendAll(msgBatch *p_buff.MessageGroupBatch, attributes ...string) error {
	attrs := MQcommon.GetSendAttributes(attributes)
	pinsFoundByAttrs := cmr.connManager.QConfig.FindQueuesByAttr(attrs)
	if len(pinsFoundByAttrs) == 0 {
		cmr.Logger.Fatal().Msg("No such queue to send message")
		return nil
	}
	for pin, config := range pinsFoundByAttrs {
		cmr.Logger.Debug().Str("Pin", pin).Msg("Try filtering")
		if cmr.filterStrategy.Verify(msgBatch, config.Filters) {
			if e := cmr.Logger.Debug(); e.Enabled() {
				e.Fields(filter.IDFromMsgBatch(msgBatch)).
					Str("pin", pin).
					Msgf("First ID of message batch matched filter")
			}
			sender := cmr.getSender(pin)
			err := sender.Send(msgBatch)
			if err != nil {
				cmr.Logger.Fatal().Err(err).Send()
				return err
			}
			if e := log.Debug(); e.Enabled() {
				e.Fields(filter.IDFromMsgBatch(msgBatch)).
					Str("sending to pin", pin).
					Msg("First ID of sent Message batch")
			}
		} else {
			if e := log.Debug(); e.Enabled() {
				e.Fields(filter.IDFromMsgBatch(msgBatch)).
					Str("Pin", pin).
					Msg("First ID of message batch didn't match filter")
			}
		}
	}
	return nil
}

func (cmr *CommonMessageRouter) SubscribeAllWithManualAck(listener *message.ConformationMessageListener, attributes ...string) (MQcommon.Monitor, error) {
	attrs := MQcommon.GetSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	pinFoundByAttrs := cmr.connManager.QConfig.FindQueuesByAttr(attrs)
	for queuePin, _ := range pinFoundByAttrs {
		cmr.Logger.Debug().Str("Pin", queuePin).Msg("Subscribing")
		subscriber, err := cmr.subByPinWithAck(listener, queuePin)
		if err != nil {
			cmr.Logger.Fatal().Err(err).Send()
			return SubscriberMonitor{}, err
		}
		subscribers = append(subscribers, subscriber)
	}
	var m sync.Mutex

	if len(subscribers) != 0 {
		for _, s := range subscribers {
			m.Lock()
			cmr.Logger.Debug().Str("Pin", s.subscriber.th2Pin).Msg("Start confirmation subscribing of queue")
			err := s.subscriber.ConfirmationStart()
			if err != nil {
				return SubscriberMonitor{}, err
			}
			m.Unlock()
		}

		return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
	} else {
		cmr.Logger.Fatal().Msg("No such subscriber")
	}
	return SubscriberMonitor{}, nil
}

func (cmr *CommonMessageRouter) SubscribeAll(listener *message.MessageListener, attributes ...string) (MQcommon.Monitor, error) {
	attrs := MQcommon.GetSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	pinsFoundByAttrs := cmr.connManager.QConfig.FindQueuesByAttr(attrs)
	for queuePin, _ := range pinsFoundByAttrs {
		cmr.Logger.Debug().Str("Pin", queuePin).Msg("Subscribing")
		subscriber, err := cmr.subByPin(listener, queuePin)
		if err != nil {
			cmr.Logger.Fatal().Err(err).Send()
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}
	var m sync.Mutex
	if len(subscribers) != 0 {
		for _, s := range subscribers {
			m.Lock()
			cmr.Logger.Debug().Str("Pin", s.subscriber.th2Pin).Msg("Start subscribing of queue")
			err := s.subscriber.Start()
			if err != nil {
				return SubscriberMonitor{}, err
			}
			m.Unlock()
		}
		return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
	} else {
		cmr.Logger.Fatal().Msg("No such subscriber")
	}
	return nil, nil
}

func (cmr *CommonMessageRouter) subByPin(listener *message.MessageListener, pin string) (SubscriberMonitor, error) {
	subscriber := cmr.getSubscriber(pin)
	subscriber.AddListener(listener)
	cmr.Logger.Debug().Str("Pin", pin).Msg("Getting subscriber monitor")
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) subByPinWithAck(listener *message.ConformationMessageListener, pin string) (SubscriberMonitor, error) {
	subscriber := cmr.getSubscriber(pin)
	subscriber.AddConfirmationListener(listener)
	cmr.Logger.Debug().Str("Pin", pin).Msg("Getting subscriber monitor")
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) getSubscriber(pin string) *CommonMessageSubscriber {
	queueConfig := cmr.connManager.QConfig.Queues[pin] // get queue by pin
	var result CommonMessageSubscriber
	if _, ok := cmr.subscribers[pin]; ok {
		result = cmr.subscribers[pin]
		cmr.Logger.Debug().Str("Pin", pin).Msgf("Getting already existing subscriber")
		return &result
	} else {
		result = CommonMessageSubscriber{connManager: cmr.connManager, qConfig: &queueConfig,
			listener: nil, confirmationListener: nil, th2Pin: pin, Logger: zerolog.New(os.Stdout).With().Str("component", "rabbitmq_message_subscriber").Timestamp().Logger()}
		cmr.subscribers[pin] = result
		cmr.Logger.Debug().Str("Pin", pin).Msg("Created subscriber")
		return &result
	}
}

func (cmr *CommonMessageRouter) getSender(pin string) *CommonMessageSender {
	queueConfig := cmr.connManager.QConfig.Queues[pin] // get queue by pin
	var result CommonMessageSender
	if _, ok := cmr.senders[pin]; ok {
		result = cmr.senders[pin]
		cmr.Logger.Debug().Str("Pin", pin).Msg("Getting already existing sender")
		return &result
	} else {
		result = CommonMessageSender{ConnManager: cmr.connManager, exchangeName: queueConfig.Exchange,
			sendQueue: queueConfig.RoutingKey, th2Pin: pin, Logger: zerolog.New(os.Stdout).With().Str("component", "rabbitmq_message_sender").Timestamp().Logger()}
		cmr.senders[pin] = result
		cmr.Logger.Debug().Str("Pin", pin).Msg("Created sender")
		return &result
	}
}
