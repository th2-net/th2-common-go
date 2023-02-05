package metrics

type AggregatingMetricMonitor struct {
	Name      string
	AgrMetric *AggregatingMetric
}

func (agrMetMon *AggregatingMetricMonitor) IsEnabled() bool {
	return agrMetMon.AgrMetric.IsEnabled()
}

func (agrMetMon *AggregatingMetricMonitor) Enable() {
	agrMetMon.AgrMetric.Enable()
}

func (agrMetMon *AggregatingMetricMonitor) Disable() {
	agrMetMon.AgrMetric.Disable()
}
