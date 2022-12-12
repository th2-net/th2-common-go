package th2

import (
	cfg "github.com/th2-net/th2-common-go/schema/message/configuration"
	rmq_connection "github.com/th2-net/th2-common-go/schema/message/impl/rabbitmq/connection"
)

type MessageRouter interface {
	Init(*rmq_connection.ConnectionManager, *cfg.MessageRouterConfiguration) error

	Subscribe() error

	SubscribeAll() error

	//Send message to <b>SOME</b> RabbitMQ queues which match the filter for this message
	Send(interface{}) error

	//Send message to <b>ONE</b> RabbitMQ queue by intersection schemas queues attributes
	SendByQueueAttributes(interface{}, cfg.QueueAttribute) error

	//Send message to <b>SOME</b> RabbitMQ queue by intersection schemas queues attributes
	SendAll(interface{}, cfg.QueueAttribute) error
}
