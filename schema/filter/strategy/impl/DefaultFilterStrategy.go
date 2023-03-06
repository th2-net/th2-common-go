package impl

import (
	mqFilter "github.com/th2-net/th2-common-go/schema/queue/configuration"
	p_buff "th2-grpc/th2_grpc_common"
)

type DefaultFilterStrategy struct {
	extractFields Th2MsgFieldExtraction
}

func (fs *DefaultFilterStrategy) Construct() {
	fs.extractFields = Th2MsgFieldExtraction{}
}

func (fs *DefaultFilterStrategy) Verify(messages *p_buff.MessageGroup, filter []mqFilter.MqRouterFilterConfiguration) bool {
	// checking
}
