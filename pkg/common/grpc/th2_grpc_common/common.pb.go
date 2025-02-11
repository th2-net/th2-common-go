//
// Copyright 2020-2025 Exactpro (Exactpro Systems Limited)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
package th2_grpc_common

import (
	grpc "github.com/th2-net/th2-grpc-common-go"
)

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type Direction = grpc.Direction

const (
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	Direction_FIRST Direction = grpc.Direction_FIRST // Incoming message for connectivity
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	Direction_SECOND Direction = grpc.Direction_SECOND // Outgoing message for connectivity
)

// Enum value maps for Direction.
var (
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	Direction_name  = grpc.Direction_name
	Direction_value = grpc.Direction_value
)

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type NullValue = grpc.NullValue

const (
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	NullValue_NULL_VALUE NullValue = grpc.NullValue_NULL_VALUE
)

// Enum value maps for NullValue.
var (
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	NullValue_name = grpc.NullValue_name
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	NullValue_value = grpc.NullValue_value
)

// --// Settings //--//

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type FailUnexpected = grpc.FailUnexpected

const (
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FailUnexpected_NO FailUnexpected = grpc.FailUnexpected_NO // comparison won't fail in case of unexpected fields or messages
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FailUnexpected_FIELDS FailUnexpected = grpc.FailUnexpected_FIELDS // comparison will fail in case of unexpected fields only
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FailUnexpected_FIELDS_AND_MESSAGES FailUnexpected = grpc.FailUnexpected_FIELDS_AND_MESSAGES // comparison will fail in case of unexpected fields or messages
)

// Enum value maps for FailUnexpected.
var (
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FailUnexpected_name = grpc.FailUnexpected_name
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FailUnexpected_value = grpc.FailUnexpected_value
)

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type FilterOperation = grpc.FilterOperation

const (
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_EQUAL FilterOperation = grpc.FilterOperation_EQUAL
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_NOT_EQUAL FilterOperation = grpc.FilterOperation_NOT_EQUAL
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_EMPTY FilterOperation = grpc.FilterOperation_EMPTY
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_NOT_EMPTY FilterOperation = grpc.FilterOperation_NOT_EMPTY
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_IN FilterOperation = grpc.FilterOperation_IN
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_NOT_IN FilterOperation = grpc.FilterOperation_NOT_IN
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_LIKE FilterOperation = grpc.FilterOperation_LIKE
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_NOT_LIKE FilterOperation = grpc.FilterOperation_NOT_LIKE
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_MORE FilterOperation = grpc.FilterOperation_MORE
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_NOT_MORE FilterOperation = grpc.FilterOperation_NOT_MORE
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_LESS FilterOperation = grpc.FilterOperation_LESS
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_NOT_LESS FilterOperation = grpc.FilterOperation_NOT_LESS
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_WILDCARD FilterOperation = grpc.FilterOperation_WILDCARD
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_NOT_WILDCARD FilterOperation = grpc.FilterOperation_NOT_WILDCARD
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_EQ_TIME_PRECISION FilterOperation = grpc.FilterOperation_EQ_TIME_PRECISION
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_EQ_DECIMAL_PRECISION FilterOperation = grpc.FilterOperation_EQ_DECIMAL_PRECISION
)

// Enum value maps for FilterOperation.
var (
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_name = grpc.FilterOperation_name
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	FilterOperation_value = grpc.FilterOperation_value
)

// --// Event //--//
// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type EventStatus = grpc.EventStatus

const (
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	EventStatus_SUCCESS EventStatus = grpc.EventStatus_SUCCESS
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	EventStatus_FAILED EventStatus = grpc.EventStatus_FAILED
)

// Enum value maps for EventStatus.
var (
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	EventStatus_name = grpc.EventStatus_name
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	EventStatus_value = grpc.EventStatus_value
)

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type RequestStatus_Status = grpc.RequestStatus_Status

const (
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	RequestStatus_SUCCESS RequestStatus_Status = grpc.RequestStatus_SUCCESS
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	RequestStatus_ERROR RequestStatus_Status = grpc.RequestStatus_ERROR
)

// Enum value maps for RequestStatus_Status.
var (
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	RequestStatus_Status_name = grpc.RequestStatus_Status_name
	// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
	RequestStatus_Status_value = grpc.RequestStatus_Status_value
)

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type ConnectionID = grpc.ConnectionID

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type MessageID = grpc.MessageID

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type EventBatchMetadata = grpc.EventBatchMetadata

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type MessageGroupBatchMetadata = grpc.MessageGroupBatchMetadata

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type MessageMetadata = grpc.MessageMetadata

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type RawMessageMetadata = grpc.RawMessageMetadata

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type Value = grpc.Value

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type Value_NullValue = grpc.Value_NullValue

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type Value_SimpleValue = grpc.Value_SimpleValue

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type Value_MessageValue = grpc.Value_MessageValue

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type Value_ListValue = grpc.Value_ListValue

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type ListValue = grpc.ListValue

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type Message = grpc.Message

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type RawMessage = grpc.RawMessage

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type AnyMessage = grpc.AnyMessage

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type AnyMessage_Message = grpc.AnyMessage_Message

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type AnyMessage_RawMessage = grpc.AnyMessage_RawMessage

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type MessageGroup = grpc.MessageGroup

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type MessageBatch = grpc.MessageBatch

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type RawMessageBatch = grpc.RawMessageBatch

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type MessageGroupBatch = grpc.MessageGroupBatch

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type RequestStatus = grpc.RequestStatus

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type ComparisonSettings = grpc.ComparisonSettings

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type RootComparisonSettings = grpc.RootComparisonSettings

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type ValueFilter = grpc.ValueFilter

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type ValueFilter_SimpleFilter = grpc.ValueFilter_SimpleFilter

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type ValueFilter_MessageFilter = grpc.ValueFilter_MessageFilter

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type ValueFilter_ListFilter = grpc.ValueFilter_ListFilter

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type ValueFilter_SimpleList = grpc.ValueFilter_SimpleList

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type ValueFilter_NullValue = grpc.ValueFilter_NullValue

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type ListValueFilter = grpc.ListValueFilter

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type SimpleList = grpc.SimpleList

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type MessageFilter = grpc.MessageFilter

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type MetadataFilter = grpc.MetadataFilter

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type RootMessageFilter = grpc.RootMessageFilter

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type Checkpoint = grpc.Checkpoint

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type EventID = grpc.EventID

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type Event = grpc.Event

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type EventBatch = grpc.EventBatch

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type MetadataFilter_SimpleFilter = grpc.MetadataFilter_SimpleFilter

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type MetadataFilter_SimpleFilter_Value = grpc.MetadataFilter_SimpleFilter_Value

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type MetadataFilter_SimpleFilter_SimpleList = grpc.MetadataFilter_SimpleFilter_SimpleList

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type Checkpoint_SessionAliasToDirectionCheckpoint = grpc.Checkpoint_SessionAliasToDirectionCheckpoint

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type Checkpoint_DirectionCheckpoint = grpc.Checkpoint_DirectionCheckpoint

// Deprecated: moved to github.com/th2-net/th2-grpc-common-go
type Checkpoint_CheckpointData = grpc.Checkpoint_CheckpointData
