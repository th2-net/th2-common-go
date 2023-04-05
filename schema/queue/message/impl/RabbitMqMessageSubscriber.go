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
	"github.com/th2-net/th2-common-go/schema/filter"
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
	listener             *message.MessageListener
	confirmationListener *message.ConformationMessageListener
	th2Pin               string

	Logger zerolog.Logger
}

func (cs *CommonMessageSubscriber) Handler(msgDelivery amqp.Delivery) {
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.Logger.Error().Err(err).Str("Method", "Handler").Msg("Can't unmarshal proto")
		//TODO Return error to the caller side
	}
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	if cs.listener == nil {
		cs.Logger.Error().Str("Method", "Handler").Msgf("No Listener to Handle : %s ", cs.listener)
		//TODO Return error to the caller side
	}
	metrics.UpdateMessageMetrics(result, th2_message_subscribe_total, cs.th2Pin)
	handleErr := (*cs.listener).Handle(&delivery, result)
	if handleErr != nil {
		cs.Logger.Error().Err(handleErr).Str("Method", "Handler").Msg("Can't Handle")
		//TODO Return error to the caller side
	}
	cs.Logger.Info().Msg("Successfully Handled")
	if e := cs.Logger.Debug(); e.Enabled() {
		e.Str("Method", "Handler").
			Fields(filter.FirstIDFromMsgBatch(result)).
			Msgf("First message ID of message batch that handled successfully")
	}
}

func (cs *CommonMessageSubscriber) ConfirmationHandler(msgDelivery amqp.Delivery, timer *prometheus.Timer) {
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.Logger.Error().Err(err).Str("Method", "ConfirmationHandler").Msg("Can't unmarshal proto")
		//TODO Return error to the caller side
	}
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	deliveryConfirm := MQcommon.DeliveryConfirmation{Delivery: &msgDelivery, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(), Timer: timer}
	var confirmation MQcommon.Confirmation = deliveryConfirm

	if cs.confirmationListener == nil {
		cs.Logger.Error().Str("Method", "ConfirmationHandler").Msgf("No Confirmation Listener to Handle : %s ", cs.confirmationListener)
		//TODO Return error to the caller side
	}
	metrics.UpdateMessageMetrics(result, th2_message_subscribe_total, cs.th2Pin)
	handleErr := (*cs.confirmationListener).Handle(&delivery, result, &confirmation)
	if handleErr != nil {
		cs.Logger.Error().Err(handleErr).Str("Method", "ConfirmationHandler").Msg("Can't Handle")
		//TODO Return error to the caller side
	}
	if e := cs.Logger.Debug(); e.Enabled() {
		e.Str("Method", "ConfirmationHandler").
			Interface("MessageID", filter.FirstIDFromMsgBatch(result)).
			Msg("First message ID of message batch that was handled successfully")
	}
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
	cs.Logger.Trace().Msg("Removed listeners")
}

func (cs *CommonMessageSubscriber) AddListener(listener *message.MessageListener) {
	cs.listener = listener
	cs.Logger.Trace().Msg("Added listener")
}

func (cs *CommonMessageSubscriber) AddConfirmationListener(listener *message.ConformationMessageListener) {
	cs.confirmationListener = listener
	cs.Logger.Trace().Msg("Added confirmation listener")
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
