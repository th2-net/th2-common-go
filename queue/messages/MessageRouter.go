package message

import (
	p_buff "github.com/th2-net/th2-common-go/proto"
	c "github.com/th2-net/th2-common-go/queue/common"
)

type MessageRouter interface {
	SendAll(batch *p_buff.MessageGroupBatch, attributes ...string) error
	SubscribeAll(listener *MessageListener, attributes ...string) (c.Monitor, error)
	SubscribeAllWithManualAck(listener *ConformationMessageListener, attributes ...string) (c.Monitor, error)
}
