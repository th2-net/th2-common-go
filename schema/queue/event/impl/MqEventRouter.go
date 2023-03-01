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
package event

import (
	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/event"
	"log"
	"os"
	"sync"
	p_buff "th2-grpc/th2_grpc_common"
)

type CommonEventRouter struct {
	connManager *MQcommon.ConnectionManager
	subscribers map[string]CommonEventSubscriber
	senders     map[string]CommonEventSender

	Logger zerolog.Logger
}

func (cer *CommonEventRouter) Construct(manager *MQcommon.ConnectionManager) {
	cer.connManager = manager
	cer.subscribers = map[string]CommonEventSubscriber{}
	cer.senders = map[string]CommonEventSender{}
	cer.Logger.Debug().Msg("CommonEventRouter was initialized")
}

func (cer *CommonEventRouter) Close() {
	_ = cer.connManager.Close()
}

func (cer *CommonEventRouter) SendAll(EventBatch *p_buff.EventBatch, attributes ...string) error {
	attrs := MQcommon.GetSendAttributes(attributes)
	pinsFoundByAttrs := cer.connManager.QConfig.FindQueuesByAttr(attrs)
	if len(pinsFoundByAttrs) != 0 {
		for pin, _ := range pinsFoundByAttrs {
			sender := cer.getSender(pin)
			err := sender.Send(EventBatch)
			if err != nil {
				cer.Logger.Fatal().Err(err).Send()
				return err
			}
		}
	} else {
		cer.Logger.Fatal().Msg("no such queue to send message")
	}
	return nil

}

func (cer *CommonEventRouter) SubscribeAll(listener *event.EventListener, attributes ...string) (MQcommon.Monitor, error) {
	attrs := MQcommon.GetSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	pinsFoundByAttrs := cer.connManager.QConfig.FindQueuesByAttr(attrs)
	for queuePin, _ := range pinsFoundByAttrs {
		cer.Logger.Debug().Msgf("Subscribing %s ", queuePin)
		subscriber, err := cer.subByPin(listener, queuePin)
		if err != nil {
			cer.Logger.Fatal().Err(err).Send()
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
				log.Printf("CONSUMING ERROR : %v \n", err)
				return SubscriberMonitor{}, err
			}
			m.Unlock()
		}
		return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
	} else {
		cer.Logger.Fatal().Msg("No such subscriber")
	}
	return nil, nil
}

func (cer *CommonEventRouter) SubscribeAllWithManualAck(listener *event.ConformationEventListener, attributes ...string) (MQcommon.Monitor, error) {
	attrs := MQcommon.GetSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	pinFoundByAttrs := cer.connManager.QConfig.FindQueuesByAttr(attrs)
	for queuePin, _ := range pinFoundByAttrs {
		cer.Logger.Debug().Msgf("Subscribing %s ", queuePin)
		subscriber, err := cer.subByPinWithAck(listener, queuePin)
		if err != nil {
			cer.Logger.Fatal().Err(err).Send()
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
				log.Printf("CONSUMING ERROR : %v \n", err)
				return SubscriberMonitor{}, err
			}
			m.Unlock()
		}

		return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
	} else {
		cer.Logger.Fatal().Msg("No such subscriber")
	}
	return SubscriberMonitor{}, nil
}

func (cer *CommonEventRouter) subByPin(listener *event.EventListener, pin string) (SubscriberMonitor, error) {
	subscriber := cer.getSubscriber(pin)
	subscriber.AddListener(listener)
	cer.Logger.Debug().Msgf("Getting subscriber monitor for pin %s", pin)
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cer *CommonEventRouter) subByPinWithAck(listener *event.ConformationEventListener, pin string) (SubscriberMonitor, error) {
	subscriber := cer.getSubscriber(pin)
	subscriber.AddConfirmationListener(listener)
	cer.Logger.Debug().Msgf("Getting subscriber monitor for pin %s", pin)
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cer *CommonEventRouter) getSubscriber(pin string) *CommonEventSubscriber {
	queueConfig := cer.connManager.QConfig.Queues[pin] // get queue by pin
	var result CommonEventSubscriber
	if _, ok := cer.subscribers[pin]; ok {
		result = cer.subscribers[pin]
		cer.Logger.Debug().Msgf("Getting already existing subscriber for pin %s", pin)
		return &result
	} else {
		result = CommonEventSubscriber{connManager: cer.connManager, qConfig: &queueConfig,
			listener: nil, confirmationListener: nil, th2Pin: pin, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
		cer.subscribers[pin] = result
		cer.Logger.Debug().Msgf("Created subscriber for pin %s", pin)
		return &result
	}
}

func (cer *CommonEventRouter) getSender(pin string) *CommonEventSender {
	queueConfig := cer.connManager.QConfig.Queues[pin] // get queue by pin
	var result CommonEventSender
	if _, ok := cer.senders[pin]; ok {
		result = cer.senders[pin]
		cer.Logger.Debug().Msgf("Getting already existing sender for pin %s", pin)
		return &result
	} else {
		result = CommonEventSender{ConnManager: cer.connManager, exchangeName: queueConfig.Exchange,
			sendQueue: queueConfig.RoutingKey, th2Pin: pin, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
		cer.senders[pin] = result
		cer.Logger.Debug().Msgf("Created sender for pin %s", pin)
		return &result
	}
}
