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

package rabbitmq

import (
	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/pkg/common"
	"github.com/th2-net/th2-common-go/pkg/log"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/event"
	"github.com/th2-net/th2-common-go/pkg/queue/message"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/connection"
	internal "github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/connection"
	eventImpl "github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/event"
	messageImpl "github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/message"
	"io"
)

func NewRouters(
	boxConfig common.BoxConfig,
	connection connection.Config,
	config *queue.RouterConfig,
) (messageRouter message.Router, eventRouter event.Router, closer io.Closer, err error) {
	manager, err := internal.NewConnectionManager(connection, boxConfig.Name, log.ForComponent("connection_manager"))
	if err != nil {
		return
	}
	go manager.ListenForBlockingNotifications()
	messageRouter = newMessageRouter(&manager, config, log.ForComponent("message_router"))
	eventRouter = newEventRouter(&manager, config, log.ForComponent("event_router"))
	closer = &manager
	return
}

func newMessageRouter(
	manager *internal.Manager,
	config *queue.RouterConfig,
	logger zerolog.Logger,
) message.Router {
	return messageImpl.NewRouter(manager, config, logger)
}

func newEventRouter(
	manager *internal.Manager,
	config *queue.RouterConfig,
	logger zerolog.Logger,
) event.Router {
	return eventImpl.NewRouter(manager, config, logger)
}
