package th2_bench

import (
	"testing"

	p_buff "github.com/th2-net/th2-common-go/proto"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

func Benchmark_CreateRawMessage(b *testing.B) {

	for i := 0; i < b.N; i++ {

		raw_msg := p_buff.RawMessage{

			Metadata: &p_buff.RawMessageMetadata{
				Id: &p_buff.MessageID{
					ConnectionId: &p_buff.ConnectionID{
						SessionAlias: "session_alias",
					},
					Sequence: 12345678,
				},
				Timestamp: &timestamppb.Timestamp{
					Seconds: 12346578,
					Nanos:   65789102,
				},
				Properties: map[string]string{"requestId": "0", "requestRef": "1"},
			},
			Body: []byte("Message bodyXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"),
		}

		msg_id := raw_msg.Metadata.GetId()

		from_client := true
		if from_client {
			msg_id.Direction = p_buff.Direction_SECOND
		} else {
			msg_id.Direction = p_buff.Direction_FIRST
		}

	}
}
