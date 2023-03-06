package strategy

import (
	mqFilter "github.com/th2-net/th2-common-go/schema/queue/configuration"
	p_buff "th2-grpc/th2_grpc_common"
)

type FilterStrategy interface {
	Verify(messages *p_buff.MessageGroup, filters []mqFilter.MqRouterFilterConfiguration) bool
}
