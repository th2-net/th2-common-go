package metrics

type AggregatingMetric struct {
	metrics []Metric
}

func NewAggregatingMetric(metrics ...Metric) *AggregatingMetric {
	return &AggregatingMetric{
		metrics: metrics,
	}
}

func (agMetric *AggregatingMetric) CreateMonitor(name string) *AggregatingMetricMonitor {
	return &AggregatingMetricMonitor{
		Name:      name,
		AgrMetric: agMetric,
	}
}

func (agMetric *AggregatingMetric) IsEnabled() bool {
	for _, metric := range agMetric.metrics {
		if !metric.IsEnabled() {
			return false
		}
	}
	return true
}

func (agMetric *AggregatingMetric) Enable() {
	for _, metric := range agMetric.metrics {
		metric.Enable()
	}
}

func (agMetric *AggregatingMetric) Disable() {
	for _, metric := range agMetric.metrics {
		metric.Disable()
	}
}
