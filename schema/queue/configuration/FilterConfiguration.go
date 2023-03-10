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
	"fmt"
	"log"
)

const (
	FIELDNAME     string = "fieldName"
	VALUE         string = "value"
	OPERATION     string = "operation"
	EXPECTEDVALUE string = "expectedValue"
)

// still thinking about structure of filters, uploaded current state before I will come up with this solution

type FilterFieldsConfiguration struct {
	FieldName     string
	ExpectedValue string
	Operation     FilterOperation
}
type FilterOperation string

const (
	EQUAL     FilterOperation = "EQUAL"
	NOT_EQUAL FilterOperation = "NOT_EQUAL"
	EMPTY     FilterOperation = "EMPTY"
	NOT_EMPTY FilterOperation = "NOT_EMPTY"
	WILDCARD  FilterOperation = "WILDCARD"
)

func ConfigureFilters(filters []MqRouterFilterConfiguration) {
	for _, filter := range filters {
		metadatas := []FilterFieldsConfiguration{}
		switch mType := filter.Metadata.(type) {
		case map[string]interface{}:
			log.Println("reading map")
			for k, v := range filter.Metadata.(map[string]interface{}) {
				metadata := FilterFieldsConfiguration{}
				metadata.FieldName = k
				metadata.Operation = pickOperation(v.(map[string]interface{})["operation"].(string))
				metadata.ExpectedValue = v.(map[string]interface{})["value"].(string)
				metadatas = append(metadatas, metadata)
			}
		case []interface{}:
			log.Println("reading slice")
			for _, metaD := range mType {
				filed := metaD.(map[string]interface{})
				metadata := FilterFieldsConfiguration{}
				metadata.FieldName = filed["fieldName"].(string)
				metadata.Operation = pickOperation(filed["operation"].(string))
				metadata.ExpectedValue = filed["expectedValue"].(string)
				metadatas = append(metadatas, metadata)
			}
		default:
			fmt.Println(mType, " is of a type I don't know how to handle")
		}
		filter.Metadata = metadatas
	}
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
		log.Panic("wrong operation ", operation)
		return ""
	}
}
