/*
 * Copyright 2022-2023 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package message

import (
	"errors"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/filter"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/connection"
	"os"
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"github.com/streadway/amqp"
	"github.com/th2-net/th2-common-go/pkg/metrics"
	"github.com/th2-net/th2-common-go/pkg/queue/message"
	"google.golang.org/protobuf/proto"
)

var th2MessageSubscribeTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_message_subscribe_total",
		Help: "Quantity of incoming messages",
	},
	[]string{
		metrics.DefaultTh2PinLabelName,
		metrics.DefaultSessionAliasLabelName,
		metrics.DefaultDirectionLabelName,
		metrics.DefaultMessageTypeLabelName,
	},
)

type CommonMessageSubscriber struct {
	connManager          *connection.Manager
	qConfig              *queue.DestinationConfig
	listener             message.Listener
	confirmationListener message.ConformationListener
	rawListener          message.RawListener
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
	delivery := queue.Delivery{Redelivered: msgDelivery.Redelivered}
	metrics.UpdateMessageMetrics(result, th2MessageSubscribeTotal, cs.th2Pin)
	handleErr := cs.listener.Handle(delivery, result)
	if handleErr != nil {
		cs.Logger.Error().Err(handleErr).Str("Method", "Handler").Msg("Can't Handle")
		return handleErr
	}
	if e := cs.Logger.Debug(); e.Enabled() {
		e.Str("Method", "Handler").
			Interface("MessageID", filter.FirstIDFromMsgBatch(result)).
			Msgf("First message ID of message batch that handled successfully")
	}
	return nil
}

func (cs *CommonMessageSubscriber) HandlerRaw(msgDelivery amqp.Delivery) error {
	listener := cs.rawListener
	if listener == nil {
		return errors.New("no Listener to handle")
	}
	delivery := queue.Delivery{Redelivered: msgDelivery.Redelivered}
	handleErr := listener.Handle(delivery, msgDelivery.Body)
	if handleErr != nil {
		cs.Logger.Error().Err(handleErr).Str("Method", "HandlerRaw").Msg("Can't Handle")
		return handleErr
	}
	if e := cs.Logger.Debug(); e.Enabled() {
		e.Str("Method", "HandlerRaw").
			Bytes("Data", msgDelivery.Body).
			Msgf("Batch has been processed")
	}
	return nil
}

func (cs *CommonMessageSubscriber) ConfirmationHandler(msgDelivery amqp.Delivery, timer *prometheus.Timer) error {
	if cs.confirmationListener == nil {
		return errors.New("no Confirmation Listener to Handle")
	}
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.Logger.Error().Err(err).Str("Method", "ConfirmationHandler").Msg("Can't unmarshal proto")
		return nil
	}
	delivery := queue.Delivery{Redelivered: msgDelivery.Redelivered}
	deliveryConfirm := internal.DeliveryConfirmation{Delivery: &msgDelivery, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(), Timer: timer}

	if cs.confirmationListener == nil {
		cs.Logger.Error().Str("Method", "ConfirmationHandler").Msgf("No Confirmation Listener to Handle : %s ", cs.confirmationListener)
		return errors.New("no Confirmation Listener to Handle")
	}
	metrics.UpdateMessageMetrics(result, th2MessageSubscribeTotal, cs.th2Pin)
	handleErr := cs.confirmationListener.Handle(delivery, result, &deliveryConfirm)
	if handleErr != nil {
		cs.Logger.Error().Err(handleErr).Str("Method", "ConfirmationHandler").Msg("Can't Handle")
		return handleErr
	}
	if e := cs.Logger.Debug(); e.Enabled() {
		e.Str("Method", "ConfirmationHandler").
			Interface("MessageID", filter.FirstIDFromMsgBatch(result)).
			Msg("First message ID of message batch that was handled successfully")
	}
	return nil
}

func (cs *CommonMessageSubscriber) Start() error {
	err := cs.connManager.Consumer.Consume(cs.qConfig.QueueName, cs.th2Pin, metrics.MessageGroupTh2Type, cs.Handler)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonMessageSubscriber) StartRaw() error {
	err := cs.connManager.Consumer.Consume(cs.qConfig.QueueName, cs.th2Pin, metrics.MessageGroupTh2Type, cs.HandlerRaw)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonMessageSubscriber) ConfirmationStart() error {
	err := cs.connManager.Consumer.ConsumeWithManualAck(cs.qConfig.QueueName, cs.th2Pin, metrics.MessageGroupTh2Type, cs.ConfirmationHandler)
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

func (cs *CommonMessageSubscriber) SetListener(listener message.Listener) {
	cs.listener = listener
	cs.Logger.Trace().Msg("set listener")
}

func (cs *CommonMessageSubscriber) SetRawListener(listener message.RawListener) {
	cs.rawListener = listener
	cs.Logger.Trace().Msg("set raw listener")
}

func (cs *CommonMessageSubscriber) AddConfirmationListener(listener message.ConformationListener) {
	cs.confirmationListener = listener
	cs.Logger.Trace().Msg("Added confirmation listener")
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
