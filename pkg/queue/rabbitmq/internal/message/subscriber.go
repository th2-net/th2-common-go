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
	"github.com/th2-net/th2-common-go/pkg/common"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/filter"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/connection"
	"io"
	"os"
	"sync"
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

var DoubleStartError = errors.New("the subscription already started")

type subscriberType = int

const (
	autoSubscriber subscriberType = iota
	manualSubscriber
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
	subscriberType subscriberType,
	contentType contentType,
) (Subscriber, error) {
	logger := zerolog.New(os.Stdout).With().
		Str(common.ComponentLoggerKey, "rabbitmq_message_subscriber").
		Timestamp().
		Logger()
	base := commonMessageSubscriber{
		connManager: manager,
		qConfig:     config,
		th2Pin:      pinName,
		lock:        &sync.RWMutex{},
	}
	baseHandler := baseMessageHandler{&logger, pinName}
	switch subscriberType {
	case autoSubscriber:
		var handler handler
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
		return &messageSubscriber{
			commonMessageSubscriber: base,
			handler:                 handler,
		}, nil
	case manualSubscriber:
		var handler confirmationHandler
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
		return &confirmationMessageSubscriber{
			commonMessageSubscriber: base,
			handler:                 handler,
		}, nil
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

type commonMessageSubscriber struct {
	connManager *connection.Manager
	qConfig     *queue.DestinationConfig
	th2Pin      string

	lock    *sync.RWMutex
	started bool
}

func (cs *commonMessageSubscriber) Pin() string {
	return cs.th2Pin
}

type handler interface {
	Handle(delivery amqp.Delivery) error
	io.Closer
}

type confirmationHandler interface {
	Handle(delivery amqp.Delivery, timer *prometheus.Timer) error
	io.Closer
}

type Subscriber interface {
	IsStarted() bool
	Start() error
	Pin() string
	io.Closer
}

type AutoSubscriber interface {
	Subscriber
	getHandler() handler
}

type ManualSubscriber interface {
	Subscriber
	getHandler() confirmationHandler
}

type messageSubscriber struct {
	commonMessageSubscriber
	handler handler
}

func (cs *messageSubscriber) getHandler() handler {
	return cs.handler
}

type confirmationMessageSubscriber struct {
	commonMessageSubscriber
	handler confirmationHandler
}

func (cs *confirmationMessageSubscriber) getHandler() confirmationHandler {
	return cs.handler
}

func (cs *messageHandler) Handle(msgDelivery amqp.Delivery) error {
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
	if cs.listener == nil {
		return errors.New("no Confirmation Listener to Handle")
	}
	result := &p_buff.MessageGroupBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.logger.Error().Err(err).Str("Method", "ConfirmationHandler").Msg("Can't unmarshal proto")
		return nil
	}
	delivery := queue.Delivery{Redelivered: msgDelivery.Redelivered}
	deliveryConfirm := internal.DeliveryConfirmation{Delivery: &msgDelivery, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger(), Timer: timer}

	if cs.listener == nil {
		cs.logger.Error().Str("Method", "ConfirmationHandler").Msgf("No Confirmation Listener to Handle : %s ", cs.listener)
		return errors.New("no Confirmation Listener to Handle")
	}
	metrics.UpdateMessageMetrics(result, th2MessageSubscribeTotal, cs.th2Pin)
	handleErr := cs.listener.Handle(delivery, result, &deliveryConfirm)
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

func (cs *messageSubscriber) Start() error {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	if cs.started {
		return DoubleStartError
	}
	err := cs.connManager.Consumer.Consume(cs.qConfig.QueueName, cs.th2Pin, metrics.MessageGroupTh2Type, cs.handler.Handle)
	if err != nil {
		return err
	}
	cs.started = true
	return nil
	//use th2Pin for metrics
}

func (cs *messageSubscriber) Close() error {
	return cs.handler.Close()
}

func (cs *confirmationMessageSubscriber) Start() error {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	if cs.started {
		return DoubleStartError
	}
	err := cs.connManager.Consumer.ConsumeWithManualAck(cs.qConfig.QueueName, cs.th2Pin, metrics.MessageGroupTh2Type, cs.handler.Handle)
	if err != nil {
		return err
	}
	cs.started = true
	return nil
	//use th2Pin for metrics
}

func (cs *confirmationMessageSubscriber) Close() error {
	return cs.handler.Close()
}

func (cs *commonMessageSubscriber) IsStarted() bool {
	cs.lock.RLock()
	defer cs.lock.RUnlock()
	return cs.started
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

type SubscriberMonitor struct {
	subscriber Subscriber
}

func (sub SubscriberMonitor) Unsubscribe() error {

	err := sub.subscriber.Close()
	if err != nil {
		return err
	}

	return nil
}

type MultiplySubscribeMonitor struct {
	subscriberMonitors []SubscriberMonitor
}

func (sub MultiplySubscribeMonitor) Unsubscribe() error {
	var errs []error
	for _, subM := range sub.subscriberMonitors {
		err := subM.Unsubscribe()
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return fmt.Errorf("errors during unsubsribing: %v", errs)
	}
	return nil
}
