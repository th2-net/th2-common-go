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
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"log"
	"os"
)

type QueueConfigFields struct {
	RoutingKey string   `json:"name"`
	QueueName  string   `json:"queue"`
	Exchange   string   `json:"exchange"`
	Attributes []string `json:"attributes"`
}
type QueueConfig struct {
	QueueConfigFields
	Filters []MqRouterFilterConfiguration `json:"filters"`
}

type MessageRouterConfiguration struct {
	Queues map[string]QueueConfig `json:"queues"`

	Logger zerolog.Logger
}

func (mrc *MessageRouterConfiguration) Init(path string) error {
	content, err := os.ReadFile(path) // Read json file
	if err != nil {
		mrc.Logger.Error().Err(err).Msg("Json file reading error for QueueConfig")
		return err
	}
	fail := mrc.UnmarshalJSON(content)
	if fail != nil {
		mrc.Logger.Error().Err(err).Msg("Deserialization error for QueueConfig")
		return err
	}
	return nil
}
func (mrc *MessageRouterConfiguration) UnmarshalJSON(data []byte) error {
	type RawConfig struct {
		Metadata json.RawMessage `json:"metadata"`
		Message  json.RawMessage `json:"message"`
	}
	type QConfigRaw struct {
		QueueConfigFields
		Filters []RawConfig `json:"filters"`
	}
	RouterConfigRaw := struct {
		Queues map[string]QConfigRaw `json:"queues"`
	}{Queues: map[string]QConfigRaw{}}
	if err := json.Unmarshal(data, &RouterConfigRaw); err != nil {
		mrc.Logger.Error().Err(err).Msg("Deserialization error for RouterConfigRaw")
		return err
	}

	for name, queue := range RouterConfigRaw.Queues {
		filters := []MqRouterFilterConfiguration{}
		for _, filter := range queue.Filters {
			filterConfig := MqRouterFilterConfiguration{Message: []FilterFieldsConfig{}}
			if string(filter.Metadata[0]) == "[" {
				filterC, err := unmarshalList(filter.Metadata)
				if err != nil {
					mrc.Logger.Error().Err(err).Msg("Deserialization error for filter Metadata list")
					return err
				}
				filterConfig.WasList = true
				filterConfig.Metadata = filterC

			} else {
				if string(filter.Metadata[0]) == "{" {
					filterC, err := unmarshalObj(filter.Metadata)
					if err != nil {
						mrc.Logger.Error().Err(err).Msg("Deserialization error for filter Metadata object")
						return err
					}
					filterConfig.Metadata = filterC
				} else {
					err := fmt.Errorf("wrong object type starting with %v", filter.Metadata[0])
					mrc.Logger.Error().Err(err).Msg("Wrong object type")
					return err
				}
			}
			filters = append(filters, filterConfig)
		}
		mrc.Queues[name] = QueueConfig{Filters: filters, QueueConfigFields: queue.QueueConfigFields}
	}
	mrc.Logger.Debug().Msg("Queue was unmarshalled successfully")
	return nil
}

func unmarshalList(b []byte) ([]FilterFieldsConfig, error) {
	into := []struct {
		FieldName     string `json:"fieldName"`
		ExpectedValue string `json:"expectedValue"`
		Operation     string `json:"operation"`
	}{}
	if err := json.Unmarshal(b, &into); err != nil {
		log.Println("*Err", err)
		return nil, err
	}
	res := []FilterFieldsConfig{}
	for _, f := range into {
		metaFilter := FilterFieldsConfig{}
		metaFilter.FieldName = f.FieldName
		metaFilter.Operation = PickOperation(f.Operation)
		metaFilter.ExpectedValue = f.ExpectedValue

		res = append(res, metaFilter)
	}
	return res, nil
}
func unmarshalObj(b []byte) ([]FilterFieldsConfig, error) {
	type metaFields struct {
		Value     string `json:"value"`
		Operation string `json:"operation"`
	}

	metaDFilter := map[string]metaFields{}
	err := json.Unmarshal(b, &metaDFilter)
	if err != nil {
		return nil, err
	}
	res := []FilterFieldsConfig{}
	for K, f := range metaDFilter {
		metaFilter := FilterFieldsConfig{}
		metaFilter.FieldName = K
		metaFilter.Operation = PickOperation(f.Operation)
		metaFilter.ExpectedValue = f.Value

		res = append(res, metaFilter)
	}

	return res, nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false

}
func (mrc *MessageRouterConfiguration) FindQueuesByAttr(attrs []string) map[string]QueueConfig {
	result := make(map[string]QueueConfig)
	for pin, config := range mrc.Queues {
		var containsAttr []bool
		for _, attr := range attrs {
			containsAttr = append(containsAttr, contains(config.Attributes, attr))
		}
		for i, v := range containsAttr {
			if v == false {
				break
			}
			if i == (len(containsAttr) - 1) {
				result[pin] = config
			}
		}
	}
	mrc.Logger.Debug().Msg("Queue was found")
	return result
}
