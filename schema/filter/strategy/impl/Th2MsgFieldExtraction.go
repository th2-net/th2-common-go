package impl

import (
	"log"
	p_buff "th2-grpc/th2_grpc_common"
)

type Th2MsgFieldExtraction struct{}

const SESSION_ALIAS_KEY = "session_alias"
const MESSAGE_TYPE_KEY = "message_type"
const DIRECTION_KEY = "direction"
const PROTOCOL_KEY = "protocol"

func (mfe *Th2MsgFieldExtraction) GetFields(anyMsg *p_buff.AnyMessage) map[string]string {
	if anyMsg.GetMessage() != nil {
		return mfe.messageMetadataToMap(anyMsg.GetMessage())
	} else {
		if anyMsg.GetRawMessage() != nil {
			return mfe.rawMessageMetadataToMap(anyMsg.GetRawMessage())

		} else {
			log.Printf("wrong type message %T \n", anyMsg)
			return nil
		}
	}
}

func (mfe *Th2MsgFieldExtraction) messageMetadataToMap(msg *p_buff.Message) map[string]string {
	metadata := make(map[string]string)
	metadata[SESSION_ALIAS_KEY] = msg.Metadata.Id.ConnectionId.SessionAlias
	metadata[DIRECTION_KEY] = string(msg.Metadata.Id.Direction)
	metadata[MESSAGE_TYPE_KEY] = msg.Metadata.MessageType
	metadata[PROTOCOL_KEY] = msg.Metadata.Protocol
	return metadata
}

func (mfe *Th2MsgFieldExtraction) rawMessageMetadataToMap(msg *p_buff.RawMessage) map[string]string {
	metadata := make(map[string]string)
	metadata[SESSION_ALIAS_KEY] = msg.Metadata.Id.ConnectionId.SessionAlias
	metadata[DIRECTION_KEY] = string(msg.Metadata.Id.Direction)
	metadata[PROTOCOL_KEY] = msg.Metadata.Protocol
	return metadata
}
