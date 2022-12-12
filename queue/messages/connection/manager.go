package connection

import (
	"fmt"
	p_buff "github.com/th2-net/th2-common-go/proto"
	mq "github.com/th2-net/th2-common-go/queue/configuration"
	q_conf "github.com/th2-net/th2-common-go/queue/queueConfiguration"
	"github.com/wagslane/go-rabbitmq"
	"google.golang.org/protobuf/proto"
	"log"
	"strconv"
)

type ConnectionManager struct {
	QConfig      *q_conf.MessageRouterConfiguration
	MqConnConfig *mq.RabbitMQConfiguration
	url          string
}

func (manager *ConnectionManager) Init(queueConfig string, connConfig string) {

	mqConnConfig := mq.RabbitMQConfiguration{}
	err := mqConnConfig.Init(connConfig)
	failOnError(err, "Initialization error")
	manager.MqConnConfig = &mqConnConfig

	MQConfig := q_conf.MessageRouterConfiguration{}
	fail := MQConfig.Init(queueConfig)
	failOnError(fail, "Initialization error")
	manager.QConfig = &MQConfig

	port, err := strconv.Atoi(manager.MqConnConfig.Port)
	if err != nil {
		log.Fatalf("%v", err)
	}
	manager.url = fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		manager.MqConnConfig.Username,
		manager.MqConnConfig.Password,
		manager.MqConnConfig.Host,
		port,
		manager.MqConnConfig.VHost)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func (manager *ConnectionManager) BasicPublish(message *p_buff.MessageGroupBatch, routingKey string, exchange string) error {
	conn, err := rabbitmq.NewConn(
		manager.url,
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
		failOnError(err, "Error during marshaling message into proto message. ")
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

//func (manager *MqManager) BasicConsume(queueName string, msgBatch *p_buff.MessageBatch) {
//
//	conn, err := amqp.Dial(manager.url)
//	failOnError(err, "Failed to connect to RabbitMQ")
//	defer conn.Close()
//
//	ch, err := conn.Channel()
//	failOnError(err, "Failed to open a channel")
//	defer ch.Close()
//
//	msgs, err := ch.Consume(
//		queueName, // queue
//		"",        // consumer
//		false,     // auto-ack
//		false,     // exclusive
//		false,     // no-local
//		false,     // no-wait
//		nil,       // args
//	)
//	failOnError(err, "Failed to register a consumer")
//
//	var messages []*p_buff.Message
//	var list []amqp.Delivery
//	var forever chan struct{}
//	go func() {
//		for d := range msgs {
//			list = append(list, d)
//		}
//	}()
//
//	go func() {
//		for {
//			time.Sleep(10 * time.Second)
//			if len(list) > 0 {
//				for _, v := range list {
//					result := &p_buff.Message{}
//					proto.Unmarshal(v.Body, result)
//					log.Printf("consumed: %T", result)
//					v.Ack(false)
//					messages = append(messages, result)
//				}
//				list = list[:0]
//				msgBatch.Messages = messages
//			}
//		}
//	}()
//
//	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
//	<-forever
//
//}
