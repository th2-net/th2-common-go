package th2

import (
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog"

	amqp "github.com/streadway/amqp"

	rmq_cfg "exactpro/th2/th2-common-go/schema/message/impl/rabbitmq/configuration"
)

type ConnectionManager struct {
	rabbitMQConfiguration *rmq_cfg.RabbitMQConfiguration

	subscriber_name string
	conn            *amqp.Connection
	channel         *amqp.Channel
	debug           bool

	logger zerolog.Logger
}

// ConnectionManager constructor
func NewConnectionManager(cfg *rmq_cfg.RabbitMQConfiguration) *ConnectionManager {

	cm := ConnectionManager{}

	cm.rabbitMQConfiguration = cfg
	cm.debug = false

	return &cm
}

func (cm *ConnectionManager) Init() error {

	now := time.Now()
	umillisec := now.UnixNano() / 1000000

	if cm.rabbitMQConfiguration.GetSubscriberName() == "" {
		cm.subscriber_name = "rabbit_mq_subscriber." + strconv.FormatInt(umillisec, 10)

	} else {
		cm.subscriber_name = cm.rabbitMQConfiguration.GetSubscriberName() + "." + strconv.FormatInt(umillisec, 10)
	}

	cm.logger.Debug().Msgf("connecting amqp://%s:***@%s:%d/%s\n",
		cm.rabbitMQConfiguration.GetUsername(),
		cm.rabbitMQConfiguration.GetHost(),
		cm.rabbitMQConfiguration.GetPort(),
		cm.rabbitMQConfiguration.GetVHost())

	url := fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cm.rabbitMQConfiguration.GetUsername(),
		cm.rabbitMQConfiguration.GetPassword(),
		cm.rabbitMQConfiguration.GetHost(),
		cm.rabbitMQConfiguration.GetPort(),
		cm.rabbitMQConfiguration.GetVHost())

	var err error
	cm.conn, err = amqp.Dial(url)

	if err != nil {
		err_str := fmt.Errorf("[ConnectionManager] RabbitMQ connection failed: %s", err)
		cm.logger.Error().Msg(err_str.Error())
		return err_str
	}

	cm.channel, err = cm.conn.Channel()

	if err != nil {
		cm.conn.Close()
		err_str := fmt.Errorf("[ConnectionManager] RabbitMQ channel open failed: %s", err)
		cm.logger.Error().Msg(err_str.Error())
		return err_str
	}

	return nil
}

func (cm *ConnectionManager) BasicPublish(exchange string, routingKey string, body []byte) error {

	queue, err := cm.channel.QueueDeclare(
		routingKey,
		false,
		false,
		false,
		false,
		nil)

	if err != nil {
		fmt.Println("[ConnectionManager] Failed to declare a queue: ", err)
		return err
	}

	msg := amqp.Publishing{
		ContentType:  "text/plain",
		Body:         []byte(body),
		DeliveryMode: 2,
	}

	err = cm.channel.Publish(
		exchange,
		queue.Name,
		false,
		false,
		msg)

	if err != nil {
		fmt.Println("[ConnectionManager] Failed to publish a message: ", err)
		return err
	}

	return nil
}

func (cm *ConnectionManager) Close() {

	cm.conn.Close()
}

func (cm *ConnectionManager) SetLogger(l zerolog.Logger) {
	cm.logger = l.With().Str("component", "ConnectionManager").Logger()
}
