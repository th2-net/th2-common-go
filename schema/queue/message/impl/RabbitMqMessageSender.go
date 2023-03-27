/*
 * Copyright 2022 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package message

import (
	"errors"
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"github.com/th2-net/th2-common-go/schema/metrics"
	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"google.golang.org/protobuf/proto"
)

var th2_message_publish_total = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "th2_message_publish_total",
		Help: "Quantity of outgoing messages",
	},
	[]string{
		metrics.DEFAULT_TH2_PIN_LABEL_NAME,
		metrics.DEFAULT_SESSION_ALIAS_LABEL_NAME,
		metrics.DEFAULT_DIRECTION_LABEL_NAME,
		metrics.DEFAULT_MESSAGE_TYPE_LABEL_NAME,
	},
)

type CommonMessageSender struct {
	ConnManager  *MQcommon.ConnectionManager
	exchangeName string
	sendQueue    string
	th2Pin       string

	Logger zerolog.Logger
}

func (sender *CommonMessageSender) Send(batch *p_buff.MessageGroupBatch) error {

	if batch == nil {
		return errors.New("value for send can't be null")
	}
	body, err := proto.Marshal(batch)
	if err != nil {
		sender.Logger.Panic().Err(err).Msg("Error during marshaling message into proto message")
		return err
	}

	fail := sender.ConnManager.Publisher.Publish(body, sender.sendQueue, sender.exchangeName, sender.th2Pin, metrics.MESSAGE_GROUP_TH2_TYPE)
	if fail != nil {
		return fail
	}
	metrics.UpdateMessageMetrics(batch, th2_message_publish_total, sender.th2Pin)

	return nil
}

func (sender *CommonMessageSender) SendRaw(data []byte) error {
	if data == nil {
		return errors.New("nil raw data")
	}
	return sender.ConnManager.Publisher.Publish(data, sender.sendQueue, sender.exchangeName, sender.th2Pin, metrics.MESSAGE_GROUP_TH2_TYPE)
}
