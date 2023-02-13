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
