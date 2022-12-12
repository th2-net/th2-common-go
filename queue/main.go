package main

import (
	p_buff "github.com/th2-net/th2-common-go/proto"
	manager "github.com/th2-net/th2-common-go/queue/messages/connection"
	config "github.com/th2-net/th2-common-go/queue/messages/impl"
	"log"
)

func main() {
	MqRouter := config.CommonMessageRouter{ConnManager: manager.ConnectionManager{}, Senders: make(map[string]config.CommonMessageSender)}
	MqRouter.ConnManager.Init("../resources/routermq.json", "../resources/rabbitmq.json")

	err := MqRouter.SendAll(&p_buff.MessageGroupBatch{}, "publish", "raw")
	if err != nil {
		log.Fatalf("Cannt send, reason : ", err)
	}
}
