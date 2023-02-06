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
	p_buff "th2-grpc/th2_grpc_common"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"

	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
	"google.golang.org/protobuf/proto"
)

var OUTGOING_MSG_SIZE = promauto.NewCounter(
	prometheus.CounterOpts{
		Name: "th2_rabbitmq_message_size_publish_bytes",
		Help: "Amount of bytes sent",
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
		sender.Logger.Fatal().Msg("Value for send can't be null")
	}
	body, err := proto.Marshal(batch)
	if err != nil {
		sender.Logger.Panic().Err(err).Msg("Error during marshaling message into proto message")
		return err
	}
	OUTGOING_MSG_SIZE.Add(float64(len(body)))

	fail := sender.ConnManager.Publisher.Publish(body, sender.sendQueue, sender.exchangeName)
	if fail != nil {
		return fail
	}
	// th2Pin will be used for Metrics
	return nil
}
