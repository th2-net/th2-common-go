package modules

import (
	"fmt"
	"github.com/th2-net/th2-common-go/schema/common"
	"github.com/th2-net/th2-common-go/schema/factory"
	"github.com/th2-net/th2-common-go/schema/messages/configuration"
	message "github.com/th2-net/th2-common-go/schema/messages/impl"
	"log"
	"reflect"
)

type ConnectionManager struct {
	QConfig      *configuration.MessageRouterConfiguration
	MqConnConfig *configuration.RabbitMQConfiguration
}

type QueueModule struct {
	MqMessageRouter message.CommonMessageRouter
	connManager     ConnectionManager
}

func (m *QueueModule) GetKey() common.ModuleKey {
	return queueModuleKey
}

var queueModuleKey = common.ModuleKey("queue")

func NewQueueModule(provider factory.ConfigProvider) common.Module {

	queueConfiguration := configuration.MessageRouterConfiguration{}
	err := provider.GetConfig("routermq.json", &queueConfiguration)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("queue config : %v \n", queueConfiguration)
	connConfiguration := configuration.RabbitMQConfiguration{}
	fail := provider.GetConfig("rabbitmq.json", &connConfiguration)
	if fail != nil {
		log.Fatalln(fail)
	}
	connectionManager := ConnectionManager{QConfig: &queueConfiguration, MqConnConfig: &connConfiguration}

	return &QueueModule{connManager: connectionManager, MqMessageRouter: message.CommonMessageRouter{}}
}

type Identity struct{}

func (id *Identity) GetModule(factory *factory.CommonFactory) (*QueueModule, error) {
	module, err := factory.Get(queueModuleKey)
	if err != nil {
		return nil, err
	}
	casted, success := module.(*QueueModule)
	if !success {
		return nil, fmt.Errorf("module with key %s is a %s", queueModuleKey, reflect.TypeOf(module))
	}
	return casted, nil
}

var ModuleID = &Identity{}
