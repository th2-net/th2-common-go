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

package configuration

import (
	"encoding/json"
	"log"
)

const (
	EQUAL     FilterOperation = "EQUAL"
	NOT_EQUAL FilterOperation = "NOT_EQUAL"
	EMPTY     FilterOperation = "EMPTY"
	NOT_EMPTY FilterOperation = "NOT_EMPTY"
	WILDCARD  FilterOperation = "WILDCARD"
	UNKNOWN   FilterOperation = "UNKNOWN"
)

type MqRouterFilterConfiguration struct {
	Metadata FilterSpec `json:"metadata"`
	Message  FilterSpec `json:"message"`
}

type FilterSpec struct {
	Filters []FilterFieldsConfig
}

type FilterFieldsConfig struct {
	FieldName     string
	ExpectedValue string
	Operation     FilterOperation
}

type FilterOperation string

func (fc *FilterSpec) UnmarshalJSON(data []byte) error {
	if string(data[0]) == "[" {
		FilterFields := []struct {
			FieldName     string `json:"fieldName"`
			ExpectedValue string `json:"expectedValue"`
			Operation     string `json:"operation"`
		}{}
		if err := json.Unmarshal(data, &FilterFields); err != nil {
			return err
		}
		for _, filter := range FilterFields {
			fc.Filters = append(fc.Filters, FilterFieldsConfig{FieldName: filter.FieldName, ExpectedValue: filter.ExpectedValue, Operation: pickOperation(filter.Operation)})
		}
	} else if string(data[0]) == "{" {
		type mapFilt struct {
			ExpectedValue string `json:"value"`
			Operation     string `json:"operation"`
		}
		res := map[string]mapFilt{}
		if err := json.Unmarshal(data, &res); err != nil {
			return err
		}
		for k, v := range res {
			fc.Filters = append(fc.Filters, FilterFieldsConfig{FieldName: k, ExpectedValue: v.ExpectedValue, Operation: pickOperation(v.Operation)})
		}
	}
	return nil
}

func pickOperation(operation string) FilterOperation {
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
