package message

import p_buff "github.com/th2-net/th2-common-go/proto"

type MessageGroupBatchSender interface {
	//Start()
	//isClose()
	Send(batch p_buff.MessageGroupBatch)
	//close()
}
