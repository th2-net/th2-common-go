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
	"errors"
	"os"
	"sync"
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/rs/zerolog"

	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/configuration"
	"github.com/th2-net/th2-common-go/schema/queue/message"
)

type CommonMessageRouter struct {
	connManager *MQcommon.ConnectionManager
	subscribers map[string]CommonMessageSubscriber
	senders     map[string]CommonMessageSender

	Logger zerolog.Logger
}

func (cmr *CommonMessageRouter) Construct(manager *MQcommon.ConnectionManager) {
	cmr.connManager = manager
	cmr.subscribers = map[string]CommonMessageSubscriber{}
	cmr.senders = map[string]CommonMessageSender{}
	cmr.Logger.Debug().Msg("CommonMessageRouter was initialized")
}

func (cmr *CommonMessageRouter) Close() {
	_ = cmr.connManager.Close()
}

func (cmr *CommonMessageRouter) SendAll(MsgBatch *p_buff.MessageGroupBatch, attributes ...string) error {
	attrs := MQcommon.GetSendAttributes(attributes)
	pinsFoundByAttrs := cmr.connManager.QConfig.FindQueuesByAttr(attrs)
	pinsAndMessageGroup := cmr.getMessageGroupWithPins(pinsFoundByAttrs, MsgBatch)
	if len(pinsAndMessageGroup) != 0 {
		for pin, messageGroup := range pinsAndMessageGroup {
			sender := cmr.getSender(pin)
			err := sender.Send(messageGroup)
			if err != nil {
				return err
			}
		}
	} else {
		return errors.New("no such queue to send message")
	}
	return nil

}

func (cmr *CommonMessageRouter) SendRawAll(rawData []byte, attributes ...string) error {
	attrs := MQcommon.GetSendAttributes(attributes)
	pinsFoundByAttrs := cmr.connManager.QConfig.FindQueuesByAttr(attrs)
	if len(pinsFoundByAttrs) == 0 {
		return errors.New("no such queue to send message")
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

func (cmr *CommonMessageRouter) SubscribeAllWithManualAck(listener message.ConformationMessageListener, attributes ...string) (MQcommon.Monitor, error) {
	attrs := MQcommon.GetSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	pinFoundByAttrs := cmr.connManager.QConfig.FindQueuesByAttr(attrs)
	for queuePin, _ := range pinFoundByAttrs {
		cmr.Logger.Debug().Msgf("Subscribing %s ", queuePin)
		subscriber, err := cmr.subByPinWithAck(listener, queuePin)
		if err != nil {
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
		return SubscriberMonitor{}, errors.New("no such subscriber")
	}
	return SubscriberMonitor{}, nil
}

func (cmr *CommonMessageRouter) SubscribeAll(listener message.MessageListener, attributes ...string) (MQcommon.Monitor, error) {
	attrs := MQcommon.GetSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	pinsFoundByAttrs := cmr.connManager.QConfig.FindQueuesByAttr(attrs)
	for queuePin, _ := range pinsFoundByAttrs {
		cmr.Logger.Debug().Msgf("Subscribing %s ", queuePin)
		subscriber, err := cmr.subByPin(listener, queuePin)
		if err != nil {
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
				return nil, err
			}
			m.Unlock()
		}
		return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
	} else {
		return nil, errors.New("no such subscriber")
	}
	return nil, nil
}

func (cmr *CommonMessageRouter) subByPin(listener message.MessageListener, pin string) (SubscriberMonitor, error) {
	subscriber := cmr.getSubscriber(pin)
	subscriber.AddListener(listener)
	cmr.Logger.Debug().Msgf("Getting subscriber monitor for pin %s", pin)
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) subByPinWithAck(listener message.ConformationMessageListener, pin string) (SubscriberMonitor, error) {
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

func (cmr *CommonMessageRouter) getMessageGroupWithPins(queue map[string]configuration.QueueConfig, message *p_buff.MessageGroupBatch) map[string]*p_buff.MessageGroupBatch {
	//Here will be added filter handling
	result := make(map[string]*p_buff.MessageGroupBatch)
	for pin, _ := range queue {

		msgBatch := p_buff.MessageGroupBatch{}
		for _, messageGroup := range message.Groups {
			//doing filtering based on queue filters on message_group
			msgBatch.Groups = append(msgBatch.Groups, messageGroup)
		}

		result[pin] = &msgBatch
	}
	return result

}
