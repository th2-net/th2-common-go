package message

import (
	p_buff "github.com/th2-net/th2-common-go/proto"
	c "github.com/th2-net/th2-common-go/queue/common"
	message "github.com/th2-net/th2-common-go/queue/messages"
	conn "github.com/th2-net/th2-common-go/queue/messages/connection"
	q_config "github.com/th2-net/th2-common-go/queue/queueConfiguration"
	"log"
	"strings"
)

type CommonMessageRouter struct {
	ConnManager conn.ConnectionManager
	Subscribers map[string]CommonMessageSubscriber

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

func (cmr *CommonMessageRouter) SubscribeWithManualAck(listener *message.ConformationMessageListener, attributes ...string) (c.Monitor, error) {
	attrs := cmr.getSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	defer cmr.ConnManager.CloseConn()
	aliasesFoundByAttrs := cmr.ConnManager.QConfig.FindQueuesByAttr(attrs)
	for queueAlias, _ := range aliasesFoundByAttrs {
		log.Println(queueAlias)
		subscriber, err := cmr.subByAliasWithAck(listener, queueAlias)
		if err != nil {
			log.Fatalln(err)
			return SubscriberMonitor{}, err
		}
		subscribers = append(subscribers, subscriber)
	}
	if len(subscribers) != 0 {

		return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
	} else {
		log.Fatalln("no subscriber ")
	}
	return SubscriberMonitor{}, nil
}

func (cmr *CommonMessageRouter) SubscribeAll(listener *message.MessageListener, attributes ...string) (c.Monitor, error) {
	attrs := cmr.getSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	defer cmr.ConnManager.CloseConn()
	aliasesFoundByAttrs := cmr.ConnManager.QConfig.FindQueuesByAttr(attrs)
	for queueAlias, _ := range aliasesFoundByAttrs {
		log.Println(queueAlias)
		subscriber, err := cmr.subByAlias(listener, queueAlias)
		if err != nil {
			log.Fatalln(err)
			return SubscriberMonitor{}, err
		}
		subscribers = append(subscribers, subscriber)
	}
	if len(subscribers) != 0 {

		return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
	} else {
		log.Fatalln("no subscriber ")
	}
	return SubscriberMonitor{}, nil
}

//// //////////////////////////change
//func (cmr *CommonMessageRouter) getAttributes(attrs []string, keyAttr string) []string {
//	for _, attr := range attrs {
//		if strings.ToLower(attr) == keyAttr {
//			return attrs
//		}
//	}
//	log.Fatalf("not appropriate attribute in list")
//	return attrs[:0]
//
//}

func (cmr *CommonMessageRouter) getSendAttributes(attrs []string) []string {
	res := []string{}
	if len(attrs) == 0 {
		return res
	} else {
		attrMap := make(map[string]bool)
		for _, attr := range attrs {
			attrMap[strings.ToLower(attr)] = true
		}
		attrMap["publish"] = true
		for k, _ := range attrMap {
			res = append(res, k)
		}
		return res
	}
}

func (cmr *CommonMessageRouter) getSubscribeAttributes(attrs []string) []string {
	res := []string{}
	if len(attrs) == 0 {
		return res
	} else {
		attrMap := make(map[string]bool)
		for _, attr := range attrs {
			attrMap[strings.ToLower(attr)] = true
		}
		attrMap["subscribe"] = true
		for k, _ := range attrMap {
			res = append(res, k)
		}
		return res
	}
}

func (cmr *CommonMessageRouter) subByAlias(listener *message.MessageListener, alias string) (SubscriberMonitor, error) {
	subscriber := cmr.getSubscriber(alias)
	subscriber.AddListener(listener)
	err := subscriber.Start()
	if err != nil {
		return SubscriberMonitor{}, err
	}
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) subByAliasWithAck(listener *message.ConformationMessageListener, alias string) (SubscriberMonitor, error) {
	subscriber := cmr.getSubscriber(alias)
	subscriber.AddConfListener(listener)
	err := subscriber.StartConf()
	if err != nil {
		return SubscriberMonitor{}, err
	}
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cmr *CommonMessageRouter) getSubscriber(alias string) *CommonMessageSubscriber {
	queueConfig := cmr.ConnManager.QConfig.Queues[alias] // get queue by alias
	var result CommonMessageSubscriber
	if _, ok := cmr.Subscribers[alias]; ok {
		result = cmr.Subscribers[alias]
		return &result
	} else {
		result = CommonMessageSubscriber{ConnManager: cmr.ConnManager, qConfig: queueConfig, th2Pin: alias}

		cmr.Subscribers[alias] = result
		log.Printf("cmr.subs[alias]s : %v \n", alias)

		return &result
	}
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
		log.Printf("cmr.senders[alias] : %v \n", cmr.Senders[alias])

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
