/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metrics

import (
	"fmt"
	"reflect"
)

var DEFAULT_BUCKETS = []float64{0.000_25, 0.000_5, 0.001, 0.005, 0.010, 0.015, 0.025, 0.050, 0.100, 0.250, 0.500, 1.0}

const (
	DEFAULT_SESSION_ALIAS_LABEL_NAME = "session_alias"
	DEFAULT_DIRECTION_LABEL_NAME     = "direction"
	DEFAULT_EXCHANGE_LABEL_NAME      = "exchange"
	DEFAULT_ROUTING_KEY_LABEL_NAME   = "routing_key"
	DEFAULT_QUEUE_LABEL_NAME         = "queue"
	DEFAULT_MESSAGE_TYPE_LABEL_NAME  = "message_type"
	DEFAULT_TH2_PIN_LABEL_NAME       = "th2_pin"
	DEFAULT_TH2_TYPE_LABEL_NAME      = "th2_type"
	TH2_TYPE                         = "MESSAGE_GROUP"
	RAW_MESSAGE_TYPE                 = "RAW_MESSAGE"
	MESSAGE_TYPE                     = "MESSAGE"
)

var SENDER_LABELS = []string{
	DEFAULT_TH2_PIN_LABEL_NAME,
	DEFAULT_TH2_TYPE_LABEL_NAME,
	DEFAULT_EXCHANGE_LABEL_NAME,
	DEFAULT_ROUTING_KEY_LABEL_NAME,
}
var SUBSCRIBER_LABELS = []string{
	DEFAULT_TH2_PIN_LABEL_NAME,
	DEFAULT_TH2_TYPE_LABEL_NAME,
	DEFAULT_QUEUE_LABEL_NAME,
}

type HealthMetrics struct {
	LivenessMonitor  *Monitor
	ReadinessMonitor *Monitor
}

func NewHealthMetrics(obj interface{}, liveness *FlagArbiter, readiness *FlagArbiter) *HealthMetrics {
	return &HealthMetrics{
		LivenessMonitor:  liveness.RegisterMonitor(fmt.Sprintf("%v_liveness", reflect.TypeOf(obj).Name())),
		ReadinessMonitor: readiness.RegisterMonitor(fmt.Sprintf("%v_readiness", reflect.TypeOf(obj).Name())),
	}
}

func (hlmetr *HealthMetrics) Enable() {
	hlmetr.LivenessMonitor.Enable()
	hlmetr.ReadinessMonitor.Enable()
}

func (hlmetr *HealthMetrics) Disable() {
	hlmetr.LivenessMonitor.Disable()
	hlmetr.ReadinessMonitor.Disable()
}
