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

package configuration

import (
	"log"
)

type FilterFieldsConfig struct {
	FieldName     string          `json:"fieldName"`
	ExpectedValue string          `json:"expectedValue"`
	Operation     FilterOperation `json:"operation"`
}
type FilterOperation string

const (
	EQUAL     FilterOperation = "EQUAL"
	NOT_EQUAL FilterOperation = "NOT_EQUAL"
	EMPTY     FilterOperation = "EMPTY"
	NOT_EMPTY FilterOperation = "NOT_EMPTY"
	WILDCARD  FilterOperation = "WILDCARD"
	UNKNOWN   FilterOperation = "UNKNOWN"
)

type MqRouterFilterConfiguration struct {
	Metadata []FilterFieldsConfig `json:"metadata"`
	Message  []FilterFieldsConfig `json:"message"`
	WasList  bool
}

func PickOperation(operation string) FilterOperation {
	switch operation {
	case string(EQUAL):
		return EQUAL
	case string(NOT_EQUAL):
		return NOT_EQUAL
	case string(EMPTY):
		return EMPTY
	case string(NOT_EMPTY):
		return NOT_EMPTY
	case string(WILDCARD):
		return WILDCARD
	default:
		log.Panic("unknown operation ", operation)
		return UNKNOWN
	}
}
