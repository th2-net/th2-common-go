package message

import (
	p_buff "github.com/th2-net/th2-common-go/proto"
)

type MessageRouter interface {
	SendAll(batch *p_buff.MessageGroupBatch, attributes ...string) error
	//SubscribeAll(listener *MessageListener, attributes ...string) (q.Monitor, error)
	//SubscribeWithManualAck(listener *ConformationMessageListener, attributes ...string) (q.Monitor, error)
}
