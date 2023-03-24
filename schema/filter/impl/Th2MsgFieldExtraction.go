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

package filter

import (
	p_buff "th2-grpc/th2_grpc_common"
)

const SESSION_ALIAS_KEY = "session_alias"
const MESSAGE_TYPE_KEY = "message_type"
const DIRECTION_KEY = "direction"
const PROTOCOL_KEY = "protocol"

func GetFieldValue(anyMsg *p_buff.AnyMessage, fieldName string) string {
	if anyMsg.GetMessage() != nil {
		return msgFieldValue(anyMsg.GetMessage(), fieldName)
	} else {
		if anyMsg.GetRawMessage() != nil {
			return rawMsgFieldValue(anyMsg.GetRawMessage(), fieldName)

		} else {
			return ""
		}
	}
}

func msgFieldValue(msg *p_buff.Message, fieldName string) string {
	switch fieldName {
	case SESSION_ALIAS_KEY:
		return msg.Metadata.Id.ConnectionId.SessionAlias
	case MESSAGE_TYPE_KEY:
		return msg.Metadata.MessageType
	case DIRECTION_KEY:
		return string(msg.Metadata.Id.Direction)
	case PROTOCOL_KEY:
		return msg.Metadata.Protocol
	default:
		return ""
	}
}

func rawMsgFieldValue(msg *p_buff.RawMessage, fieldName string) string {
	switch fieldName {
	case SESSION_ALIAS_KEY:
		return msg.Metadata.Id.ConnectionId.SessionAlias
	case DIRECTION_KEY:
		return string(msg.Metadata.Id.Direction)
	case PROTOCOL_KEY:
		return msg.Metadata.Protocol
	default:
		return ""
	}
}
