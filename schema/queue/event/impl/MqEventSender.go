package event

import (
	"log"
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"google.golang.org/protobuf/proto"
)

type CommonEventSender struct {
	ConnManager  *MQcommon.ConnectionManager
	exchangeName string
	sendQueue    string
	th2Pin       string
}

func (sender *CommonEventSender) Send(batch *p_buff.EventBatch) error {

	if batch == nil {
		log.Fatalln("Value for send can not be null")
	}
	body, err := proto.Marshal(batch)
	if err != nil {
		if err != nil {
			log.Panicf("Error during marshaling message into proto message. : %s", err)
		}
		return err
	}

	fail := sender.ConnManager.Publisher.Publish(body, sender.sendQueue, sender.exchangeName)
	if fail != nil {
		return fail
	}
	// th2Pin will be used for Metrics
	return nil
}
