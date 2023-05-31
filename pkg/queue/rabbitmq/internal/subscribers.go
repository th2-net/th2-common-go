/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
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

package internal

import (
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/connection"
	"io"
	"sync"
)

var DoubleStartError = errors.New("the subscription already started")

type SubscriberType = int

const (
	AutoSubscriberType SubscriberType = iota
	ManualSubscriberType
)

func NewAutoSubscriber(
	manager *connection.Manager,
	config *queue.DestinationConfig,
	pinName string,
	handler AutoHandler,
	metric string,
) AutoSubscriber {
	return &autoSubscriber{
		subscriber: subscriber{
			connManager:  manager,
			qConfig:      config,
			th2Pin:       pinName,
			metricsLabel: metric,
			lock:         &sync.RWMutex{},
		},
		handler: handler,
	}
}

func NewManualSubscriber(
	manager *connection.Manager,
	config *queue.DestinationConfig,
	pinName string,
	handler ConfirmationHandler,
	metric string,
) Subscriber {
	return &confirmationSubscriber{
		subscriber: subscriber{
			connManager:  manager,
			qConfig:      config,
			th2Pin:       pinName,
			metricsLabel: metric,
			lock:         &sync.RWMutex{},
		},
		handler: handler,
	}
}

type subscriber struct {
	connManager  *connection.Manager
	qConfig      *queue.DestinationConfig
	th2Pin       string
	metricsLabel string

	lock    *sync.RWMutex
	started bool
}

func (cs *subscriber) Pin() string {
	return cs.th2Pin
}

type AutoHandler interface {
	Handle(delivery amqp.Delivery) error
	io.Closer
}

type ConfirmationHandler interface {
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
	GetHandler() AutoHandler
}

type ManualSubscriber interface {
	Subscriber
	GetHandler() ConfirmationHandler
}

type autoSubscriber struct {
	subscriber
	handler AutoHandler
}

func (cs *autoSubscriber) GetHandler() AutoHandler {
	return cs.handler
}

type confirmationSubscriber struct {
	subscriber
	handler ConfirmationHandler
}

func (cs *confirmationSubscriber) GetHandler() ConfirmationHandler {
	return cs.handler
}

func (cs *autoSubscriber) Start() error {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	if cs.started {
		return DoubleStartError
	}
	err := cs.connManager.Consumer.Consume(cs.qConfig.QueueName, cs.th2Pin, cs.metricsLabel, cs.handler.Handle)
	if err != nil {
		return err
	}
	cs.started = true
	return nil
	//use th2Pin for metrics
}

func (cs *autoSubscriber) Close() error {
	defer cs.Stop()
	return cs.handler.Close()
}

func (cs *confirmationSubscriber) Start() error {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	if cs.started {
		return DoubleStartError
	}
	err := cs.connManager.Consumer.ConsumeWithManualAck(cs.qConfig.QueueName, cs.th2Pin, cs.metricsLabel, cs.handler.Handle)
	if err != nil {
		return err
	}
	cs.started = true
	return nil
	//use th2Pin for metrics
}

func (cs *confirmationSubscriber) Close() error {
	defer cs.Stop()
	return cs.handler.Close()
}

func (cs *subscriber) IsStarted() bool {
	cs.lock.RLock()
	defer cs.lock.RUnlock()
	return cs.started
}

func (cs *subscriber) Stop() {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	cs.started = false
}

type SubscriberMonitor interface {
	queue.Monitor
	GetSubscriber() Subscriber
}

func MonitorFor(subscriber Subscriber) SubscriberMonitor {
	return subscriberMonitor{subscriber: subscriber}
}

type subscriberMonitor struct {
	subscriber Subscriber
}

func (sub subscriberMonitor) GetSubscriber() Subscriber {
	return sub.subscriber
}

func (sub subscriberMonitor) Unsubscribe() error {

	err := sub.subscriber.Close()
	if err != nil {
		return err
	}

	return nil
}

type MultiplySubscribeMonitor struct {
	SubscriberMonitors []SubscriberMonitor
}

func (sub MultiplySubscribeMonitor) Unsubscribe() error {
	var errs []error
	for _, subM := range sub.SubscriberMonitors {
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

func SubscribeAll[T any](
	router T,
	pinFoundByAttrs map[string]queue.DestinationConfig,
	logger *zerolog.Logger,
	subscribeFunc func(router T, pinName string) (SubscriberMonitor, error),
) ([]SubscriberMonitor, error) {
	subscribers := make([]SubscriberMonitor, 0, len(pinFoundByAttrs))
	for queuePin, _ := range pinFoundByAttrs {
		logger.Debug().Str("Pin", queuePin).Msg("Subscribing")
		subscriber, err := subscribeFunc(router, queuePin)
		if err != nil {
			logger.Error().Err(err).Str("Pin", queuePin).Msg("cannot subscribe")
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}
	return subscribers, nil
}

func StartAll(subscribers []SubscriberMonitor, logger *zerolog.Logger) error {
	for _, s := range subscribers {
		logger.Trace().Str("Pin", s.GetSubscriber().Pin()).Msg("Start subscribing of queue")
		if s.GetSubscriber().IsStarted() {
			logger.Trace().Str("Pin", s.GetSubscriber().Pin()).Msg("subscriber already started")
			continue
		}
		err := s.GetSubscriber().Start()
		if err != nil {
			if err == DoubleStartError {
				logger.Info().Str("Pin", s.GetSubscriber().Pin()).Msg("already started")
				continue
			}
			logger.Error().Err(err).Str("Pin", s.GetSubscriber().Pin()).
				Msg("cannot start subscriber")
			return err
		}
	}
	return nil
}
