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
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
	p_buff "github.com/th2-net/th2-common-go/pkg/common/grpc/th2_grpc_common"
	"github.com/th2-net/th2-common-go/pkg/log"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/filter"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/connection"

	amqp "github.com/rabbitmq/amqp091-go"
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

type contentType = int

const (
	parsedContentType contentType = iota
	rawContentType
)

func newSubscriber(
	manager *connection.Manager,
	config *queue.DestinationConfig,
	pinName string,
	subscriberType internal.SubscriberType,
	contentType contentType,
) (internal.Subscriber, error) {
	logger := log.ForComponent("rabbitmq_message_subscriber")
	baseHandler := baseMessageHandler{&logger, pinName}
	switch subscriberType {
	case internal.AutoSubscriberType:
		var handler internal.AutoHandler
		switch contentType {
		case parsedContentType:
			handler = &messageHandler{
				baseMessageHandler: baseHandler,
			}
		case rawContentType:
			handler = &rawMessageHandler{
				baseMessageHandler: baseHandler,
			}
		default:
			return nil, fmt.Errorf("unknown content type: %d", contentType)
		}
		return internal.NewAutoSubscriber(manager, config, pinName, handler, metrics.MessageGroupTh2Type), nil
	case internal.ManualSubscriberType:
		var handler internal.ConfirmationHandler
		switch contentType {
		case parsedContentType:
			handler = &confirmationMessageHandler{
				baseMessageHandler: baseHandler,
			}
		case rawContentType:
			return nil, errors.New("raw content is not supported for manual subscriber")
		default:
			return nil, fmt.Errorf("unknown content type: %d", contentType)
		}
		return internal.NewManualSubscriber(manager, config, pinName, handler, metrics.MessageGroupTh2Type), nil
	default:
		return nil, fmt.Errorf("unsupported subscriber type: %d", subscriberType)
	}
}

type baseMessageHandler struct {
	logger *zerolog.Logger
	th2Pin string
}

type rawMessageHandler struct {
	baseMessageHandler
	listener message.RawListener
}

func (cs *rawMessageHandler) Close() error {
	listener := cs.listener
	if listener == nil {
		return nil
	}
	cs.listener = nil
	return listener.OnClose()
}

type confirmationMessageHandler struct {
	baseMessageHandler
	listener message.ConformationListener
}

func (cs *confirmationMessageHandler) Close() error {
	listener := cs.listener
	if listener == nil {
		return nil
	}
	cs.listener = nil
	return listener.OnClose()
}

type messageHandler struct {
	baseMessageHandler
	listener message.Listener
}

func (cs *messageHandler) Close() error {
	listener := cs.listener
	if listener == nil {
		return nil
	}
	cs.listener = nil
	return listener.OnClose()
}

func (cs *messageHandler) Handle(msgDelivery amqp.Delivery) error {
	listener := cs.listener
	if listener == nil {
		return errors.New("no Listener to handle")
	}
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		return err
	}
	delivery := queue.Delivery{Redelivered: msgDelivery.Redelivered}
	metrics.UpdateMessageMetrics(result, th2MessageSubscribeTotal, cs.th2Pin)
	handleErr := listener.Handle(delivery, result)
	if handleErr != nil {
		cs.logger.Error().Err(handleErr).Str("Method", "Handler").Msg("Can't Handle")
		return handleErr
	}
	if e := cs.logger.Debug(); e.Enabled() {
		e.Str("Method", "Handler").
			Interface("MessageID", filter.FirstIDFromMsgBatch(result)).
			Msgf("First message ID of message batch that handled successfully")
	}
	return nil
}

func (cs *rawMessageHandler) Handle(msgDelivery amqp.Delivery) error {
	listener := cs.listener
	if listener == nil {
		return errors.New("no Listener to handle")
	}
	delivery := queue.Delivery{Redelivered: msgDelivery.Redelivered}
	handleErr := listener.Handle(delivery, msgDelivery.Body)
	if handleErr != nil {
		cs.logger.Error().Err(handleErr).Str("Method", "HandlerRaw").Msg("Can't Handle")
		return handleErr
	}
	if e := cs.logger.Debug(); e.Enabled() {
		e.Str("Method", "HandlerRaw").
			Bytes("Data", msgDelivery.Body).
			Msgf("Batch has been processed")
	}
	return nil
}

func (cs *confirmationMessageHandler) Handle(msgDelivery amqp.Delivery, timer *prometheus.Timer) error {
	listener := cs.listener
	if listener == nil {
		return errors.New("no Confirmation Listener to Handle")
	}
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.logger.Error().Err(err).Str("Method", "ConfirmationHandler").Msg("Can't unmarshal proto")
		return nil
	}
	delivery := queue.Delivery{Redelivered: msgDelivery.Redelivered}
	deliveryConfirm := internal.DeliveryConfirmation{Delivery: &msgDelivery, Logger: log.ForComponent("confirmation"), Timer: timer}

	metrics.UpdateMessageMetrics(result, th2MessageSubscribeTotal, cs.th2Pin)
	handleErr := listener.Handle(delivery, result, &deliveryConfirm)
	if handleErr != nil {
		cs.logger.Error().Err(handleErr).Str("Method", "ConfirmationHandler").Msg("Can't Handle")
		return handleErr
	}
	if e := cs.logger.Debug(); e.Enabled() {
		e.Str("Method", "ConfirmationHandler").
			Interface("MessageID", filter.FirstIDFromMsgBatch(result)).
			Msg("First message ID of message batch that was handled successfully")
	}
	return nil
}

func (cs *messageHandler) SetListener(listener message.Listener) {
	cs.listener = listener
	cs.logger.Trace().Msg("set listener")
}

func (cs *rawMessageHandler) SetListener(listener message.RawListener) {
	cs.listener = listener
	cs.logger.Trace().Msg("set raw listener")
}

func (cs *confirmationMessageHandler) SetListener(listener message.ConformationListener) {
	cs.listener = listener
	cs.logger.Trace().Msg("Added confirmation listener")
}
