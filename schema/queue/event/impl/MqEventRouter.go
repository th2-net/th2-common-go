package event

import (
	p_buff "github.com/th2-net/th2-common-go/proto"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/event"
	"log"
	"sync"
)

type CommonEventRouter struct {
	connManager *MQcommon.ConnectionManager
	subscribers map[string]CommonEventSubscriber
	senders     map[string]CommonEventSender
}

func (cer *CommonEventRouter) Construct(manager *MQcommon.ConnectionManager) {
	cer.connManager = manager
	cer.subscribers = map[string]CommonEventSubscriber{}
	cer.senders = map[string]CommonEventSender{}
}

func (cer *CommonEventRouter) Close() {
	_ = cer.connManager.Close()
}

func (cer *CommonEventRouter) SendAll(EventBatch *p_buff.EventBatch, attributes ...string) error {
	attrs := MQcommon.GetSendAttributes(attributes)
	pinsFoundByAttrs := cer.connManager.QConfig.FindQueuesByAttr(attrs)
	if len(pinsFoundByAttrs) != 0 {
		for pin, _ := range pinsFoundByAttrs {
			sender := cer.getSender(pin)
			err := sender.Send(EventBatch)
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

func (cer *CommonEventRouter) SubscribeAll(listener *event.EventListener, attributes ...string) (MQcommon.Monitor, error) {
	attrs := MQcommon.GetSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	pinsFoundByAttrs := cer.connManager.QConfig.FindQueuesByAttr(attrs)
	for queuePin, _ := range pinsFoundByAttrs {
		log.Printf("Subscrubing %v \n", queuePin)
		subscriber, err := cer.subByPin(listener, queuePin)
		if err != nil {
			log.Fatalln(err)
			return nil, err
		}
		subscribers = append(subscribers, subscriber)
	}
	var m sync.Mutex
	if len(subscribers) != 0 {
		for _, s := range subscribers {
			m.Lock()
			err := s.subscriber.Start()
			if err != nil {
				log.Printf("CONSUMING ERROR : %v \n", err)
				return SubscriberMonitor{}, err
			}
			m.Unlock()
		}
		return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
	} else {
		log.Fatalln("no subscriber ")
	}
	return nil, nil
}

func (cer *CommonEventRouter) SubscribeWithManualAck(listener *event.ConformationEventListener, attributes ...string) (MQcommon.Monitor, error) {
	attrs := MQcommon.GetSubscribeAttributes(attributes)
	subscribers := []SubscriberMonitor{}
	pinFoundByAttrs := cer.connManager.QConfig.FindQueuesByAttr(attrs)
	for queuePin, _ := range pinFoundByAttrs {
		log.Printf("Subscrubing %v \n", queuePin)
		subscriber, err := cer.subByPinWithAck(listener, queuePin)
		if err != nil {
			log.Fatalln(err)
			return SubscriberMonitor{}, err
		}
		subscribers = append(subscribers, subscriber)
	}
	var m sync.Mutex

	if len(subscribers) != 0 {
		for _, s := range subscribers {
			m.Lock()
			err := s.subscriber.ConfirmationStart()
			if err != nil {
				log.Printf("CONSUMING ERROR : %v \n", err)
				return SubscriberMonitor{}, err
			}
			m.Unlock()
		}

		return MultiplySubscribeMonitor{subscriberMonitors: subscribers}, nil
	} else {
		log.Fatalln("no subscriber ")
	}
	return SubscriberMonitor{}, nil
}

func (cer *CommonEventRouter) subByPin(listener *event.EventListener, pin string) (SubscriberMonitor, error) {
	subscriber := cer.getSubscriber(pin)
	subscriber.AddListener(listener)
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cer *CommonEventRouter) subByPinWithAck(listener *event.ConformationEventListener, alias string) (SubscriberMonitor, error) {
	subscriber := cer.getSubscriber(alias)
	subscriber.AddConfirmationListener(listener)
	return SubscriberMonitor{subscriber: subscriber}, nil
}

func (cer *CommonEventRouter) getSubscriber(pin string) *CommonEventSubscriber {
	queueConfig := cer.connManager.QConfig.Queues[pin] // get queue by pin
	var result CommonEventSubscriber
	if _, ok := cer.subscribers[pin]; ok {
		result = cer.subscribers[pin]
		return &result
	} else {
		result = CommonEventSubscriber{connManager: cer.connManager, qConfig: &queueConfig,
			listener: nil, confirmationListener: nil, th2Pin: pin}
		cer.subscribers[pin] = result
		return &result
	}
}

func (cer *CommonEventRouter) getSender(pin string) *CommonEventSender {
	queueConfig := cer.connManager.QConfig.Queues[pin] // get queue by pin
	var result CommonEventSender
	if _, ok := cer.senders[pin]; ok {
		result = cer.senders[pin]
		return &result
	} else {
		result = CommonEventSender{ConnManager: cer.connManager, exchangeName: queueConfig.Exchange,
			sendQueue: queueConfig.RoutingKey, th2Pin: pin}
		cer.senders[pin] = result

		return &result
	}
}
