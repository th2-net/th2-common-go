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
	"os"
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"github.com/streadway/amqp"
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
	[]string{"th2Pin", "session_alias", "direction", "message_type"},
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
		cs.Logger.Fatal().Err(err).Msg("Can't unmarshal proto")
	}
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	if cs.listener == nil {
		cs.Logger.Fatal().Msgf("No Listener to Handle : %s ", cs.listener)
	}
	handleErr := (*cs.listener).Handle(&delivery, result)
	if handleErr != nil {
		cs.Logger.Fatal().Err(handleErr).Msg("Can't Handle")
	}
	for _, group := range result.Groups {
		for _, msg := range group.Messages {
			switch msg.GetKind().(type) {
			case *p_buff.AnyMessage_RawMessage:
				msg := msg.GetRawMessage()
				th2_message_subscribe_total.WithLabelValues(cs.th2Pin, msg.Metadata.Id.ConnectionId.SessionAlias, string(msg.Metadata.Id.Direction), RAW_MESSAGE_TYPE).Inc()
			case *p_buff.AnyMessage_Message:
				msg := msg.GetMessage()
				th2_message_subscribe_total.WithLabelValues(cs.th2Pin, msg.Metadata.Id.ConnectionId.SessionAlias, string(msg.Metadata.Id.Direction), MESSAGE_TYPE).Inc()
			}
		}
	}
	cs.Logger.Debug().Msg("Successfully Handled")
}

func (cs *CommonMessageSubscriber) ConfirmationHandler(msgDelivery amqp.Delivery, timer *prometheus.Timer) {
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.Logger.Fatal().Err(err).Msg("Can't unmarshal proto")
	}
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	deliveryConfirm := MQcommon.DeliveryConfirmation{Delivery: &msgDelivery, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(), Timer: timer}
	var confirmation MQcommon.Confirmation = deliveryConfirm

	if cs.confirmationListener == nil {
		cs.Logger.Fatal().Msgf("No Confirmation Listener to Handle : %s ", cs.confirmationListener)
	}
	handleErr := (*cs.confirmationListener).Handle(&delivery, result, &confirmation)
	if handleErr != nil {
		cs.Logger.Fatal().Err(handleErr).Msg("Can't Handle")
	}
	for _, group := range result.Groups {
		for _, msg := range group.Messages {
			switch msg.GetKind().(type) {
			case *p_buff.AnyMessage_RawMessage:
				msg := msg.GetRawMessage()
				th2_message_subscribe_total.WithLabelValues(cs.th2Pin, msg.Metadata.Id.ConnectionId.SessionAlias, string(msg.Metadata.Id.Direction), RAW_MESSAGE_TYPE).Inc()
			case *p_buff.AnyMessage_Message:
				msg := msg.GetMessage()
				th2_message_subscribe_total.WithLabelValues(cs.th2Pin, msg.Metadata.Id.ConnectionId.SessionAlias, string(msg.Metadata.Id.Direction), MESSAGE_TYPE).Inc()
			}
		}
	}
	cs.Logger.Debug().Msg("Successfully Handled")
}

func (cs *CommonMessageSubscriber) Start() error {
	err := cs.connManager.Consumer.Consume(cs.qConfig.QueueName, cs.th2Pin, TH2_TYPE, cs.Handler)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonMessageSubscriber) ConfirmationStart() error {
	err := cs.connManager.Consumer.ConsumeWithManualAck(cs.qConfig.QueueName, cs.th2Pin, TH2_TYPE, cs.ConfirmationHandler)
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

func (cs *CommonMessageSubscriber) AddListener(listener *message.MessageListener) {
	cs.listener = listener
	cs.Logger.Debug().Msg("Added listener")
}

func (cs *CommonMessageSubscriber) AddConfirmationListener(listener *message.ConformationMessageListener) {
	cs.confirmationListener = listener
	cs.Logger.Debug().Msg("Added confirmation listener")
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
