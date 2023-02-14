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

	"github.com/rs/zerolog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/streadway/amqp"
	"github.com/th2-net/th2-common-go/schema/metrics"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/configuration"
	"github.com/th2-net/th2-common-go/schema/queue/message"
	"google.golang.org/protobuf/proto"
)

var INCOMING_MESSAGE_SIZE = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_rabbitmq_message_size_subscribe_bytes",
		Help: "Amount of bytes received",
	},
	metrics.SUBSCRIBER_LABELS,
)

var HANDLING_DURATION = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "th2_rabbitmq_message_process_duration_seconds",
		Help:    "Subscriber's handling process duration",
		Buckets: metrics.DEFAULT_BUCKETS,
	},
	metrics.SUBSCRIBER_LABELS,
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
	timer := prometheus.NewTimer(HANDLING_DURATION.WithLabelValues(cs.th2Pin, metrics.TH2_TYPE, cs.qConfig.QueueName))
	defer timer.ObserveDuration()
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.Logger.Fatal().Err(err).Msg("Can't unmarshal proto")
	}
	INCOMING_MESSAGE_SIZE.WithLabelValues(cs.th2Pin, metrics.TH2_TYPE, cs.qConfig.QueueName).Add(float64(len(msgDelivery.Body)))
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	if cs.listener == nil {
		cs.Logger.Fatal().Msgf("No Listener to Handle : %s ", cs.listener)
	}
	handleErr := (*cs.listener).Handle(&delivery, result)
	if handleErr != nil {
		cs.Logger.Fatal().Err(handleErr).Msg("Can't Handle")
	}
	cs.Logger.Debug().Msg("Successfully Handled")
}

func (cs *CommonMessageSubscriber) ConfirmationHandler(msgDelivery amqp.Delivery) {
	timer := prometheus.NewTimer(HANDLING_DURATION.WithLabelValues(cs.th2Pin, metrics.TH2_TYPE, cs.qConfig.QueueName))
	defer timer.ObserveDuration()
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.Logger.Fatal().Err(err).Msg("Can't unmarshal proto")
	}
	INCOMING_MESSAGE_SIZE.WithLabelValues(cs.th2Pin, metrics.TH2_TYPE, cs.qConfig.QueueName).Add(float64(len(msgDelivery.Body)))
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	deliveryConfirm := MQcommon.DeliveryConfirmation{Delivery: &msgDelivery, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
	var confirmation MQcommon.Confirmation = deliveryConfirm

	if cs.confirmationListener == nil {
		cs.Logger.Fatal().Msgf("No Confirmation Listener to Handle : %s ", cs.confirmationListener)
	}
	handleErr := (*cs.confirmationListener).Handle(&delivery, result, &confirmation)
	if handleErr != nil {
		cs.Logger.Fatal().Err(handleErr).Msg("Can't Handle")
	}
	cs.Logger.Debug().Msg("Successfully Handled")
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
