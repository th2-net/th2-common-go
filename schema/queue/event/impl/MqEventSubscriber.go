package event

import (
	"log"
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/streadway/amqp"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/configuration"
	"github.com/th2-net/th2-common-go/schema/queue/event"
	"google.golang.org/protobuf/proto"
)

type CommonEventSubscriber struct {
	connManager          *MQcommon.ConnectionManager
	qConfig              *configuration.QueueConfig
	listener             *event.EventListener
	confirmationListener *event.ConformationEventListener
	th2Pin               string
}

func (cs *CommonEventSubscriber) Start() error {
	err := cs.connManager.Consumer.Consume(cs.qConfig.QueueName, cs.Handler)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonEventSubscriber) ConfirmationStart() error {
	err := cs.connManager.Consumer.ConsumeWithManualAck(cs.qConfig.QueueName, cs.ConfirmationHandler)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonEventSubscriber) Handler(msgDelivery amqp.Delivery) {
	result := &p_buff.EventBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		log.Fatalf("Cann't unmarshal : %v \n", err)
	}
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	if cs.listener == nil {
		log.Fatalf("No Listener to Handle : %v \n", cs.listener)

	}
	fail := (*cs.listener).Handle(&delivery, result)
	if fail != nil {
		log.Fatalf("Cann't Handle : %v \n", fail)
	}
}

func (cs *CommonEventSubscriber) ConfirmationHandler(msgDelivery amqp.Delivery) {
	result := &p_buff.EventBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		log.Fatalf("Cann't unmarshal : %v \n", err)
	}

	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	deliveryConfirm := MQcommon.DeliveryConfirmation{Delivery: &msgDelivery}
	var confirmation MQcommon.Confirmation = deliveryConfirm

	if cs.confirmationListener == nil {
		log.Fatalf("No ConfirmationListener to Handle : %v \n", cs.confirmationListener)

	}
	fail := (*cs.confirmationListener).Handle(&delivery, result, &confirmation)
	if fail != nil {
		log.Fatalf("Cann't Handle : %v \n", fail)
	}
}

func (cs *CommonEventSubscriber) RemoveListener() {
	cs.listener = nil
	cs.confirmationListener = nil
	log.Println("Removing listeners ******** ")
}

func (cs *CommonEventSubscriber) AddListener(listener *event.EventListener) {
	cs.listener = listener
}

type SubscriberMonitor struct {
	subscriber *CommonEventSubscriber
}

func (cs *CommonEventSubscriber) AddConfirmationListener(listener *event.ConformationEventListener) {
	cs.confirmationListener = listener
}

func (sub SubscriberMonitor) Unsubscribe() error {
	if sub.subscriber.listener != nil {
		err := (*sub.subscriber.listener).OnClose()
		if err != nil {
			return err
		}
		sub.subscriber.RemoveListener()
	}

	if sub.subscriber.confirmationListener != nil {
		err := (*sub.subscriber.confirmationListener).OnClose()
		if err != nil {
			return err
		}
		sub.subscriber.RemoveListener()
	}
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
