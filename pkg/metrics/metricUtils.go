/*
 * Copyright 2023-2025 Exactpro (Exactpro Systems Limited)
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

package metrics

import (
	p_buff "github.com/th2-net/th2-grpc-common-go"

	"github.com/prometheus/client_golang/prometheus"
)

func UpdateMessageMetrics(batch *p_buff.MessageGroupBatch, counter *prometheus.CounterVec, th2Pin string) {
	for _, group := range batch.Groups {
		msg := group.Messages[0]
		switch msg.GetKind().(type) {
		case *p_buff.AnyMessage_RawMessage:
			msg := msg.GetRawMessage()
			counter.WithLabelValues(
				th2Pin,
				msg.Metadata.Id.ConnectionId.SessionAlias,
				string(msg.Metadata.Id.Direction),
				RawMessageType,
			).Add(float64(len(group.Messages)))
		case *p_buff.AnyMessage_Message:
			msg := msg.GetMessage()
			counter.WithLabelValues(
				th2Pin,
				msg.Metadata.Id.ConnectionId.SessionAlias,
				string(msg.Metadata.Id.Direction),
				ParsedMessageType,
			).Add(float64(len(group.Messages)))
		}
	}
}
