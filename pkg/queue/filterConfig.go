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

package queue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	Equal    FilterOperation = "EQUAL"
	NotEqual FilterOperation = "NOT_EQUAL"
	Empty    FilterOperation = "EMPTY"
	NotEmpty FilterOperation = "NOT_EMPTY"
	Wildcard FilterOperation = "WILDCARD"
)

var (
	filterOperationsMap = map[string]FilterOperation{
		string(Equal):    Equal,
		string(NotEqual): NotEqual,
		string(Empty):    Empty,
		string(NotEmpty): NotEmpty,
		string(Wildcard): Wildcard,
	}
)

type MqRouterFilterConfiguration struct {
	Metadata FilterSpec `json:"metadata"`
	Message  FilterSpec `json:"message"`
}

type FilterSpec struct {
	Filters []FilterFieldsConfig
}

type FilterOperation string

func (op *FilterOperation) UnmarshalJSON(data []byte) error {
	var operation string
	if err := json.Unmarshal(data, &operation); err != nil {
		return err
	}
	operation = strings.TrimSpace(operation)
	value, ok := filterOperationsMap[operation]
	if !ok {
		return fmt.Errorf("unknown operation '%s'", operation)
	}
	*op = value
	return nil
}

type FilterFieldsConfig struct {
	FieldName     string
	ExpectedValue string
	Operation     FilterOperation
}

func (fc *FilterSpec) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	switch token {
	case json.Delim('['):
		var filterFields []struct {
			FieldName     string          `json:"fieldName"`
			ExpectedValue string          `json:"expectedValue"`
			Operation     FilterOperation `json:"operation"`
		}
		if err := json.Unmarshal(data, &filterFields); err != nil {
			return err
		}
		for _, filter := range filterFields {
			fc.Filters = append(fc.Filters, FilterFieldsConfig{FieldName: filter.FieldName, ExpectedValue: filter.ExpectedValue, Operation: filter.Operation})
		}
	case json.Delim('{'):
		type mapFilt struct {
			ExpectedValue string          `json:"value"`
			Operation     FilterOperation `json:"operation"`
		}
		res := map[string]mapFilt{}
		if err := json.Unmarshal(data, &res); err != nil {
			return err
		}
		for k, v := range res {
			fc.Filters = append(fc.Filters, FilterFieldsConfig{FieldName: k, ExpectedValue: v.ExpectedValue, Operation: v.Operation})
		}
	default:
		return fmt.Errorf("expect either array or single object but had %v", token)
	}
	return nil
}
