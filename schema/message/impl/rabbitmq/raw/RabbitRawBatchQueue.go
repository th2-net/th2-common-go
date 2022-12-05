package th2

import (
	"fmt"

	"github.com/rs/zerolog"

	message "exactpro/th2/th2-common-go/schema/message"
	cfg "exactpro/th2/th2-common-go/schema/message/configuration"
	connection "exactpro/th2/th2-common-go/schema/message/impl/rabbitmq/connection"
)

type RabbitRawBatchQueue struct {
	connectionManager  *connection.ConnectionManager
	queueConfiguration *cfg.QueueConfiguration

	sender *message.MessageSender

	Logger zerolog.Logger

	//TODO
	//messageSubscriber *message.MessageSubscriber
}

func (q *RabbitRawBatchQueue) Init(conn *connection.ConnectionManager, cfg *cfg.QueueConfiguration) error {

	if conn == nil {
		err_str := "[RabbitRawBatchQueue] Connection can not be null"
		q.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	if q.connectionManager != nil {
		err_str := "[RabbitRawBatchQueue] Queue is already initialized"
		q.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	q.connectionManager = conn
	q.queueConfiguration = cfg

	return nil
}

func (q *RabbitRawBatchQueue) GetSender() *message.MessageSender {

	if q.connectionManager == nil || q.queueConfiguration == nil {
		q.Logger.Error().Msg("[RabbitRawBatchQueue] Queue is not initialized")
		return nil
	}

	if !q.queueConfiguration.IsWritable() {
		q.Logger.Error().Msg("[RabbitRawBatchQueue] Queue can not write")
		return nil
	}

	if q.sender == nil {
		q.sender = q.createSender(q.connectionManager, q.queueConfiguration)
	}

	return q.sender
}

func (q *RabbitRawBatchQueue) createSender(conn *connection.ConnectionManager, cfg *cfg.QueueConfiguration) *message.MessageSender {

	var sender message.MessageSender = &RabbitRawBatchSender{Logger: q.Logger}
	sender.Init(conn, cfg.GetExchange(), cfg.GetRoutingKey())

	return &sender
}
