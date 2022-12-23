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
	p_buff "github.com/th2-net/th2-common-go/proto"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/message"
	"github.com/th2-net/th2-common-go/schema/queue/message/configuration"
	"log"
	"strings"
)

type CommonMessageRouter struct {
	connManager *MQcommon.ConnectionManager
	subscribers map[string]CommonMessageSubscriber
	senders     map[string]CommonMessageSender
}

func (cmr *CommonMessageRouter) Construct(manager *MQcommon.ConnectionManager) {
	cmr.connManager = manager
	cmr.subscribers = map[string]CommonMessageSubscriber{}
	cmr.senders = map[string]CommonMessageSender{}
}

func (cmr *CommonMessageRouter) Close() {
	_ = cmr.connManager.Close()
}

func (cmr *CommonMessageRouter) SendAll(MsgBatch *p_buff.MessageGroupBatch, attributes ...string) error {
	attrs := cmr.getSendAttributes(attributes)
	pinsFoundByAttrs := cmr.connManager.QConfig.FindQueuesByAttr(attrs)
	pinsAndMessageGroup := cmr.getMessageGroupWithPins(pinsFoundByAttrs, MsgBatch)
	if len(pinsAndMessageGroup) != 0 {
		for pin, messageGroup := range pinsAndMessageGroup {
			sender := cmr.getSender(pin)
			err := sender.Send(messageGroup)
			if err != nil {
				log.Fatalln(err)
				return err
			}
		}
	} else {
		log.Fatalln("no such queue to send message")
	}
	return nil

}

func (cmr *CommonMessageRouter) SubscribeWithManualAck(listener *message.ConformationMessageListener, attributes ...string) (MQcommon.Monitor, error) {
	attrs := cmr.getSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	pinFoundByAttrs := cmr.connManager.QConfig.FindQueuesByAttr(attrs)
	for queuePin, _ := range pinFoundByAttrs {
		log.Printf("Subscrubing %v \n", queuePin)
		subscriber, err := cmr.subByPinWithAck(listener, queuePin)
		if err != nil {
			log.Fatalln(err)
			return SubscriberMonitor{}, err
		}
		subscribers = append(subscribers, subscriber)
	}
	if len(subscribers) != 0 {

		return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
	} else {
		log.Fatalln("no subscriber ")
	}
	return SubscriberMonitor{}, nil
}

func (cmr *CommonMessageRouter) SubscribeAll(listener *message.MessageListener, attributes ...string) (MQcommon.Monitor, error) {
	attrs := cmr.getSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	pinsFoundByAttrs := cmr.connManager.QConfig.FindQueuesByAttr(attrs)
	for queuePin, _ := range pinsFoundByAttrs {
		log.Printf("Subscrubing %v \n", queuePin)
		subscriber, err := cmr.subByPin(listener, queuePin)
		if err != nil {
			log.Fatalln(err)
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}
	if len(subscribers) != 0 {
		return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
	} else {
		log.Fatalln("no subscriber ")
	}
	return nil, nil
}

func (cmr *CommonMessageRouter) getSendAttributes(attrs []string) []string {
	res := []string{}
	if len(attrs) == 0 {
		return res
	} else {
		attrMap := make(map[string]bool)
		for _, attr := range attrs {
			attrMap[strings.ToLower(attr)] = true
		}
		attrMap["publish"] = true
		for k, _ := range attrMap {
			res = append(res, k)
		}
		return res
	}
}

func (cmr *CommonMessageRouter) getSubscribeAttributes(attrs []string) []string {
	res := []string{}
	if len(attrs) == 0 {
		return res
	} else {
		attrMap := make(map[string]bool)
		for _, attr := range attrs {
			attrMap[strings.ToLower(attr)] = true
		}
		attrMap["subscribe"] = true
		for k, _ := range attrMap {
			res = append(res, k)
		}
		return res
	}
}

func (cmr *CommonMessageRouter) subByPin(listener *message.MessageListener, pin string) (SubscriberMonitor, error) {
	subscriber := cmr.getSubscriber(pin)
	subscriber.AddListener(listener)
	err := subscriber.Start()
	if err != nil {
		return SubscriberMonitor{}, err
	}
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) subByPinWithAck(listener *message.ConformationMessageListener, alias string) (SubscriberMonitor, error) {
	subscriber := cmr.getSubscriber(alias)
	subscriber.AddConfirmationListener(listener)
	err := subscriber.ConfirmationStart()
	if err != nil {
		return SubscriberMonitor{}, err
	}
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) getSubscriber(pin string) *CommonMessageSubscriber {
	queueConfig := cmr.connManager.QConfig.Queues[pin] // get queue by alias
	var result CommonMessageSubscriber
	if _, ok := cmr.subscribers[pin]; ok {
		result = cmr.subscribers[pin]
		return &result
	} else {
		result = CommonMessageSubscriber{connManager: cmr.connManager, qConfig: &queueConfig,
			listener: nil, confirmationListener: nil, th2Pin: pin}
		cmr.subscribers[pin] = result
		return &result
	}
}

func (cmr *CommonMessageRouter) getSender(pin string) *CommonMessageSender {
	queueConfig := cmr.connManager.QConfig.Queues[pin] // get queue by pin
	var result CommonMessageSender
	if _, ok := cmr.senders[pin]; ok {
		result = cmr.senders[pin]
		return &result
	} else {
		result = CommonMessageSender{ConnManager: cmr.connManager, exchangeName: queueConfig.Exchange,
			sendQueue: queueConfig.RoutingKey, th2Pin: pin}
		cmr.senders[pin] = result

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
