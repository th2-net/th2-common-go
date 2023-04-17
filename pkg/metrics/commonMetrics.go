/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
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

package metrics

import (
	"fmt"
	"reflect"
)

var DefaultBuckets = []float64{0.000_25, 0.000_5, 0.001, 0.005, 0.010, 0.015, 0.025, 0.050, 0.100, 0.250, 0.500, 1.0}

const (
	DefaultSessionAliasLabelName = "session_alias"
	DefaultDirectionLabelName    = "direction"
	DefaultExchangeLabelName     = "exchange"
	DefaultRoutingKeyLabelName   = "routing_key"
	DefaultQueueLabelName        = "queue"
	DefaultMessageTypeLabelName  = "message_type"
	DefaultTh2PinLabelName       = "th2_pin"
	DefaultTh2TypeLabelName      = "th2_type"
	MessageGroupTh2Type          = "MESSAGE_GROUP"
	EventTh2Type                 = "EVENT"
	RawMessageType               = "RAW_MESSAGE"
	ParsedMessageType            = "MESSAGE"
)

var SenderLabels = []string{
	DefaultTh2PinLabelName,
	DefaultTh2TypeLabelName,
	DefaultExchangeLabelName,
	DefaultRoutingKeyLabelName,
}
var SubscriberLabels = []string{
	DefaultTh2PinLabelName,
	DefaultTh2TypeLabelName,
	DefaultQueueLabelName,
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
