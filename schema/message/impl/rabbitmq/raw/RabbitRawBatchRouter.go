package th2

import (
	"fmt"

	"github.com/rs/zerolog"

	p_buff "github.com/th2-net/th2-common-go/proto"
	message "github.com/th2-net/th2-common-go/schema/message"
	cfg "github.com/th2-net/th2-common-go/schema/message/configuration"
	rmq_connection "github.com/th2-net/th2-common-go/schema/message/impl/rabbitmq/connection"
)

type AliasAndMessageToSend = map[string]*p_buff.RawMessageBatch

// MessageRouter interface
type RabbitRawBatchRouter struct {
	connectionManager          *rmq_connection.ConnectionManager
	messageRouterConfiguration *cfg.MessageRouterConfiguration

	Logger zerolog.Logger

	//TODO QueueAttribute
	REQUIRED_SUBSCRIBE_ATTRIBUTES cfg.QueueAttribute
	REQUIRED_SEND_ATTRIBUTES      cfg.QueueAttribute
}

func (br *RabbitRawBatchRouter) Init(conn *rmq_connection.ConnectionManager, mr_cfg *cfg.MessageRouterConfiguration) error {

	//TODO get from QueueConfiguration
	br.REQUIRED_SUBSCRIBE_ATTRIBUTES = []string{"raw", "subscribe"}
	br.REQUIRED_SEND_ATTRIBUTES = []string{"raw", "publish"}

	if conn != nil {
		br.connectionManager = conn
	} else {
		err_str := "[RabbitRawBatchRouter] Connection owner can not be null"
		br.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	if mr_cfg != nil {
		br.messageRouterConfiguration = mr_cfg
	} else {
		err_str := "[RabbitRawBatchRouter] Configuration cannot be null"
		br.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	return nil
}

func (br *RabbitRawBatchRouter) Send(msg interface{}) error {

	if msg == nil {
		err_str := "[RabbitRawBatchRouter] Batch cannot be nil"
		br.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	batch, ok := msg.(*p_buff.RawMessageBatch)
	if !ok {
		err_str := "[RabbitRawBatchRouter] Incoming message should be an RawMessageBatch type"
		br.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	required_send_attributes := br.requiredSendAttributes()

	var queuesMap *map[string]cfg.QueueConfiguration

	if len(required_send_attributes) > 0 {
		queuesMap = br.messageRouterConfiguration.FindQueuesByAttr(required_send_attributes)
	} else {
		queuesMap = br.messageRouterConfiguration.GetQueues()
	}

	filteredByAttrAndFilter := br.findByFilter(queuesMap, batch)

	if len(filteredByAttrAndFilter) != 1 {

		err_str := "[RabbitRawBatchRouter] Wrong count of queues for send. Should be equal to 1. Find queues = " + fmt.Sprint(len(filteredByAttrAndFilter))
		br.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	br.send(filteredByAttrAndFilter)

	return nil
}

func (br *RabbitRawBatchRouter) SendByQueueAttributes(msg interface{}, attr cfg.QueueAttribute) error {

	if msg == nil {
		return fmt.Errorf("[RabbitRawBatchRouter] Batch cannot be nil")
	}

	batch, ok := msg.(*p_buff.RawMessageBatch)
	if !ok {
		err_str := "[RabbitRawBatchRouter] Incoming message should be an RawMessageBatch type"
		br.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	filteredByAttr := br.messageRouterConfiguration.FindQueuesByAttr(attr)

	filteredByAttrAndFilter := br.findByFilter(filteredByAttr, batch)

	if len(filteredByAttrAndFilter) != 1 {
		err_str := "[RabbitRawBatchRouter] Wrong size of queues for send. Should be equal to 1"
		br.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	br.send(filteredByAttrAndFilter)

	return nil
}

func (br *RabbitRawBatchRouter) SendAll(msg interface{}, attr cfg.QueueAttribute) error {

	if msg == nil {
		err_str := "[RabbitRawBatchRouter] Batch cannot be nil"
		br.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	batch, ok := msg.(*p_buff.RawMessageBatch)
	if !ok {
		err_str := "[RabbitRawBatchRouter] Incoming message should be an RawMessageBatch type"
		br.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	filteredByAttr := br.messageRouterConfiguration.FindQueuesByAttr(attr)

	filteredByAttrAndFilter := br.findByFilter(filteredByAttr, batch)

	if len(filteredByAttrAndFilter) < 1 {
		err_str := "[RabbitRawBatchRouter] Wrong size of queues for send. Can't be equal to 0"
		br.Logger.Error().Msg(err_str)
		return fmt.Errorf(err_str)
	}

	br.send(filteredByAttrAndFilter)

	return nil
}

func (br *RabbitRawBatchRouter) Subscribe() error {

	err_str := "[RabbitRawBatchRouter] Method Subscribe() not implemented yet"
	br.Logger.Error().Msg(err_str)
	defer panic(err_str)

	return nil
}

func (br *RabbitRawBatchRouter) SubscribeAll() error {

	err_str := "[RabbitRawBatchRouter] Method SubscribeAll() not implemented yet"
	br.Logger.Error().Msg(err_str)
	defer panic(err_str)
	return nil
}

func (br *RabbitRawBatchRouter) findByFilter(queues_map *cfg.QueuesMap, batch *p_buff.RawMessageBatch) AliasAndMessageToSend {

	res := AliasAndMessageToSend{}

	for _, msg := range batch.Messages {

		for queue_alias := range br.filter(*queues_map, msg) {

			a := res[queue_alias]
			if a == nil {

				res[queue_alias] = &p_buff.RawMessageBatch{}
			}

			res[queue_alias].Messages = append(res[queue_alias].Messages, msg)
		}

	}
	return res
}

func (br *RabbitRawBatchRouter) send(msg_map AliasAndMessageToSend) error {

	var err error
	for alias, msg := range msg_map {

		sender := (*br.getMessageQueue(alias)).GetSender()
		err = (*sender).Send(msg)

		if err != nil {
			err_str := "[RabbitRawBatchRouter] unable to send message: " + err.Error()
			br.Logger.Error().Msg(err_str)
		}
	}
	return err
}

func (br *RabbitRawBatchRouter) filter(queues cfg.QueuesMap, msg *p_buff.RawMessage) map[string]bool {

	//TODO Utils module
	//Idiomatic SET
	aliases := map[string]bool{}

	for alias, queue := range queues {

		filters := queue.GetFilters()
		//TODO filterStrategy
		if len(*filters) == 0 {
			aliases[alias] = true
		}
	}
	return aliases
}

func (br *RabbitRawBatchRouter) getMessageQueue(queueAlias string) *message.MessageQueue {

	queueCfg, _ := br.messageRouterConfiguration.GetQueueByAlias(queueAlias) //TODO: do not ignore err
	queue := br.createQueue(queueCfg)

	return queue
}

func (br *RabbitRawBatchRouter) requiredSendAttributes() cfg.QueueAttribute {
	return br.REQUIRED_SEND_ATTRIBUTES
}

func (br *RabbitRawBatchRouter) requiredSubscribeAttributes() cfg.QueueAttribute {
	return br.REQUIRED_SUBSCRIBE_ATTRIBUTES
}

func (br *RabbitRawBatchRouter) createQueue(queueCfg *cfg.QueueConfiguration) *message.MessageQueue {

	var queue message.MessageQueue = &RabbitRawBatchQueue{Logger: br.Logger}
	queue.Init(br.connectionManager, queueCfg)

	return &queue
}
