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
	"github.com/rs/zerolog"
	"os"
)

type MqRouterFilterConfiguration struct{}

type QueueConfig struct {
	RoutingKey string                        `json:"name"`
	QueueName  string                        `json:"queue"`
	Exchange   string                        `json:"exchange"`
	Attributes []string                      `json:"attributes"`
	Filters    []MqRouterFilterConfiguration `json:"filters"`
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
	fail := json.Unmarshal(content, mrc)
	if fail != nil {
		mrc.Logger.Error().Err(err).Msg("Deserialization error for QueueConfig")
		return err
	}
	return nil
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
