package main

import (
	p_buff "github.com/th2-net/th2-common-go/proto"
	"github.com/th2-net/th2-common-go/queue/common"
	message "github.com/th2-net/th2-common-go/queue/messages"
	manager "github.com/th2-net/th2-common-go/queue/messages/connection"
	config "github.com/th2-net/th2-common-go/queue/messages/impl"
	"log"
)

type confirmationListener struct {
}

func (cl confirmationListener) Handle(delivery *common.Delivery, batch *p_buff.MessageGroupBatch, confirm *common.Confirmation) error {
	log.Println("Handling")
	log.Println(batch)
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

func (l listener) Handle(delivery *common.Delivery, batch *p_buff.MessageGroupBatch) error {
	log.Println("Handling")
	log.Println(batch)
	return nil

}

func (l listener) OnClose() error {
	log.Println("Listener OnClose")
	return nil
}

func main() {
	MqRouter := config.CommonMessageRouter{ConnManager: manager.ConnectionManager{}, Senders: make(map[string]config.CommonMessageSender), Subscribers: make(map[string]config.CommonMessageSubscriber)}
	MqRouter.ConnManager.Init("../resources/routermq.json", "../resources/rabbitmq.json")

	//err := MqRouter.SendAll(&p_buff.MessageGroupBatch{}, "group")
	//if err != nil {
	//	log.Fatalf("Cannt send, reason : ", err)
	//}

	l := listener{}
	var ml message.MessageListener = l
	//cml := confirmationListener{}
	//var c message.ConformationMessageListener = cml
	monitor, _ := MqRouter.SubscribeAll(&ml, "group")
	err := monitor.Unsubscribe()
	if err != nil {
		log.Println(err)
	}

}
