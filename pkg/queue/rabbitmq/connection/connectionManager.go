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
	"github.com/rs/zerolog"
	"os"
)

type Manager struct {
	Publisher *Publisher
	Consumer  *Consumer

	Logger zerolog.Logger
}

func NewConnectionManager(connConfiguration Config, logger zerolog.Logger) (Manager, error) {
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		connConfiguration.Username,
		connConfiguration.Password,
		connConfiguration.Host,
		connConfiguration.Port,
		connConfiguration.VHost)
	publisher, err := NewPublisher(url, zerolog.New(os.Stdout).With().Str("component", "publisher").Timestamp().Logger())
	if err != nil {
		return Manager{}, err
	}
	consumer, err := NewConsumer(url, zerolog.New(os.Stdout).With().Str("component", "consumer").Timestamp().Logger())
	if err != nil {
		if pubErr := publisher.Close(); pubErr != nil {
			logger.Err(pubErr).
				Msg("error when closing publisher")
		}
		return Manager{}, err
	}
	return Manager{
		Publisher: &publisher,
		Consumer:  &consumer,
		Logger:    logger,
	}, nil
}

func (manager *Manager) Close() error {
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
