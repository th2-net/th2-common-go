package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetric struct {
	Metric prometheus.Gauge
}

func NewPrometheusMetric(name string, help string) *PrometheusMetric {
	gauge := promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: name,
			Help: help,
		},
	)
	return &PrometheusMetric{
		Metric: gauge,
	}
}

func (promMetric *PrometheusMetric) OnValueChange(value bool) {
	if value {
		promMetric.Metric.Set(1.0)
	} else {
		promMetric.Metric.Set(0.0)
	}
}
