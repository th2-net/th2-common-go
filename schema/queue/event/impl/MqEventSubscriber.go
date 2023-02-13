package event

import (
	"os"
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"github.com/streadway/amqp"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"github.com/th2-net/th2-common-go/schema/queue/configuration"
	"github.com/th2-net/th2-common-go/schema/queue/event"
	"google.golang.org/protobuf/proto"
)

var INCOMING_EVENTS_QUANTITY = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "th2_event_subscribe_total",
		Help: "Amount of events received",
	},
)

type CommonEventSubscriber struct {
	connManager          *MQcommon.ConnectionManager
	qConfig              *configuration.QueueConfig
	listener             *event.EventListener
	confirmationListener *event.ConformationEventListener
	th2Pin               string

	Logger zerolog.Logger
}

func (cs *CommonEventSubscriber) Start() error {
	err := cs.connManager.Consumer.Consume(cs.qConfig.QueueName, cs.Handler)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonEventSubscriber) ConfirmationStart() error {
	err := cs.connManager.Consumer.ConsumeWithManualAck(cs.qConfig.QueueName, cs.ConfirmationHandler)
	if err != nil {
		return err
	}
	return nil
	//use th2Pin for metrics
}

func (cs *CommonEventSubscriber) Handler(msgDelivery amqp.Delivery) {
	result := &p_buff.EventBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.Logger.Fatal().Err(err).Msg("Can't unmarshal proto")
	}
	INCOMING_EVENTS_QUANTITY.Add(len(result.Events))
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	if cs.listener == nil {
		cs.Logger.Fatal().Msgf("No Listener to Handle : %s ", cs.listener)
	}
	handleErr := (*cs.listener).Handle(&delivery, result)
	if handleErr != nil {
		cs.Logger.Fatal().Err(handleErr).Msg("Can't Handle")
	}
	cs.Logger.Debug().Msg("Successfully Handled")
}

func (cs *CommonEventSubscriber) ConfirmationHandler(msgDelivery amqp.Delivery) {
	result := &p_buff.EventBatch{}
	err := proto.Unmarshal(msgDelivery.Body, result)
	if err != nil {
		cs.Logger.Fatal().Err(err).Msg("Can't unmarshal proto")
	}
	INCOMING_EVENTS_QUANTITY.Add(len(result.Events))
	delivery := MQcommon.Delivery{Redelivered: msgDelivery.Redelivered}
	deliveryConfirm := MQcommon.DeliveryConfirmation{Delivery: &msgDelivery, Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
	var confirmation MQcommon.Confirmation = deliveryConfirm

	if cs.confirmationListener == nil {
		cs.Logger.Fatal().Msgf("No Confirmation Listener to Handle : %s ", cs.confirmationListener)
	}
	handleErr := (*cs.confirmationListener).Handle(&delivery, result, &confirmation)
	if handleErr != nil {
		cs.Logger.Fatal().Err(handleErr).Msg("Can't Handle")
	}
	cs.Logger.Debug().Msg("Successfully Handled")
}

func (cs *CommonEventSubscriber) RemoveListener() {
	cs.listener = nil
	cs.confirmationListener = nil
	cs.Logger.Info().Msg("Removed listeners")
}

func (cs *CommonEventSubscriber) AddListener(listener *event.EventListener) {
	cs.listener = listener
	cs.Logger.Debug().Msg("Added listener")
}

func (cs *CommonEventSubscriber) AddConfirmationListener(listener *event.ConformationEventListener) {
	cs.confirmationListener = listener
	cs.Logger.Debug().Msg("Added confirmation listener")
}

type SubscriberMonitor struct {
	subscriber *CommonEventSubscriber
}

func (sub SubscriberMonitor) Unsubscribe() error {
	if sub.subscriber.listener != nil {
		err := (*sub.subscriber.listener).OnClose()
		if err != nil {
			return err
		}
		sub.subscriber.RemoveListener()
	}

	if sub.subscriber.confirmationListener != nil {
		err := (*sub.subscriber.confirmationListener).OnClose()
		if err != nil {
			return err
		}
		sub.subscriber.RemoveListener()
	}
	return nil
}

type MultiplySubscribeMonitor struct {
	subscriberMonitors []SubscriberMonitor
}

func (sub MultiplySubscribeMonitor) Unsubscribe() error {
	for _, subM := range sub.subscriberMonitors {
		err := subM.Unsubscribe()
		if err != nil {
			return err
		}
	}
	return nil
}
