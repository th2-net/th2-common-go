/*
 * Copyright 2022-2023 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"github.com/th2-net/th2-common-go/pkg/queue"
)

const (
	SubscribeAttribute = "subscribe"
	SendAttribute      = "publish"
	EventAttribute     = "event"
)

func FindSendQueuesByAttr(mrc *queue.RouterConfig, attrs []string) map[string]queue.DestinationConfig {
	return findQueueByAttr(mrc, attrs, SendAttribute)
}

func FindSubscribeQueuesByAttr(mrc *queue.RouterConfig, attrs []string) map[string]queue.DestinationConfig {
	return findQueueByAttr(mrc, attrs, SubscribeAttribute)
}

func FindSendEventQueuesByAttr(mrc *queue.RouterConfig, attrs []string) map[string]queue.DestinationConfig {
	return findQueueByAttr(mrc, attrs, SendAttribute, EventAttribute)
}

func FindSubscribeEventQueuesByAttr(mrc *queue.RouterConfig, attrs []string) map[string]queue.DestinationConfig {
	return findQueueByAttr(mrc, attrs, SubscribeAttribute, EventAttribute)
}

func findQueueByAttr(mrc *queue.RouterConfig, attrs []string, extraAttr ...string) map[string]queue.DestinationConfig {
	result := make(map[string]queue.DestinationConfig)
	for pin, config := range mrc.Queues {
		match := true
		for _, extra := range extraAttr {
			if !contains(config.Attributes, extra) {
				match = false
				break
			}
		}
		if !match {
			continue
		}
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

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false

}
