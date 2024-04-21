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

package connection

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/pkg/log"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/connection"
)

type Manager struct {
	Publisher *Publisher
	Consumer  *Consumer

	Logger zerolog.Logger
	closed chan struct{}
}

func NewConnectionManager(connConfiguration connection.Config, logger zerolog.Logger) (Manager, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		connConfiguration.Username,
		connConfiguration.Password,
		connConfiguration.Host,
		connConfiguration.Port,
		connConfiguration.VHost)
	publisher, err := NewPublisher(url, connConfiguration, log.ForComponent("publisher"))
	if err != nil {
		return Manager{}, err
	}
	consumer, err := NewConsumer(url, connConfiguration, log.ForComponent("consumer"))
	if err != nil {
		if pubErr := publisher.Close(); pubErr != nil {
			logger.Err(pubErr).
				Msg("error when closing publisher")
		}
		return Manager{}, err
	}
	go publisher.runConnectionRoutine()
	// TODO: run consumer connection routine
	return Manager{
		Publisher: &publisher,
		Consumer:  &consumer,
		Logger:    logger,
		// capacity is one to avoid blocking close call
		closed: make(chan struct{}),
	}, nil
}

func (manager *Manager) ListenForBlockingNotifications() {
	var run = true
	var consumerClosed = true
	var publisherClosed = true

	var publisherNotifications <-chan amqp.Blocking
	var consumerNotifications <-chan amqp.Blocking
	for run {
		// We try to reinitialize the listeners on each iteration
		// because in case of connection problems
		// the connections for publisher and consumer will be recreated.
		// Old channels will be closed and never receive a new value
		if publisherClosed {
			publisherNotifications = manager.Publisher.registerBlockingListener(make(chan amqp.Blocking, 1))
			publisherClosed = false
		}
		if consumerClosed {
			consumerNotifications = manager.Consumer.registerBlockingListener(make(chan amqp.Blocking, 1))
			consumerClosed = false
		}
		select {
		case <-manager.closed:
			manager.Logger.Info().Msg("stop listening for blocking notifications")
			run = false
			break
		case consumerBlocked, ok := <-consumerNotifications:
			if !ok {
				consumerClosed = true
				break
			}
			manager.Logger.Warn().
				Str("reason", consumerBlocked.Reason).
				Bool("active", consumerBlocked.Active).
				Msg("received blocked notification for consumer")
		case publisherBlocked, ok := <-publisherNotifications:
			if !ok {
				publisherClosed = true
				break
			}
			manager.Logger.Warn().
				Str("reason", publisherBlocked.Reason).
				Bool("active", publisherBlocked.Active).
				Msg("received blocked notification for publisher")
		}
	}
}

func (manager *Manager) Close() error {
	close(manager.closed)

	err := manager.Publisher.Close()
	if err != nil {
		manager.Logger.Error().Err(err).Msg("cannot close publisher")
		return err
	}

	fail := manager.Consumer.Close()
	if fail != nil {
		manager.Logger.Error().Err(err).Msg("cannot close consumer")
		return fail
	}

	manager.Logger.Info().Msg("connections closed gracefully")
	return nil
}
