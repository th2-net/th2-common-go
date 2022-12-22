package main

import (
	"fmt"
	p_buff "github.com/th2-net/th2-common-go/proto"
	fc "github.com/th2-net/th2-common-go/schema/factory"
	rabbitmq "github.com/th2-net/th2-common-go/schema/modules"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"log"
	"os"
	"reflect"
)

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
	factory := fc.NewFactory(os.Args)
	if err := factory.Register(rabbitmq.NewRabbitMQModule); err != nil {
		panic(err)
	}

	module, err := rabbitmq.ModuleID.GetModule(factory)
	if err != nil {
		panic("no module")
	} else {
		fmt.Println("module found", reflect.TypeOf(module))
	}
	messageRouter := module.MqMessageRouter
	log.Printf("message router %v \n ", messageRouter)

	fail := messageRouter.SendAll(&p_buff.MessageGroupBatch{}, "group")
	if fail != nil {
		log.Fatalf("Cannt send, reason : ", fail)
	}
	//l := listener{}
	//var ml message.MessageListener = l
	////cml := confirmationListener{}
	////var c message.ConformationMessageListener = cml
	//monitor, _ := MqRouter.SubscribeAll(&ml, "group")
	//err := monitor.Unsubscribe()
	//if err != nil {
	//	log.Println(err)
	//}
}
