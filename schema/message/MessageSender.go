package th2

import (
	rmq_connection "exactpro/th2/th2-common-go/schema/message/impl/rabbitmq/connection"
)

type MessageSender interface {
	Init(*rmq_connection.ConnectionManager, string, string) error

	Send(interface{}) error
}
