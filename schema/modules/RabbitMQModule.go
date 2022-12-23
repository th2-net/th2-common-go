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

package modules

import (
	"fmt"
	"github.com/th2-net/th2-common-go/schema/common"
	"github.com/th2-net/th2-common-go/schema/factory"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/message/configuration"
	"github.com/th2-net/th2-common-go/schema/queue/message/impl"
	"log"
	"reflect"
	"strconv"
)

type RabbitMQModule struct {
	MqMessageRouter message.CommonMessageRouter
	connManager     MQcommon.ConnectionManager
	// event router
}

func (m *RabbitMQModule) GetKey() common.ModuleKey {
	return queueModuleKey
}
func (m *RabbitMQModule) Close() {
	m.MqMessageRouter.Close()
	//same for event Router
}

var queueModuleKey = common.ModuleKey("queue")

func NewRabbitMQModule(provider factory.ConfigProvider) common.Module {

	queueConfiguration := configuration.MessageRouterConfiguration{}
	err := provider.GetConfig("routermq", &queueConfiguration)
	if err != nil {
		log.Fatalln(err)
	}
	connConfiguration := configuration.RabbitMQConfiguration{}
	fail := provider.GetConfig("rabbitmq", &connConfiguration)
	if fail != nil {
		log.Fatalln(fail)
	}
	connectionManager := MQcommon.ConnectionManager{QConfig: &queueConfiguration, MqConnConfig: &connConfiguration}
	port, err := strconv.Atoi(connectionManager.MqConnConfig.Port)
	if err != nil {
		log.Fatalf("%v", err)
	}
	connectionManager.Url = fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		connectionManager.MqConnConfig.Username,
		connectionManager.MqConnConfig.Password,
		connectionManager.MqConnConfig.Host,
		port,
		connectionManager.MqConnConfig.VHost)
	connectionManager.Construct()

	messageRouter := message.CommonMessageRouter{}
	messageRouter.Construct(&connectionManager)

	return &RabbitMQModule{connManager: connectionManager,
		MqMessageRouter: messageRouter}
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
