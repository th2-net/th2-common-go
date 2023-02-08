package metrics

import (
	"fmt"
	"reflect"
)

// var LIVENESS_ARBITER = NewAggregatingMetric(
// 	NewPrometheusMetric("th2_liveness", "Service liveness"),
// )
// var READINESS_ARBITER = NewAggregatingMetric(
// 	NewPrometheusMetric("th2_readiness", "Service readiness"),
// )

// func init() {
// 	RegisterLiveness("user_liveness")
// 	RegisterReadiness("user_readiness")
// }

// func RegisterReadiness(name string) *AggregatingMetricMonitor {
// 	return READINESS_ARBITER.CreateMonitor(name)
// }

func RegisterMonitor(name string, agMetric *AggregatingMetric) *AggregatingMetricMonitor {
	return agMetric.CreateMonitor(name)
}

type HealthMetrics struct {
	LivenessMonitor  *AggregatingMetricMonitor
	ReadinessMonitor *AggregatingMetricMonitor
}

func NewHealthMetrics(obj interface{}, liveness *AggregatingMetric, readiness *AggregatingMetric) *HealthMetrics {
	return &HealthMetrics{
		LivenessMonitor:  RegisterMonitor(fmt.Sprintf("%v_liveness", reflect.TypeOf(obj).Name()), liveness),
		ReadinessMonitor: RegisterMonitor(fmt.Sprintf("%v_readiness", reflect.TypeOf(obj).Name()), readiness),
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
