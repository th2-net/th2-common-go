package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type MetricFlag struct {
	Metric prometheus.Gauge

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
		Metric:  gauge,
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
		promMetric.Metric.Set(1.0)
	} else {
		promMetric.Metric.Set(0.0)
	}
}
