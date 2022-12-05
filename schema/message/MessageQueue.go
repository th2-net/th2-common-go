package th2

import (
	cfg "exactpro/th2/th2-common-go/schema/message/configuration"
	rmq_connection "exactpro/th2/th2-common-go/schema/message/impl/rabbitmq/connection"
)

type MessageQueue interface {
	Init(*rmq_connection.ConnectionManager, *cfg.QueueConfiguration) error

	GetSender() *MessageSender

	//TODO
	//GetSubscriber() MessageSubscriber
}
