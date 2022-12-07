package schema

import (
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	factory "github.com/th2-net/th2-common-go/schema/factory"
	msg "github.com/th2-net/th2-common-go/schema/message/configuration"
	mq "github.com/th2-net/th2-common-go/schema/message/impl/rabbitmq/configuration"
	"os"
	"path/filepath"
	"testing"
)

var rabbit, _ = filepath.Abs("../resources/rabbitmq.json")
var router, _ = filepath.Abs("../resources/routermq.json")

func TestNewCommonFactoryFromArgs(t *testing.T) {
	cf, creationErr := factory.NewCommonFactoryFromArgs(rabbit, router)
	assert.Nil(t, creationErr)
	initErr := cf.Init()
	assert.Nil(t, initErr)
}

func TestRabbitMQConfig(t *testing.T) {
	bytes, _ := os.ReadFile(rabbit)
	rc := mq.RabbitMQConfiguration{}
	err := rc.Init(bytes)
	assert.Nil(t, err)
}

func TestMessageRouterConfig(t *testing.T) {
	bytes, _ := os.ReadFile(router)
	mc := msg.MessageRouterConfiguration{}
	initErr := mc.Init(string(bytes))
	assert.Nil(t, initErr)
	_, invalidAliasErr := mc.GetQueueByAlias("not_exist")
	assert.NotNil(t, invalidAliasErr)
	_, validAliasErr := mc.GetQueueByAlias("test_queue")
	assert.Nil(t, validAliasErr)
}

func testMcDeserializationFailure(t *testing.T, path string) msg.MessageRouterConfiguration {
	invalidRouter, _ := filepath.Abs(path)
	bytes, _ := os.ReadFile(invalidRouter)
	mc := msg.MessageRouterConfiguration{
		Logger: log.Logger.With().Str("component", "MessageRouterConfiguration").Logger(),
	}
	err := mc.Init(string(bytes))
	assert.NotNil(t, err)
	return mc
}

///*
//*
//Tests if invalid name of routingKey field causes deserialization to fail
//Expected behavior: It causes deserialization to fail, err is not nil
//*/
//func TestInvalidMessageRouterConfig1(t *testing.T) {
//	testMcDeserializationFailure(t, "resources/invalidroutermq.json")
//}
//
//func TestInvalidMessageRouterConfig2(t *testing.T) {
//	testMcDeserializationFailure(t, "resources/invalidroutermq2.json")
//
//}
