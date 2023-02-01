package metrics

type AggregatingMetricMonitor struct {
	Name      string
	AgrMetric *AggregatingMetric
}

func (agrMetMon *AggregatingMetricMonitor) isEnabled() bool {
	return agrMetMon.AgrMetric.isEnabled()
}

func (agrMetMon *AggregatingMetricMonitor) enable() {
	agrMetMon.AgrMetric.enable()
}

func (agrMetMon *AggregatingMetricMonitor) disable() {
	agrMetMon.AgrMetric.disable()
}
