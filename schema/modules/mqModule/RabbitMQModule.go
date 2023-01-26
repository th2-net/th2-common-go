/*
 * Copyright 2022 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package mqModule

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/schema/common"
	"github.com/th2-net/th2-common-go/schema/factory"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/configuration"
	event "github.com/th2-net/th2-common-go/schema/queue/event/impl"
	message "github.com/th2-net/th2-common-go/schema/queue/message/impl"
)

const (
	RABBIT_MQ_CONFIG_FILENAME = "rabbitMQ"
	MQ_ROUTER_CONFIG_FILENAME = "mq"
	RABBIT_MQ_MODULE_KEY      = "queue"
)

type RabbitMQModule struct {
	MqMessageRouter message.CommonMessageRouter
	connManager     MQcommon.ConnectionManager
	MqEventRouter   event.CommonEventRouter
}

func (m *RabbitMQModule) GetKey() common.ModuleKey {
	return queueModuleKey
}
func (m *RabbitMQModule) Close() {
	m.MqMessageRouter.Close()
	m.MqEventRouter.Close()
}

var queueModuleKey = common.ModuleKey(RABBIT_MQ_MODULE_KEY)

func NewRabbitMQModule(provider factory.ConfigProvider) common.Module {

	queueConfiguration := configuration.MessageRouterConfiguration{Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
	err := provider.GetConfig(MQ_ROUTER_CONFIG_FILENAME, &queueConfiguration)
	if err != nil {
		queueConfiguration.Logger.Fatal().Err(err)
	}
	connConfiguration := configuration.RabbitMQConfiguration{Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
	configErr := provider.GetConfig(RABBIT_MQ_CONFIG_FILENAME, &connConfiguration)
	if configErr != nil {
		connConfiguration.Logger.Fatal().Err(configErr)
	}
	connectionManager := MQcommon.ConnectionManager{QConfig: &queueConfiguration, MqConnConfig: &connConfiguration,
		Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
	port, portErr := strconv.Atoi(connectionManager.MqConnConfig.Port)
	if err != nil {
		connectionManager.Logger.Fatal().Err(portErr)
	}
	connectionManager.Url = fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		connectionManager.MqConnConfig.Username,
		connectionManager.MqConnConfig.Password,
		connectionManager.MqConnConfig.Host,
		port,
		connectionManager.MqConnConfig.VHost)
	connectionManager.Construct()

	messageRouter := message.CommonMessageRouter{Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
	messageRouter.Construct(&connectionManager)

	eventRouter := event.CommonEventRouter{Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
	eventRouter.Construct(&connectionManager)

	return &RabbitMQModule{connManager: connectionManager,
		MqMessageRouter: messageRouter, MqEventRouter: eventRouter}
}

type Identity struct{}

func (id *Identity) GetModule(factory *factory.CommonFactory) (*RabbitMQModule, error) {
	module, err := factory.Get(queueModuleKey)
	if err != nil {
		return nil, err
	}
	casted, success := module.(*RabbitMQModule)
	if !success {
		return nil, fmt.Errorf("module with key %s is a %s", queueModuleKey, reflect.TypeOf(module))
	}
	return casted, nil
}

var ModuleID = &Identity{}
