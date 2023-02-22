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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricFlag struct {
	metric prometheus.Gauge

	Enabled bool
}

func NewMetricFlag(name string, help string) *MetricFlag {
	gauge := promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
	)
	return &MetricFlag{
		metric:  gauge,
		Enabled: false,
	}
}

func (promMetric *MetricFlag) IsEnabled() bool {
	return promMetric.Enabled
}

func (promMetric *MetricFlag) Enable() {
	if !promMetric.Enabled {
		promMetric.Enabled = true
		promMetric.OnValueChange(promMetric.Enabled)
	}
}

func (promMetric *MetricFlag) Disable() {
	if promMetric.Enabled {
		promMetric.Enabled = false
		promMetric.OnValueChange(promMetric.Enabled)
	}
}

func (promMetric *MetricFlag) OnValueChange(value bool) {
	if value {
		promMetric.metric.Set(1.0)
	} else {
		promMetric.metric.Set(0.0)
	}
}
