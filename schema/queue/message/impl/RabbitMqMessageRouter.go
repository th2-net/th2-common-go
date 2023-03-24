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
	"github.com/th2-net/th2-common-go/schema/filter"
	defaultStrategy "github.com/th2-net/th2-common-go/schema/filter/impl"
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
	cmr.filterStrategy = defaultStrategy.Default
}

func (cmr *CommonMessageRouter) Close() {
	_ = cmr.connManager.Close()
}

func (cmr *CommonMessageRouter) SendAll(msgBatch *p_buff.MessageGroupBatch, attributes ...string) error {
	attrs := MQcommon.GetSendAttributes(attributes)
	pinsFoundByAttrs := cmr.connManager.QConfig.FindQueuesByAttr(attrs)
	if len(pinsFoundByAttrs) != 0 {
		for pin, config := range pinsFoundByAttrs {
			if config.Verify(msgBatch) {
				sender := cmr.getSender(pin)
				err := sender.Send(msgBatch)
				if err != nil {
					cmr.Logger.Fatal().Err(err).Send()
					return err
				}
			}
		}
	} else {
		cmr.Logger.Fatal().Msg("no such queue to send message")
	}
	return nil

}

func (cmr *CommonMessageRouter) SubscribeAllWithManualAck(listener *message.ConformationMessageListener, attributes ...string) (MQcommon.Monitor, error) {
	attrs := MQcommon.GetSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	pinFoundByAttrs := cmr.connManager.QConfig.FindQueuesByAttr(attrs)
	for queuePin, _ := range pinFoundByAttrs {
		cmr.Logger.Debug().Msgf("Subscribing %s ", queuePin)
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
		cmr.Logger.Debug().Msgf("Subscribing %s ", queuePin)
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
	cmr.Logger.Debug().Msgf("Getting subscriber monitor for pin %s", pin)
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) subByPinWithAck(listener *message.ConformationMessageListener, pin string) (SubscriberMonitor, error) {
	subscriber := cmr.getSubscriber(pin)
	subscriber.AddConfirmationListener(listener)
	cmr.Logger.Debug().Msgf("Getting subscriber monitor for pin %s", pin)
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) getSubscriber(pin string) *CommonMessageSubscriber {
	queueConfig := cmr.connManager.QConfig.Queues[pin] // get queue by pin
	var result CommonMessageSubscriber
	if _, ok := cmr.subscribers[pin]; ok {
		result = cmr.subscribers[pin]
		cmr.Logger.Debug().Msgf("Getting already existing subscriber for pin %s", pin)
		return &result
	} else {
		result = CommonMessageSubscriber{connManager: cmr.connManager, qConfig: &queueConfig,
			listener: nil, confirmationListener: nil, th2Pin: pin, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
		cmr.subscribers[pin] = result
		cmr.Logger.Debug().Msgf("Created subscriber for pin %s", pin)
		return &result
	}
}

func (cmr *CommonMessageRouter) getSender(pin string) *CommonMessageSender {
	queueConfig := cmr.connManager.QConfig.Queues[pin] // get queue by pin
	var result CommonMessageSender
	if _, ok := cmr.senders[pin]; ok {
		result = cmr.senders[pin]
		cmr.Logger.Debug().Msgf("Getting already existing sender for pin %s", pin)
		return &result
	} else {
		result = CommonMessageSender{ConnManager: cmr.connManager, exchangeName: queueConfig.Exchange,
			sendQueue: queueConfig.RoutingKey, th2Pin: pin, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
		cmr.senders[pin] = result
		cmr.Logger.Debug().Msgf("Created sender for pin %s", pin)
		return &result
	}
}
