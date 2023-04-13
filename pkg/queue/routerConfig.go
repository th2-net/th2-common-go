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

package queue

type RouterConfig struct {
	Queues map[string]DestinationConfig `json:"queues"`
}

type DestinationConfig struct {
	RoutingKey string                        `json:"name"`
	QueueName  string                        `json:"queue"`
	Exchange   string                        `json:"exchange"`
	Attributes []string                      `json:"attributes"`
	Filters    []MqRouterFilterConfiguration `json:"filters"`
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false

}

func (mrc *RouterConfig) FindQueuesByAttr(attrs []string) map[string]DestinationConfig {
	result := make(map[string]DestinationConfig)
	for pin, config := range mrc.Queues {
		match := true
		for _, attr := range attrs {
			if !contains(config.Attributes, attr) {
				match = false
				break
			}
		}
		if match {
			result[pin] = config
		}
	}
	return result
}
