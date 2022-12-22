package MQcommon

import (
	"github.com/streadway/amqp"
	p_buff "github.com/th2-net/th2-common-go/proto"
	"github.com/th2-net/th2-common-go/schema/queue/message"
	"github.com/th2-net/th2-common-go/schema/queue/message/configuration"
	"github.com/wagslane/go-rabbitmq"
	"google.golang.org/protobuf/proto"
	"log"
)

type ConnectionManager struct {
	QConfig      *configuration.MessageRouterConfiguration
	MqConnConfig *configuration.RabbitMQConfiguration
	Url          string
	conn         *amqp.Connection
	channel      *amqp.Channel
}

func (manager *ConnectionManager) Connect() {
	conn, err := amqp.Dial(manager.Url)
	if err != nil {
		log.Fatalln(err)
	}
	manager.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalln(err)
	}
	manager.channel = ch
}

func (manager *ConnectionManager) CloseConn() error {
	err := manager.channel.Close()
	if err != nil {
		return err
	}
	fail := manager.conn.Close()
	if fail != nil {
		return err
	}
	log.Println("Closed gracefully")
	return nil
}

type DeliveryConfirmation struct {
	delivery *amqp.Delivery
}

func (dc DeliveryConfirmation) Confirm() error {
	err := dc.delivery.Ack(false)
	if err != nil {
		return err
	}
	return nil
}
func (dc DeliveryConfirmation) Reject() error {
	err := dc.delivery.Reject(false)
	if err != nil {
		return err
	}
	return nil
}

///////// EventBatch Methods /////
//func (manager ConnectionManager) EventBatchPublish(event *p_buff.EventBatch, routingKey string, exchange string) error {}
//func (manager *ConnectionManager) EventBatchConsumeManualAck(queueName string, listener *message.ConformationEventListener) error {}
//func (manager *ConnectionManager) EventBatchConsume(queueName string, listener *message.EventListener) error {}

// // MessageGroupBatch methods ////////////////
func (manager ConnectionManager) MessageGroupPublish(message *p_buff.MessageGroupBatch, routingKey string, exchange string) error {
	conn, err := rabbitmq.NewConn(
		manager.Url,
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer conn.Close()

	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName(exchange),
	)
	if err != nil {
		log.Fatal(err)
		return err

	}
	defer publisher.Close()

	body, err := proto.Marshal(message)
	if err != nil {
		if err != nil {
			log.Panicf("Error during marshaling message into proto message. : %s", err)
		}
		return err
	}

	err = publisher.Publish(
		body,
		[]string{routingKey},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsExchange(exchange),
	)
	if err != nil {
		log.Println(err)
		return err

	}
	publisher.NotifyReturn(func(r rabbitmq.Return) {
		log.Printf("message returned from server: %s", string(r.Body))
	})

	publisher.NotifyPublish(func(c rabbitmq.Confirmation) {
		log.Printf("message confirmed from server. tag: %v, ack: %v", c.DeliveryTag, c.Ack)
	})
	return nil

}

func (manager *ConnectionManager) MessageGroupConsumeManualAck(queueName string, listener *message.ConformationMessageListener) error {

	msgs, err := manager.channel.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		log.Println(err)
		return err
	}
	go func() {
		for d := range msgs {
			log.Printf("in queue %v \n", queueName)
			result := &p_buff.MessageGroupBatch{}
			err := proto.Unmarshal(d.Body, result)
			if err != nil {
				log.Fatalf("Cann't unmarshal : %v \n", err)
			}
			delivery := Delivery{Redelivered: d.Redelivered}
			deliveryConfirm := DeliveryConfirmation{delivery: &d}
			var conf Confirmation = deliveryConfirm
			fail := (*listener).Handle(&delivery, result, &conf)
			if fail != nil {
				log.Fatalf("Cann't Handle : %v \n", fail)
			}
		}
	}()

	log.Printf(" [*] Waiting for message. To exit press CTRL+C")
	return nil
}

func (manager *ConnectionManager) MessageGroupConsume(queueName string, listener *message.MessageListener) error {

	msgs, err := manager.channel.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		log.Println(err)
		return err
	}

	go func() {
		for d := range msgs {
			log.Printf("in queue %v \n", queueName)
			result := &p_buff.MessageGroupBatch{}
			err := proto.Unmarshal(d.Body, result)
			if err != nil {
				log.Fatalf("Cann't unmarshal : %v \n", err)
			}
			delivery := Delivery{Redelivered: d.Redelivered}
			fail := (*listener).Handle(&delivery, result)
			if fail != nil {
				log.Fatalf("Cann't Handle : %v \n", fail)
			}
		}
	}()

	log.Printf(" [*] Waiting for message. To exit press CTRL+C")
	return nil
}
