package main

import (
	"fmt"
	"github.com/th2-net/th2-common-go/schema/queue/event"
	//conversion "github.com/th2-net/th2-common-go/example/MessageConverter"
	p_buff "github.com/th2-net/th2-common-go/proto"
	"github.com/th2-net/th2-common-go/schema/factory"

	rabbitmq "github.com/th2-net/th2-common-go/schema/modules"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"log"
	"os"
	"reflect"
)

type confirmationListener struct {
}

func (cl confirmationListener) Handle(delivery *MQcommon.Delivery, batch *p_buff.EventBatch,
	confirm *MQcommon.Confirmation) error {
	log.Println("in confirmation Handle function with batch")
	err := (*confirm).Confirm()
	log.Printf("type of pabtch is %T \n", batch)
	if err != nil {
		log.Fatalf("errors in concrimation %v \n", err)
		return err
	}
	return nil
}

func (cl confirmationListener) OnClose() error {
	log.Println("ConfirmationListener OnClose")
	return nil
}

type listener struct {
}

func (l listener) Handle(delivery *MQcommon.Delivery, batch *p_buff.EventBatch) error {
	log.Println("in Handle function with batch")
	log.Printf(" type of batch is : %T \n", batch)
	log.Printf(" %v handled \n", delivery)
	return nil

}

func (l listener) OnClose() error {
	log.Println("Listener OnClose")
	return nil
}

func main() {
	newFactory := factory.NewFactory(os.Args)
	if err := newFactory.Register(rabbitmq.NewRabbitMQModule); err != nil {
		panic(err)
	}

	module, err := rabbitmq.ModuleID.GetModule(newFactory)
	if err != nil {
		panic("no module")
	} else {
		fmt.Println("module found", reflect.TypeOf(module))
	}

	eventRouter := module.MqEventRouter

	l := confirmationListener{}
	var ml event.ConformationEventListener = l
	monitor, err := eventRouter.SubscribeWithManualAck(&ml, "group")
	if err != nil {
		log.Println(err)
	}

	_ = monitor.Unsubscribe()
	module.Close()

	//fail := eventRouter.SendAll(&p_buff.EventBatch{}, "group")
	//if fail != nil {
	//	log.Fatalf("Cannt send, reason : ", fail)
	//}
	//module.Close()

}
