/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
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
	mqFilter "github.com/th2-net/th2-common-go/schema/queue/configuration"
	"os"
	p_buff "th2-grpc/th2_grpc_common"
)

type defaultFilterStrategy struct {
	extractFields th2MsgFieldExtraction

	logger zerolog.Logger
}

var Default FilterStrategy = defaultFilterStrategy{logger: zerolog.New(os.Stdout).With().Str("component", "default_filter_strategy").Timestamp().Logger()}

func (dfs defaultFilterStrategy) Verify(messages *p_buff.MessageGroupBatch, filters []mqFilter.MqRouterFilterConfiguration) bool {
	// returns true if MessageGroupBatch entirely matches at least one filter(any) from list of filters in the queueConfig,
	// returns true if filters are not at all
	// returns false otherwise
	if len(filters) == 0 {
		return true
	}
	dfs.logger.Debug().Msg("Started filtering of MessageGroupBatch for sending")
	for n, flt := range filters {
		res := true
		for _, msgGroup := range messages.Groups {
			if !dfs.CheckValues(msgGroup, flt) {
				res = false
				if e := dfs.logger.Debug(); e.Enabled() {
					e.Int("filter N", n+1).
						Interface("Metadata", flt.Metadata.Filters).
						Msgf("Filter fields to which didn't match MessageGroupBatch with first message id: %v  ", IDFromMsgGroup(msgGroup))
				}
				break
			}
		}
		if res {
			// as batch matches one filter (ANY), that is enough and returns true
			if e := dfs.logger.Debug(); e.Enabled() {
				e.Int("filter N", n+1).
					Interface("Metadata", flt.Metadata.Filters).
					Msg("Filter fields to which matched MessageGroupBatch")
			}
			return res
		}
	}
	return false
}
func (dfs defaultFilterStrategy) CheckValues(msgGroup *p_buff.MessageGroup, filter mqFilter.MqRouterFilterConfiguration) bool {
	// return true if all messages match all simple filters (metadata in this case)
	// return false if at least one message doesn't match any simple filter
	for _, anyMessage := range msgGroup.Messages {
		for _, filterFields := range filter.Metadata.Filters {
			if !checkValue(dfs.extractFields.GetFieldValue(anyMessage, filterFields.FieldName), filterFields) {
				if e := dfs.logger.Debug(); e.Enabled() {
					e.Interface("Filter", filterFields).Msgf("Message with id: %v didn't match filter with fields: %v ", IDFromAnyMsg(anyMessage))
				}
				return false
			}
		}
	}
	// if there was not any mismatching in metadata, therefore ALL messages matches ALL filters and returning true
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
