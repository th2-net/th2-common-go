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

package queue

import (
	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/pkg/common"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/connection"
	"os"
)

const (
	connectionConfigFilename = "rabbitMQ"
)

type rabbitMqImpl struct {
	connManager *connection.Manager
	baseImpl
}

func (impl rabbitMqImpl) Close() error {
	// FIXME: error aggregation
	impl.baseImpl.Close()
	return impl.connManager.Close()
}

func newRabbitMq(
	provider common.ConfigProvider,
	queueConfiguration queue.RouterConfig,
) (Module, error) {
	connConfiguration := connection.Config{}
	configErr := provider.GetConfig(connectionConfigFilename, &connConfiguration)
	if configErr != nil {
		return nil, configErr
	}
	return NewRabbitMq(connConfiguration, queueConfiguration)
}

func NewRabbitMq(
	connConfiguration connection.Config,
	queueConfiguration queue.RouterConfig,
) (Module, error) {
	managerLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	connectionManager, err := connection.NewConnectionManager(connConfiguration, managerLogger)
	if err != nil {
		return nil, err
	}

	messageRouterLogger := zerolog.New(os.Stdout).With().Str("component", "mq_message_router").Timestamp().Logger()
	messageRouter := rabbitmq.NewMessageRouter(&connectionManager, &queueConfiguration, messageRouterLogger)

	eventRouterLogger := zerolog.New(os.Stdout).With().Str("component", "mq_event_router").Timestamp().Logger()
	eventRouter := rabbitmq.NewEventRouter(&connectionManager, &queueConfiguration, eventRouterLogger)

	return &rabbitMqImpl{
		connManager: &connectionManager,
		baseImpl:    baseImpl{messageRouter: messageRouter, eventRouter: eventRouter},
	}, nil
}
