/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package message

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	grpcCommon "th2-grpc/th2_grpc_common"
)

func createBatch() *grpcCommon.EventBatch {
	return &grpcCommon.EventBatch{
		Events: []*grpcCommon.Event{
			{
				Id:             &grpcCommon.EventID{Id: "eventId"},
				Name:           "TestEvent",
				Type:           "TestType",
				ParentId:       &grpcCommon.EventID{Id: "parentId"},
				Status:         grpcCommon.EventStatus_FAILED,
				StartTimestamp: timestamppb.Now(),
				EndTimestamp:   timestamppb.Now(),
				Body:           []byte(`[{}]`),
			},
		},
		ParentEventId: &grpcCommon.EventID{Id: "batchParentId"},
	}
}
