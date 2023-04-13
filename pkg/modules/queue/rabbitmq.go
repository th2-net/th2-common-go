package queue

import (
	"fmt"
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
	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		connConfiguration.Username,
		connConfiguration.Password,
		connConfiguration.Host,
		connConfiguration.Port,
		connConfiguration.VHost)
	managerLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	connectionManager, err := connection.NewConnectionManager(url, &managerLogger)
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
