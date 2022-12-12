package message

import (
	p_buff "github.com/th2-net/th2-common-go/proto"
	"github.com/th2-net/th2-common-go/queue/common"
)

type MessageListener interface {
	common.CloseListener
	Handle(delivery *common.Delivery, batch *p_buff.MessageGroupBatch) error
}

type ConformationMessageListener interface {
	common.CloseListener
	Handle(delivery *common.Delivery, batch *p_buff.MessageGroupBatch, confirm *common.Confirmation) error
}
