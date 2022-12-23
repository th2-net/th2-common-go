package main

import (
	"fmt"
	p_buff "github.com/th2-net/th2-common-go/proto"
	"github.com/th2-net/th2-common-go/schema/factory"
	rabbitmq "github.com/th2-net/th2-common-go/schema/modules"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/message"
	"log"
	"os"
	"reflect"
)

type confirmationListener struct {
}

func (cl confirmationListener) Handle(delivery *MQcommon.Delivery, batch *p_buff.MessageGroupBatch,
	confirm *MQcommon.Confirmation) error {
	log.Println("Handling")
	log.Println(batch)
	log.Printf("redelivered : %v \n", delivery)
	err := (*confirm).Confirm()
	if err != nil {
		log.Println(err)
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

func (l listener) Handle(delivery *MQcommon.Delivery, batch *p_buff.MessageGroupBatch) error {
	log.Println("Handling")
	log.Println(batch)
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
	messageRouter := module.MqMessageRouter
	// Send messages
	fail := messageRouter.SendAll(&p_buff.MessageGroupBatch{}, "group")
	if fail != nil {
		log.Fatalf("Cannt send, reason : ", fail)
	}

	//subscribe with listener
	l := listener{}
	var ml message.MessageListener = l
	monitor, err := messageRouter.SubscribeAll(&ml, "group")
	if err != nil {
		log.Println(err)
	}
	_ = monitor.Unsubscribe()
	module.Close()

	// subscribe with confirmation listener
	//cml := confirmationListener{}
	//var cl message.ConformationMessageListener = cml
	//monitor, err := messageRouter.SubscribeWithManualAck(&cl, "group")
	//if err != nil {
	//	log.Println(err)
	//}
	//_ = monitor.Unsubscribe()

}
