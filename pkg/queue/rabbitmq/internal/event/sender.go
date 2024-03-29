/*
 * Copyright 2022-2023 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package event

import (
	"errors"
	p_buff "github.com/th2-net/th2-common-go/pkg/common/grpc/th2_grpc_common"
	"github.com/th2-net/th2-common-go/pkg/queue/rabbitmq/internal/connection"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"github.com/th2-net/th2-common-go/pkg/metrics"
	"google.golang.org/protobuf/proto"
)

var (
	errNullMsg = errors.New("null value for sending")
)

var th2EventPublishTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_event_publish_total",
		Help: "Quantity of outgoing events",
	},
	[]string{metrics.DefaultTh2PinLabelName},
)

type CommonEventSender struct {
	ConnManager  *connection.Manager
	exchangeName string
	sendQueue    string
	th2Pin       string

	Logger zerolog.Logger
}

func (sender *CommonEventSender) Send(batch *p_buff.EventBatch) error {

	if batch == nil {
		sender.Logger.Error().
			Str("routingKey", sender.sendQueue).
			Str("exchange", sender.exchangeName).
			Msg("Value for send can't be null")
		return errNullMsg
	}
	body, err := proto.Marshal(batch)
	if err != nil {
		sender.Logger.Error().
			Err(err).
			Str("routingKey", sender.sendQueue).
			Str("exchange", sender.exchangeName).
			Msg("Error during marshaling message into proto event")
		return err
	}

	fail := sender.ConnManager.Publisher.Publish(body, sender.sendQueue, sender.exchangeName, sender.th2Pin, metrics.EventTh2Type)
	if fail != nil {
		return fail
	}

	th2EventPublishTotal.WithLabelValues(sender.th2Pin).Add(float64(len(batch.Events)))
	return nil
}
