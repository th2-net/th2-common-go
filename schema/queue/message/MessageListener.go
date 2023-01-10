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

	"github.com/th2-net/th2-common-go/schema/queue/MQcommon"
)

type MessageListener interface {
	MQcommon.CloseListener
	Handle(delivery *MQcommon.Delivery, batch *p_buff.MessageGroupBatch) error
}

type ConformationMessageListener interface {
	MQcommon.CloseListener
	Handle(delivery *MQcommon.Delivery, batch *p_buff.MessageGroupBatch, confirm *MQcommon.Confirmation) error
}
