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

package event

import (
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	p_buff "github.com/th2-net/th2-common-go/pkg/common/grpc/th2_grpc_common"
	"github.com/th2-net/th2-common-go/pkg/log"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/connection"

	"github.com/streadway/amqp"
	"github.com/th2-net/th2-common-go/pkg/metrics"
	"github.com/th2-net/th2-common-go/pkg/queue/event"
	"google.golang.org/protobuf/proto"
)

var (
	errNoListener = errors.New("no listener to handle delivery")
)

var th2EventSubscribeTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_event_subscribe_total",
		Help: "Amount of events received",
	},
	[]string{metrics.DefaultTh2PinLabelName},
)

func newSubscriber(
	manager *connection.Manager,
	config *queue.DestinationConfig,
	pinName string,
	subscriberType internal.SubscriberType,
) (internal.Subscriber, error) {
	logger := log.ForComponent("rabbitmq_event_subscriber")
	baseHandler := baseEventHandler{&logger, pinName}
	switch subscriberType {
	case internal.AutoSubscriberType:
		return internal.NewAutoSubscriber(
			manager,
			config,
			pinName,
			&autoEventHandler{baseEventHandler: baseHandler},
			metrics.EventTh2Type,
		), nil
	case internal.ManualSubscriberType:
		return internal.NewManualSubscriber(
			manager,
			config,
			pinName,
			&confirmationEventHandler{baseEventHandler: baseHandler},
			metrics.EventTh2Type,
		), nil
	}
	return nil, fmt.Errorf("unsupported subscriber type: %d", subscriberType)
}

type baseEventHandler struct {
	logger *zerolog.Logger
	th2Pin string
}

type autoEventHandler struct {
	baseEventHandler
	listener event.Listener
}

func (cs *autoEventHandler) Close() error {
	listener := cs.listener
	if listener == nil {
		return nil
	}
	cs.listener = nil
	return listener.OnClose()
}

type confirmationEventHandler struct {
	baseEventHandler
	listener event.ConformationListener
}

func (cs *confirmationEventHandler) Close() error {
	listener := cs.listener
	if listener == nil {
		return nil
	}
	cs.listener = nil
	return listener.OnClose()
}

func (cs *autoEventHandler) Handle(msgDelivery amqp.Delivery) error {
	listener := cs.listener
	if listener == nil {
		cs.logger.Error().
			Str("routingKey", msgDelivery.RoutingKey).
			Str("exchange", msgDelivery.Exchange).
			Msg("No Listener to Handle")
		return errNoListener
	}
	result := &p_buff.EventBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.logger.Error().
			Err(err).
			Str("routingKey", msgDelivery.RoutingKey).
			Str("exchange", msgDelivery.Exchange).
			Msg("Can't unmarshal proto")
		return err
	}
	th2EventSubscribeTotal.WithLabelValues(cs.th2Pin).Add(float64(len(result.Events)))
	delivery := queue.Delivery{Redelivered: msgDelivery.Redelivered}
	handleErr := listener.Handle(delivery, result)
	if handleErr != nil {
		cs.logger.Error().Err(handleErr).
			Str("routingKey", msgDelivery.RoutingKey).
			Str("exchange", msgDelivery.Exchange).
			Msg("Can't Handle")
		return handleErr
	}
	cs.logger.Debug().
		Str("routingKey", msgDelivery.RoutingKey).
		Str("exchange", msgDelivery.Exchange).
		Msg("Successfully Handled")
	return nil
}

func (cs *confirmationEventHandler) Handle(msgDelivery amqp.Delivery, timer *prometheus.Timer) error {
	listener := cs.listener
	if listener == nil {
		cs.logger.Error().Msg("No Confirmation Listener to Handle")
		return errors.New("no Confirmation Listener to Handle")
	}

	result := &p_buff.EventBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.logger.Error().Err(err).Msg("Can't unmarshal proto")
		return err
	}
	th2EventSubscribeTotal.WithLabelValues(cs.th2Pin).Add(float64(len(result.Events)))
	delivery := queue.Delivery{Redelivered: msgDelivery.Redelivered}
	deliveryConfirm := internal.DeliveryConfirmation{Delivery: &msgDelivery, Logger: log.ForComponent("confirmation"), Timer: timer}
	var confirmation queue.Confirmation = &deliveryConfirm

	handleErr := listener.Handle(delivery, result, confirmation)
	if handleErr != nil {
		cs.logger.Error().Err(handleErr).Msg("Can't Handle")
		return handleErr
	}
	cs.logger.Debug().Msg("Successfully Handled")
	return nil
}

func (cs *autoEventHandler) SetListener(listener event.Listener) {
	cs.listener = listener
	cs.logger.Trace().Msg("set listener")
}

func (cs *confirmationEventHandler) SetListener(listener event.ConformationListener) {
	cs.listener = listener
	cs.logger.Trace().Msg("set confirmation listener")
}
