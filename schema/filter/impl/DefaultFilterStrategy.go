/*
 * Copyright 2022 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package filter

import (
	"github.com/IGLOU-EU/go-wildcard"
	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/schema/filter"
	mqFilter "github.com/th2-net/th2-common-go/schema/queue/configuration"
	"os"
	p_buff "th2-grpc/th2_grpc_common"
)

type defaultFilterStrategy struct {
	extractFields th2MsgFieldExtraction

	Logger zerolog.Logger
}

var Default filter.FilterStrategy = defaultFilterStrategy{Logger: zerolog.New(os.Stdout).With().Timestamp().Logger()}

func (dfs defaultFilterStrategy) Verify(messages *p_buff.MessageGroupBatch, filters []mqFilter.MqRouterFilterConfiguration) bool {
	// returns true if MessageGroupBatch entirely matches at least one filter(any) from list of filters in the queueConfig,
	// returns true if filters are not at all
	// returns false otherwise
	res := true
	if len(filters) == 0 {
		dfs.Logger.Debug().Msg("no filters for MessageGroupBatch")
		return res
	}
	for _, flt := range filters {
		for _, msgGroup := range messages.Groups {
			if !dfs.CheckValues(msgGroup, flt) {
				res = false
				break
			}
		}
		if !res {
			res = true
			continue
		}
		// as batch matches one filter (ANY), that is enough and returns true
		dfs.Logger.Debug().Msg("MessageGroupBatch matched filter")
		return res
	}
	dfs.Logger.Debug().Msg("MessageGroupBatch didn't match any filter")
	return !res
}
func (dfs defaultFilterStrategy) CheckValues(msgGroup *p_buff.MessageGroup, filter mqFilter.MqRouterFilterConfiguration) bool {
	// return true if all messages match all simple filters (metadata in this case)
	// return false if at least one message doesn't match any simple filter
	for _, anyMessage := range msgGroup.Messages {
		for _, filterFields := range filter.Metadata.Filters {
			if !checkValue(dfs.extractFields.GetFieldValue(anyMessage, filterFields.FieldName), filterFields) {
				dfs.Logger.Debug().Msg("message didn't match filter")
				return false
			}
		}
	}
	// if there was not any mismatching in metadata, therefore ALL messages matches ALL filters and returning true
	dfs.Logger.Debug().Msg("MessageGroup matched all filters from metadata filter object")
	return true
}

func checkValue(value string, filter mqFilter.FilterFieldsConfig) bool {
	if value == "" {
		return false
	}
	switch filter.Operation {
	case mqFilter.EQUAL:
		return value == filter.ExpectedValue
	case mqFilter.NOT_EQUAL:
		return value != filter.ExpectedValue
	case mqFilter.EMPTY:
		return len(value) == 0
	case mqFilter.NOT_EMPTY:
		return len(value) != 0
	case mqFilter.WILDCARD:
		return wildcard.Match(filter.ExpectedValue, value)
	default:
		return false
	}
}
