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
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"github.com/streadway/amqp"
	"github.com/th2-net/th2-common-go/schema/metrics"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/configuration"
	"github.com/th2-net/th2-common-go/schema/queue/message"
	"google.golang.org/protobuf/proto"
)

var th2_message_subscribe_total = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_message_subscribe_total",
		Help: "Quantity of incoming messages",
	},
	[]string{
		metrics.DEFAULT_TH2_PIN_LABEL_NAME,
		metrics.DEFAULT_SESSION_ALIAS_LABEL_NAME,
		metrics.DEFAULT_DIRECTION_LABEL_NAME,
		metrics.DEFAULT_MESSAGE_TYPE_LABEL_NAME,
	},
)

type CommonMessageSubscriber struct {
	connManager          *MQcommon.ConnectionManager
	qConfig              *configuration.QueueConfig
	listener             message.MessageListener
	confirmationListener message.ConformationMessageListener
	th2Pin               string

	Logger zerolog.Logger
}

func (cs *CommonMessageSubscriber) Handler(msgDelivery amqp.Delivery) error {
	if cs.listener == nil {
		return errors.New("no Listener to handle")
	}
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		return err
	}
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	metrics.UpdateMessageMetrics(result, th2_message_subscribe_total, cs.th2Pin)
	handleErr := cs.listener.Handle(&delivery, result)
	if handleErr != nil {
		return handleErr
	}
	cs.Logger.Debug().Msg("Successfully Handled")
	return nil
}

func (cs *CommonMessageSubscriber) ConfirmationHandler(msgDelivery amqp.Delivery, timer *prometheus.Timer) error {
	if cs.confirmationListener == nil {
		return errors.New("no Confirmation Listener to Handle")
	}
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		return nil
	}
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	deliveryConfirm := MQcommon.DeliveryConfirmation{Delivery: &msgDelivery, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(), Timer: timer}
	var confirmation MQcommon.Confirmation = deliveryConfirm

	metrics.UpdateMessageMetrics(result, th2_message_subscribe_total, cs.th2Pin)
	handleErr := cs.confirmationListener.Handle(&delivery, result, &confirmation)
	if handleErr != nil {
		return handleErr
	}
	cs.Logger.Debug().Msg("Successfully Handled")
	return nil
}

func (cs *CommonMessageSubscriber) Start() error {
	err := cs.connManager.Consumer.Consume(cs.qConfig.QueueName, cs.th2Pin, metrics.MESSAGE_GROUP_TH2_TYPE, cs.Handler)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonMessageSubscriber) ConfirmationStart() error {
	err := cs.connManager.Consumer.ConsumeWithManualAck(cs.qConfig.QueueName, cs.th2Pin, metrics.MESSAGE_GROUP_TH2_TYPE, cs.ConfirmationHandler)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonMessageSubscriber) RemoveListener() {
	cs.listener = nil
	cs.confirmationListener = nil
	cs.Logger.Info().Msg("Removed listeners")
}

func (cs *CommonMessageSubscriber) AddListener(listener message.MessageListener) {
	cs.listener = listener
	cs.Logger.Debug().Msg("Added listener")
}

func (cs *CommonMessageSubscriber) AddConfirmationListener(listener message.ConformationMessageListener) {
	cs.confirmationListener = listener
	cs.Logger.Debug().Msg("Added confirmation listener")
}

type SubscriberMonitor struct {
	subscriber *CommonMessageSubscriber
}

func (sub SubscriberMonitor) Unsubscribe() error {
	if sub.subscriber.listener != nil {
		err := sub.subscriber.listener.OnClose()
		if err != nil {
			return err
		}
		sub.subscriber.RemoveListener()
	}

	if sub.subscriber.confirmationListener != nil {
		err := sub.subscriber.confirmationListener.OnClose()
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
