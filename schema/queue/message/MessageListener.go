package message

import (
	p_buff "github.com/th2-net/th2-common-go/proto"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
)

type MessageListener interface {
	MQcommon.CloseListener
	Handle(delivery *MQcommon.Delivery, batch *p_buff.MessageGroupBatch) error
}

type ConformationMessageListener interface {
	MQcommon.CloseListener
	Handle(delivery *MQcommon.Delivery, batch *p_buff.MessageGroupBatch, confirm *MQcommon.Confirmation) error
}
