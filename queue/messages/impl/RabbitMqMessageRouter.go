package message

import (
	p_buff "github.com/th2-net/th2-common-go/proto"
	conn "github.com/th2-net/th2-common-go/queue/messages/connection"
	q_config "github.com/th2-net/th2-common-go/queue/queueConfiguration"
	"log"
	"strings"
)

type CommonMessageRouter struct {
	ConnManager conn.ConnectionManager

	Senders map[string]CommonMessageSender
}

func (cmr *CommonMessageRouter) SendAll(MsgBatch *p_buff.MessageGroupBatch, attributes ...string) error {
	attrs := cmr.getSendAttributes(attributes)
	aliasesFoundByAttrs := cmr.ConnManager.QConfig.FindQueuesByAttr(attrs)
	aliasAndMessageGroup := cmr.getMessageGroupWithAlias(aliasesFoundByAttrs, MsgBatch)
	if len(aliasAndMessageGroup) != 0 {
		for alias, messageGroup := range aliasAndMessageGroup {
			sender := cmr.getSender(alias)
			err := sender.Send(messageGroup)
			if err != nil {
				log.Fatalln(err)
				return err
			}
		}

	} else {
		log.Fatalln("no such queue to send message")
	}
	return nil

}

func (cmr *CommonMessageRouter) getSendAttributes(attrs []string) []string {
	for _, attr := range attrs {
		if strings.ToLower(attr) == "publish" {
			return attrs
		}
	}
	log.Fatalf("not publish attribute in list")
	return attrs[:0]

}

func (cmr *CommonMessageRouter) getSender(alias string) *CommonMessageSender {
	queueConfig := cmr.ConnManager.QConfig.Queues[alias] // get queue by alias
	log.Printf("queue by alias : %v \n", queueConfig)
	var result CommonMessageSender
	if _, ok := cmr.Senders[alias]; ok {
		result = cmr.Senders[alias]
		return &result
	} else {
		result = CommonMessageSender{ConnManager: cmr.ConnManager, exchangeName: queueConfig.Exchange,
			sendQueue: queueConfig.RoutingKey, th2Pin: alias}

		cmr.Senders[alias] = result
		log.Printf("cmr.senders[alias]s : %v \n", cmr.Senders[alias])

		return &result
	}
}

func (cmr *CommonMessageRouter) getMessageGroupWithAlias(queue map[string]q_config.QueueConfig, message *p_buff.MessageGroupBatch) map[string]*p_buff.MessageGroupBatch {
	//Here will be added filter handling
	result := make(map[string]*p_buff.MessageGroupBatch)
	for alias, _ := range queue {

		msgBatch := p_buff.MessageGroupBatch{}
		for _, messageGroup := range message.Groups {
			//doing filtering based on queue filters on message_group
			msgBatch.Groups = append(msgBatch.Groups, messageGroup)
		}

		result[alias] = &msgBatch
	}
	return result

}

//
//const (
//	FIRST     string = "first"
//	SECOND           = "second"
//	SUBSCRIBE        = "subscribe"
//	PUBLISH          = "publish"
//	RAW              = "raw"
//	PARSED           = "parsed"
//	STORE            = "store"
//	EVENT            = "event"
//)
