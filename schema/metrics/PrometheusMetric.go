package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetric struct {
	Metric prometheus.Gauge

	Enabled bool
}

func NewPrometheusMetric(name string, help string) *PrometheusMetric {
	gauge := promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
	)
	return &PrometheusMetric{
		Metric:  gauge,
		Enabled: false,
	}
}

func (promMetric *PrometheusMetric) IsEnabled() bool {
	return promMetric.Enabled
}

func (promMetric *PrometheusMetric) Enable() {
	if !promMetric.Enabled {
		promMetric.Enabled = true
		promMetric.OnValueChange(promMetric.Enabled)
	}
}

func (promMetric *PrometheusMetric) Disable() {
	if promMetric.Enabled {
		promMetric.Enabled = false
		promMetric.OnValueChange(promMetric.Enabled)
	}
}

func (promMetric *PrometheusMetric) OnValueChange(value bool) {
	if value {
		promMetric.Metric.Set(1.0)
	} else {
		promMetric.Metric.Set(0.0)
	}
}
