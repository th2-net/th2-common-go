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
}

func (m *RabbitMQModule) GetKey() common.ModuleKey {
	return queueModuleKey
}

var queueModuleKey = common.ModuleKey("queue")

func NewRabbitMQModule(provider factory.ConfigProvider) common.Module {

	queueConfiguration := configuration.MessageRouterConfiguration{}
	err := provider.GetConfig("routermq", &queueConfiguration)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("queue config : %v \n", queueConfiguration)
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

	return &RabbitMQModule{
		connManager:     connectionManager,
		MqMessageRouter: message.CommonMessageRouter{ConnManager: connectionManager}}
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
