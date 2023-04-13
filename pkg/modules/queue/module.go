/*
 * Copyright 2022-2023 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package queue

import (
	"fmt"
	"github.com/th2-net/th2-common-go/pkg/common"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/event"
	"github.com/th2-net/th2-common-go/pkg/queue/message"
	"reflect"
)

const (
	routerConfigFilename = "mq"
	moduleKey            = "queue"
)

type Module interface {
	common.Module
	GetEventRouter() event.Router
	GetMessageRouter() message.Router
}

type baseImpl struct {
	messageRouter message.Router
	eventRouter   event.Router
}

func (m *baseImpl) GetEventRouter() event.Router {
	return m.eventRouter
}

func (m *baseImpl) GetMessageRouter() message.Router {
	return m.messageRouter
}

func (m *baseImpl) GetKey() common.ModuleKey {
	return queueModuleKey
}
func (m *baseImpl) Close() error {
	// FIXME: aggregate errors
	m.messageRouter.Close()
	m.eventRouter.Close()
	return nil
}

var queueModuleKey = common.ModuleKey(moduleKey)

func NewRabbitMqModule(provider common.ConfigProvider) (common.Module, error) {
	queueConfiguration := queue.RouterConfig{}
	err := provider.GetConfig(routerConfigFilename, &queueConfiguration)
	if err != nil {
		return nil, err
	}
	return newRabbitMq(provider, queueConfiguration)
}

type Identity struct{}

func (id *Identity) GetModule(factory common.Factory) (Module, error) {
	module, err := factory.Get(queueModuleKey)
	if err != nil {
		return nil, err
	}
	casted, success := module.(*rabbitMqImpl)
	if !success {
		return nil, fmt.Errorf("module with key %s is a %s", queueModuleKey, reflect.TypeOf(module))
	}
	return casted, nil
}

var ModuleID = &Identity{}
