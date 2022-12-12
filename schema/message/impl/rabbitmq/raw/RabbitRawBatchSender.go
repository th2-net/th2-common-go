package th2

import (
	"fmt"

	"github.com/rs/zerolog"

	"google.golang.org/protobuf/proto"

	p_buff "github.com/th2-net/th2-common-go/proto"
	rmq_connection "github.com/th2-net/th2-common-go/schema/message/impl/rabbitmq/connection"
)

type RabbitRawBatchSender struct {
	sendQueue         string
	exchangeName      string
	connectionManager *rmq_connection.ConnectionManager

	Logger zerolog.Logger
}

func (s *RabbitRawBatchSender) Init(conn *rmq_connection.ConnectionManager, exchange string, sendQueue string) error {

	s.connectionManager = conn
	s.exchangeName = exchange
	s.sendQueue = sendQueue

	return nil
}

func (s *RabbitRawBatchSender) Send(msg interface{}) error {

	if msg == nil {
		err_str := "[RabbitRawBatchSender] Batch cannot be nil"
		s.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	batch, ok := msg.(*p_buff.RawMessageBatch)
	if !ok {
		err_str := "[RabbitRawBatchSender] Incoming message should be an RawMessageBatch type"
		s.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	msgByte := s.valueToBytes(batch)
	err := s.connectionManager.BasicPublish(s.exchangeName, s.sendQueue, msgByte)

	return err
}

func (s *RabbitRawBatchSender) valueToBytes(batch *p_buff.RawMessageBatch) []byte {

	out_bytes, err := proto.Marshal(batch)

	if err != nil {
		s.Logger.Error().Msg("[RabbitRawBatchSender] Marshaling error")
		return nil
	}
	return out_bytes
}
