package message

import (
	p_buff "github.com/th2-net/th2-common-go/proto"
	conn "github.com/th2-net/th2-common-go/queue/messages/connection"
	"log"
)

type CommonMessageSender struct {
	ConnManager  conn.ConnectionManager
	exchangeName string
	sendQueue    string
	th2Pin       string
}

func (sender *CommonMessageSender) Send(batch *p_buff.MessageGroupBatch) error {

	if batch == nil {
		log.Fatalln("Value for send can not be null")
	}
	err := sender.ConnManager.BasicPublish(batch, sender.sendQueue, sender.exchangeName)
	if err != nil {
		return err
	}
	// th2Pin will be used for Metrics
	return nil
}

//func (*CommonMessageSender) Start() {
//
//}
