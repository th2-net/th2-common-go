package MessageConverter

import (
	"fmt"
	p_buff "github.com/th2-net/th2-common-go/proto"
)

func ToGroupBatch(batch *p_buff.MessageBatch) *p_buff.MessageGroupBatch {

	var messages []*p_buff.AnyMessage
	for _, val := range batch.Messages {
		message := p_buff.AnyMessage{Kind: &p_buff.AnyMessage_Message{
			Message: val,
		}}
		messages = append(messages, &message)
	}
	group := p_buff.MessageGroup{
		Messages: messages,
	}
	groups := []*p_buff.MessageGroup{&group}

	groupBatch := p_buff.MessageGroupBatch{Groups: groups}
	return &groupBatch
}

func FromGroupBatch(groupBatch *p_buff.MessageGroupBatch) *p_buff.MessageBatch {
	batch := p_buff.MessageBatch{}
	for _, group := range groupBatch.Groups {
		for _, anymsg := range group.Messages {
			message := anymsg.GetMessage()
			if message != nil {
				batch.Messages = append(batch.Messages, anymsg.GetMessage())
			} else {
				fmt.Println("no messages")
			}
		}
	}
	return &batch
}
