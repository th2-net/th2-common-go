/*
 * Copyright 2023-2025 Exactpro (Exactpro Systems Limited)
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

package message

import (
	"github.com/th2-net/th2-common-go/test/modules/rabbitmq"
	grpcCommon "github.com/th2-net/th2-grpc-common-go"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func createBatch() *grpcCommon.MessageGroupBatch {
	return &grpcCommon.MessageGroupBatch{
		Groups: []*grpcCommon.MessageGroup{
			{
				Messages: []*grpcCommon.AnyMessage{
					{
						Kind: &grpcCommon.AnyMessage_RawMessage{
							RawMessage: &grpcCommon.RawMessage{
								Body: []byte("hello"),
								Metadata: &grpcCommon.RawMessageMetadata{
									Id: &grpcCommon.MessageID{
										BookName: rabbitmq.TestBook,
										ConnectionId: &grpcCommon.ConnectionID{
											SessionAlias: "alias",
											SessionGroup: "group",
										},
										Direction:   grpcCommon.Direction_FIRST,
										Sequence:    42,
										Subsequence: []uint32{1},
										Timestamp:   timestamppb.Now(),
									},
									Properties: map[string]string{
										"prop": "value",
									},
									Protocol: "protocol",
								},
							},
						},
					},
				},
			},
		},
	}
}
