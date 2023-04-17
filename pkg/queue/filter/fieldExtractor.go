/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
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

package filter

import (
	p_buff "th2-grpc/th2_grpc_common"
)

type th2MsgFieldExtraction struct{}

const (
	SessionAliasKey = "session_alias"
	MessageTypeKey  = "message_type"
	DirectionKey    = "direction"
	ProtocolKey     = "protocol"
)

// GetFieldValue also can be just function, don't need th2MsgFieldExtraction this case. Do I need to change to that version?
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
	case SessionAliasKey:
		return msg.Metadata.Id.ConnectionId.SessionAlias
	case MessageTypeKey:
		return msg.Metadata.MessageType
	case DirectionKey:
		return string(msg.Metadata.Id.Direction)
	case ProtocolKey:
		return msg.Metadata.Protocol
	default:
		return ""
	}
}

func (mfe th2MsgFieldExtraction) rawMsgFieldValue(msg *p_buff.RawMessage, fieldName string) string {
	switch fieldName {
	case SessionAliasKey:
		return msg.Metadata.Id.ConnectionId.SessionAlias
	case DirectionKey:
		return string(msg.Metadata.Id.Direction)
	case ProtocolKey:
		return msg.Metadata.Protocol
	default:
		return ""
	}
}

func FirstIDFromMsgGroup(group *p_buff.MessageGroup) *p_buff.MessageID {
	if group.Messages[0].GetRawMessage() != nil {
		return group.Messages[0].GetRawMessage().Metadata.Id
	} else {
		return group.Messages[0].GetMessage().Metadata.Id
	}
}
func FirstIDFromMsgBatch(batch *p_buff.MessageGroupBatch) *p_buff.MessageID {
	if batch.Groups[0].Messages[0].GetRawMessage() != nil {
		return batch.Groups[0].Messages[0].GetRawMessage().Metadata.Id
	} else {
		return batch.Groups[0].Messages[0].GetMessage().Metadata.Id
	}
}

func IDFromAnyMsg(msg *p_buff.AnyMessage) *p_buff.MessageID {
	if msg.GetRawMessage() != nil {
		return msg.GetRawMessage().Metadata.Id
	} else {
		return msg.GetMessage().Metadata.Id
	}
}
