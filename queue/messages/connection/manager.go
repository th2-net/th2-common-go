package connection

import (
	"fmt"
	"github.com/streadway/amqp"
	p_buff "github.com/th2-net/th2-common-go/proto"
	"github.com/th2-net/th2-common-go/queue/common"
	mq "github.com/th2-net/th2-common-go/queue/configuration"
	message "github.com/th2-net/th2-common-go/queue/messages"
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
	conn         *amqp.Connection
	channel      *amqp.Channel
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
	manager.connect()
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

func (manager *ConnectionManager) connect() {
	conn, err := amqp.Dial(manager.url)
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

func (manager *ConnectionManager) BasicConsumeManualAck(queueName string, listener *message.ConformationMessageListener) error {

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
	var forever chan struct{}
	go func() {
		for d := range msgs {
			log.Printf("in queue %v \n", queueName)
			result := &p_buff.MessageGroupBatch{}
			err := proto.Unmarshal(d.Body, result)
			if err != nil {
				log.Fatalf("Cann't unmarshal : %v \n", err)
			}
			delivery := common.Delivery{Redelivered: d.Redelivered}
			deliveryConfirm := DeliveryConfirmation{delivery: &d}
			var conf common.Confirmation = deliveryConfirm
			fail := (*listener).Handle(&delivery, result, &conf)
			if fail != nil {
				log.Fatalf("Cann't Handle : %v \n", fail)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
	return nil
}

func (manager *ConnectionManager) BasicConsume(queueName string, listener *message.MessageListener) error {

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
			delivery := common.Delivery{Redelivered: d.Redelivered}
			fail := (*listener).Handle(&delivery, result)
			if fail != nil {
				log.Fatalf("Cann't Handle : %v \n", fail)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	return nil
}

func (manager *ConnectionManager) CloseConn() error {
	log.Println("Closing")

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
