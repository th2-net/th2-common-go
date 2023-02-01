package metrics

type AggregatingMetric struct {
	metrics []*Metric
}

func (agMetric *AggregatingMetric) CreateMonitor(name string) *AggregatingMetricMonitor {
	return &AggregatingMetricMonitor{
		Name:      name,
		AgrMetric: agMetric,
	}
}

func (agMetric *AggregatingMetric) isEnabled() bool {
	for _, metric := range agMetric.metrics {
		if !(*metric).isEnabled() {
			return false
		}
	}
	return true
}

func (agMetric *AggregatingMetric) enable() {
	for _, metric := range agMetric.metrics {
		(*metric).enable()
	}
}

func (agMetric *AggregatingMetric) disable() {
	for _, metric := range agMetric.metrics {
		(*metric).disable()
	}
}
