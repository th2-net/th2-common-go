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
	"log"
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/streadway/amqp"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/configuration"
	"github.com/th2-net/th2-common-go/schema/queue/message"
	"google.golang.org/protobuf/proto"
)

type CommonMessageSubscriber struct {
	connManager          *MQcommon.ConnectionManager
	qConfig              *configuration.QueueConfig
	listener             *message.MessageListener
	confirmationListener *message.ConformationMessageListener
	th2Pin               string
}

func (cs *CommonMessageSubscriber) Handler(msgDelivery amqp.Delivery) {
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		log.Fatalf("Cann't unmarshal : %v \n", err)
	}
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}

	fail := (*cs.listener).Handle(&delivery, result)
	if fail != nil {
		log.Fatalf("Cann't Handle : %v \n", fail)
	}
}

func (cs *CommonMessageSubscriber) ConfirmationHandler(msgDelivery amqp.Delivery) {
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		log.Fatalf("Cann't unmarshal : %v \n", err)
	}

	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	deliveryConfirm := MQcommon.DeliveryConfirmation{Delivery: &msgDelivery}
	var confirmation MQcommon.Confirmation = deliveryConfirm

	fail := (*cs.confirmationListener).Handle(&delivery, result, &confirmation)
	if fail != nil {
		log.Fatalf("Cann't Handle : %v \n", fail)
	}
}

func (cs *CommonMessageSubscriber) Start() error {
	err := cs.connManager.Consumer.Consume(cs.qConfig.QueueName, cs.Handler)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonMessageSubscriber) ConfirmationStart() error {
	err := cs.connManager.Consumer.ConsumeWithManualAck(cs.qConfig.QueueName, cs.ConfirmationHandler)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonMessageSubscriber) RemoveListener() {
	cs.listener = nil
	cs.confirmationListener = nil
	log.Println("Removed listeners")
}

func (cs *CommonMessageSubscriber) AddListener(listener *message.MessageListener) {
	cs.listener = listener
}

func (cs *CommonMessageSubscriber) AddConfirmationListener(listener *message.ConformationMessageListener) {
	cs.confirmationListener = listener
}

type SubscriberMonitor struct {
	subscriber *CommonMessageSubscriber
}

func (sub SubscriberMonitor) Unsubscribe() error {
	if sub.subscriber.listener != nil {
		err := (*sub.subscriber.listener).OnClose()
		if err != nil {
			return err
		}
		sub.subscriber.RemoveListener()
	}

	if sub.subscriber.confirmationListener != nil {
		err := (*sub.subscriber.confirmationListener).OnClose()
		if err != nil {
			return err
		}
		sub.subscriber.RemoveListener()
	}
	return nil
}

type MultiplySubscribeMonitor struct {
	subscriberMonitors []SubscriberMonitor
}

func (sub MultiplySubscribeMonitor) Unsubscribe() error {
	for _, subM := range sub.subscriberMonitors {
		err := subM.Unsubscribe()
		if err != nil {
			return err
		}
	}
	return nil
}
