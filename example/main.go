package main

import (
	"encoding/json"
	"fmt"
	"github.com/th2-net/th2-common-go/example/MessageConverter"
	"github.com/th2-net/th2-common-go/schema/queue/message"

	//conversion "github.com/th2-net/th2-common-go/example/MessageConverter"
	p_buff "github.com/th2-net/th2-common-go/proto"
	"github.com/th2-net/th2-common-go/schema/factory"

	rabbitmq "github.com/th2-net/th2-common-go/schema/modules"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"go.uber.org/zap"
	"log"
	"os"
	"reflect"
)

type confirmationListener struct {
}

func (cl confirmationListener) Handle(delivery *MQcommon.Delivery, batch *p_buff.MessageGroupBatch,
	confirm *MQcommon.Confirmation) error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	log.Println("in confirmation Handle function with batch")

	err := (*confirm).Confirm()
	log.Printf("type of pabtch is %T \n", batch)

	if err != nil {
		sugar.Errorf(" errors in concrimation %v \n", err)
		log.Fatalln(err)
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
	log.Println("in Handle function with batch")
	log.Printf(" type of batchi is : %T \n", batch)
	log.Println("handled")
	return nil

}

func (l listener) OnClose() error {
	log.Println("Listener OnClose")
	return nil
}

func CreateMessage(path string) *p_buff.Message {
	message := MessageConverter.MessageStruct{} // Create empty Message struct(Metadata, Fields, ParentEventId)
	content, err := os.ReadFile(path)           // Read json file
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	fail := json.Unmarshal(content, &message)
	if fail != nil {
		log.Fatal("Error during Unmarshal(): ", fail)
	}
	result := MessageConverter.MessageFromStruct(&message)
	return result
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
	log.Printf("%T", messageRouter)

	///create message group batch

	////message2 := CreateMessage("Messages/message3.json")
	//message3 := CreateMessage("Messages/message13.json")
	//messageBatch := p_buff.MessageBatch{Messages: []*p_buff.Message{message3}}
	//groupBatch := conversion.ToGroupBatch(&messageBatch)
	//
	//////Send messages
	//fail := messageRouter.SendAll(groupBatch, "group")
	//if fail != nil {
	//	log.Fatalf("Cannt send, reason : ", fail)
	//}

	//////subscribe with listener
	//l := listener{}
	//var ml message.MessageListener = l
	//monitor, err := messageRouter.SubscribeAll(&ml, "group")
	//if err != nil {
	//	log.Println(err)
	//}
	//
	////_ = monitor.Unsubscribe()
	//module.Close()

	//subscribe with confirmation listener
	cml := confirmationListener{}
	var cl message.ConformationMessageListener = cml
	monitor, err := messageRouter.SubscribeWithManualAck(&cl, "group")
	if err != nil {
		log.Println(err)
	}
	log.Println(monitor)
	//
	_ = monitor.Unsubscribe()
	module.Close()

}
