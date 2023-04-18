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
	"github.com/th2-net/th2-common-go/pkg/common"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/connection"
	"io"
)

const (
	connectionConfigFilename = "rabbitMQ"
)

type rabbitMqImpl struct {
	closer io.Closer
	baseImpl
}

func (impl rabbitMqImpl) Close() error {
	// FIXME: error aggregation
	impl.baseImpl.Close()
	return impl.closer.Close()
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
	messageRouter, eventRouter, manager, err := rabbitmq.NewRouters(connConfiguration, &queueConfiguration)
	if err != nil {
		return nil, err
	}

	return &rabbitMqImpl{
		closer:   manager,
		baseImpl: baseImpl{messageRouter: messageRouter, eventRouter: eventRouter},
	}, nil
}
