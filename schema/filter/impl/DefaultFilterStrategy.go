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
	mqFilter "github.com/th2-net/th2-common-go/schema/queue/configuration"
	p_buff "th2-grpc/th2_grpc_common"
)

func CheckValues(msgGroup *p_buff.MessageGroup, filter mqFilter.MqRouterFilterConfiguration, logger zerolog.Logger) bool {
	// return true if all messages match all simple filters (metadata in this case)
	// return false if at least one message doesn't match any simple filter
	for _, anyMessage := range msgGroup.Messages {
		for _, filterFields := range filter.Metadata {
			if filter.WasList {
				if checkValue(GetFieldValue(anyMessage, filterFields.FieldName), filterFields) {
					logger.Debug().Msg("message matched filter")
					return true
				}
			} else {
				if !checkValue(GetFieldValue(anyMessage, filterFields.FieldName), filterFields) {
					logger.Debug().Msg("message didn't match filter")
					return false
				}
			}
		}
	}
	if filter.WasList {
		// if there was not any matching in metadata list, therefore none of messages matches filters and returning false
		logger.Debug().Msg("MessageGroup didn't match any filter from metadata filter list")
		return false
	} else {
		// if there was not any mismatching in metadata object, therefore ALL messages matches filters and returning true
		logger.Debug().Msg("MessageGroup matched all filters from metadata filter object")
		return true
	}
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
