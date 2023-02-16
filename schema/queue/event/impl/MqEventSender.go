package event

import (
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"google.golang.org/protobuf/proto"
)

const (
	TH2_TYPE = "EVENT"
)

var th2_event_publish_total = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_event_publish_total",
		Help: "Quantity of outgoing events",
	},
	[]string{"th2Pin"},
)

type CommonEventSender struct {
	ConnManager  *MQcommon.ConnectionManager
	exchangeName string
	sendQueue    string
	th2Pin       string

	Logger zerolog.Logger
}

func (sender *CommonEventSender) Send(batch *p_buff.EventBatch) error {

	if batch == nil {
		sender.Logger.Fatal().Msg("Value for send can't be null")
	}
	body, err := proto.Marshal(batch)
	if err != nil {
		if err != nil {
			sender.Logger.Panic().Err(err).Msg("Error during marshaling message into proto event")
		}
		return err
	}

	fail := sender.ConnManager.Publisher.Publish(body, sender.sendQueue, sender.exchangeName, sender.th2Pin, TH2_TYPE)
	if fail != nil {
		return fail
	}

	th2_event_publish_total.WithLabelValues(sender.th2Pin).Add(float64(len(batch.Events)))
	return nil
}
