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

package strategy

import (
	"github.com/IGLOU-EU/go-wildcard"
	"github.com/th2-net/th2-common-go/schema/filter/strategy"

	mqFilter "github.com/th2-net/th2-common-go/schema/queue/configuration"
	p_buff "th2-grpc/th2_grpc_common"
)

type defaultFilterStrategy struct {
	extractFields th2MsgFieldExtraction
}

var Default strategy.FilterStrategy = defaultFilterStrategy{}

func (fs defaultFilterStrategy) Verify(messages *p_buff.MessageGroupBatch, filters []mqFilter.MqRouterFilterConfiguration) bool {
	// returns true if MessageGroupBatch entirely matches at least one filter(any) from list of filters in the queueConfig,
	// returns true if filters are not at all
	// returns false otherwise
	res := true
	if len(filters) != 0 {
		for _, filter := range filters {
			for _, msgGroup := range messages.Groups {
				if !fs.CheckValues(msgGroup, filter.Metadata.([]mqFilter.FilterFieldsConfiguration)) {
					res = false
					break
				}
			}
			if !res {
				res = true
				continue
			} else {
				return res
			}
		}
		//if code riches here, meaning that batch didn't match any filter, so returns false
		return !res
	} else {
		return res
	}

}

func (fs defaultFilterStrategy) CheckValues(msgGroup *p_buff.MessageGroup, metadataFilters []mqFilter.FilterFieldsConfiguration) bool {
	// return true if all messages match all simple filters (metadatas in this case)
	// return false if at least one message doesn't match any simple filter
	for _, anyMessage := range msgGroup.Messages {
		messageFields := fs.extractFields.GetFields(anyMessage)
		for _, filter := range metadataFilters {
			for name, value := range messageFields {
				if filter.FieldName == name {
					if !checkValue(value, filter) {
						return false
					}
				}
			}
		}
	}
	// if there was not any mismatching, therefore all messages matches filters and returning true
	return true

}

func checkValue(value string, filter mqFilter.FilterFieldsConfiguration) bool {
	expected := filter.ExpectedValue
	switch filter.Operation {
	case mqFilter.EQUAL:
		return value == expected
	case mqFilter.NOT_EQUAL:
		return value != expected
	case mqFilter.EMPTY:
		return len(value) == 0
	case mqFilter.NOT_EMPTY:
		return len(value) != 0
	case mqFilter.WILDCARD:
		return wildcard.Match(expected, value)
	default:
		return false
	}
}
