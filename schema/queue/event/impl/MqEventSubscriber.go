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
	"github.com/th2-net/th2-common-go/schema/queue/event"
	"google.golang.org/protobuf/proto"
)

var th2_event_subscribe_total = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_event_subscribe_total",
		Help: "Amount of events received",
	},
	[]string{metrics.DEFAULT_TH2_PIN_LABEL_NAME},
)

type CommonEventSubscriber struct {
	connManager          *MQcommon.ConnectionManager
	qConfig              *configuration.QueueConfig
	listener             *event.EventListener
	confirmationListener *event.ConformationEventListener
	th2Pin               string

	Logger zerolog.Logger
}

func (cs *CommonEventSubscriber) Start() error {
	err := cs.connManager.Consumer.Consume(cs.qConfig.QueueName, cs.th2Pin, metrics.EVENT_TH2_TYPE, cs.Handler)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonEventSubscriber) ConfirmationStart() error {
	err := cs.connManager.Consumer.ConsumeWithManualAck(cs.qConfig.QueueName, cs.th2Pin, metrics.EVENT_TH2_TYPE, cs.ConfirmationHandler)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonEventSubscriber) Handler(msgDelivery amqp.Delivery) error {
	result := &p_buff.EventBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.Logger.Error().Err(err).Msg("Can't unmarshal proto")
		//Maybe it is better to return err itself?
		return errors.New("can't unmarshal proto")
	}
	th2_event_subscribe_total.WithLabelValues(cs.th2Pin).Add(float64(len(result.Events)))
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	if cs.listener == nil {
		cs.Logger.Error().Msg("No Listener to Handle")
		return errors.New("no Listener to handle delivery")
	}
	handleErr := (*cs.listener).Handle(&delivery, result)
	if handleErr != nil {
		cs.Logger.Error().Err(handleErr).Msg("Can't Handle")
		return handleErr
	}
	cs.Logger.Debug().Msg("Successfully Handled")
	return nil
}

func (cs *CommonEventSubscriber) ConfirmationHandler(msgDelivery amqp.Delivery, timer *prometheus.Timer) error {
	result := &p_buff.EventBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.Logger.Error().Err(err).Msg("Can't unmarshal proto")
		return err
	}
	th2_event_subscribe_total.WithLabelValues(cs.th2Pin).Add(float64(len(result.Events)))
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	deliveryConfirm := MQcommon.DeliveryConfirmation{Delivery: &msgDelivery, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(), Timer: timer}
	var confirmation MQcommon.Confirmation = deliveryConfirm

	if cs.confirmationListener == nil {
		cs.Logger.Error().Msg("No Confirmation Listener to Handle")
		return errors.New("no Confirmation Listener to Handle")
	}
	handleErr := (*cs.confirmationListener).Handle(&delivery, result, &confirmation)
	if handleErr != nil {
		cs.Logger.Error().Err(handleErr).Msg("Can't Handle")
		return handleErr
	}
	cs.Logger.Debug().Msg("Successfully Handled")
	return nil
}

func (cs *CommonEventSubscriber) RemoveListener() {
	cs.listener = nil
	cs.confirmationListener = nil
	cs.Logger.Trace().Msg("Removed listeners")
}

func (cs *CommonEventSubscriber) AddListener(listener *event.EventListener) {
	cs.listener = listener
	cs.Logger.Trace().Msg("Added listener")
}

func (cs *CommonEventSubscriber) AddConfirmationListener(listener *event.ConformationEventListener) {
	cs.confirmationListener = listener
	cs.Logger.Trace().Msg("Added confirmation listener")
}

type SubscriberMonitor struct {
	subscriber *CommonEventSubscriber
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
