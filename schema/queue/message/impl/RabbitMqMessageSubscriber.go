package message

import (
	conn "github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/message"
	"github.com/th2-net/th2-common-go/schema/queue/message/configuration"
)

type CommonMessageSubscriber struct {
	ConnManager          conn.ConnectionManager
	qConfig              configuration.QueueConfig
	listener             *message.MessageListener
	confirmationListener *message.ConformationMessageListener
	th2Pin               string
}

func (cs *CommonMessageSubscriber) Start() error {
	err := cs.ConnManager.MessageGroupConsume(cs.qConfig.QueueName, cs.listener)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}
func (cs *CommonMessageSubscriber) StartWithConf() error {
	err := cs.ConnManager.MessageGroupConsumeManualAck(cs.qConfig.QueueName, cs.confirmationListener)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonMessageSubscriber) AddListener(listener *message.MessageListener) {
	cs.listener = listener
}

func (cs *CommonMessageSubscriber) AddConfListener(listener *message.ConformationMessageListener) {
	cs.confirmationListener = listener
}

type SubscriberMonitor struct {
	subscriber *CommonMessageSubscriber
}

func (sub SubscriberMonitor) Unsubscribe() error {
	//////////////////// Need to held
	err := (*sub.subscriber.listener).OnClose()
	if err != nil {
		return err
	}
	//fail := (*sub.subscriber.confirmationListener).OnClose()
	//if fail != nil {
	//	return fail
	//}
	return nil
}

type MultiplySubscribeMonitor struct {
	subscriberMonitors []SubscriberMonitor
}

func (sub MultiplySubscribeMonitor) Unsubscribe() error {
	for _, subM := range sub.subscriberMonitors {
		err := subM.Unsubscribe()
		if err != nil {
			return err
		}
	}
	return nil
}
