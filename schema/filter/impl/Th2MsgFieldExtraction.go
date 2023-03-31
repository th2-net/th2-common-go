/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
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

type th2MsgFieldExtraction struct{}

const SESSION_ALIAS_KEY = "session_alias"
const MESSAGE_TYPE_KEY = "message_type"
const DIRECTION_KEY = "direction"
const PROTOCOL_KEY = "protocol"

// here "GetFieldValue" also can be just function, don't need th2MsgFieldExtraction this case. Do I need to change to that version?
func (mfe th2MsgFieldExtraction) GetFieldValue(anyMsg *p_buff.AnyMessage, fieldName string) string {
	if anyMsg.GetMessage() != nil {
		return mfe.msgFieldValue(anyMsg.GetMessage(), fieldName)
	}
	if anyMsg.GetRawMessage() != nil {
		return mfe.rawMsgFieldValue(anyMsg.GetRawMessage(), fieldName)
	}
	return ""
}

func (mfe th2MsgFieldExtraction) msgFieldValue(msg *p_buff.Message, fieldName string) string {
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

func (mfe th2MsgFieldExtraction) rawMsgFieldValue(msg *p_buff.RawMessage, fieldName string) string {
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

func ExtractID(ex interface{}) *p_buff.MessageID {
	castedG, successG := ex.(*p_buff.MessageGroup)
	if successG {
		if castedG.Messages[0].GetRawMessage() != nil {
			return castedG.Messages[0].GetRawMessage().Metadata.Id
		} else {
			return castedG.Messages[0].GetMessage().Metadata.Id
		}
	}
	castedB, successB := ex.(*p_buff.MessageGroupBatch)
	if successB {
		if castedB.Groups[0].Messages[0].GetRawMessage() != nil {
			return castedB.Groups[0].Messages[0].GetRawMessage().Metadata.Id
		} else {
			return castedB.Groups[0].Messages[0].GetMessage().Metadata.Id
		}
	}
	castedM, successM := ex.(*p_buff.AnyMessage)
	if successM {
		if castedM.GetRawMessage() != nil {
			return castedM.GetRawMessage().Metadata.Id
		} else {
			return castedM.GetMessage().Metadata.Id
		}
	}
	return nil
}
