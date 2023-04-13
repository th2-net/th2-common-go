package rabbitmq

import (
	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/pkg/queue"
	"github.com/th2-net/th2-common-go/pkg/queue/event"
	"github.com/th2-net/th2-common-go/pkg/queue/message"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/connection"
	eventImpl "github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/event"
	messageImpl "github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/message"
)

func NewMessageRouter(
	manager *connection.Manager,
	config *queue.RouterConfig,
	logger zerolog.Logger,
) message.Router {
	return messageImpl.NewRouter(manager, config, logger)
}

func NewEventRouter(
	manager *connection.Manager,
	config *queue.RouterConfig,
	logger zerolog.Logger,
) event.Router {
	return eventImpl.NewRouter(manager, config, logger)
}
