package metrics

import (
	"fmt"
	"reflect"
)

var LIVENESS_ARBITER = NewAggregatingMetric(
	NewPrometheusMetric("th2_liveness", "Service liveness"),
)
var READINESS_ARBITER = NewAggregatingMetric(
	NewPrometheusMetric("th2_readiness", "Service readiness"),
)

func init() {
	RegisterLiveness("user_liveness")
	RegisterReadiness("user_readiness")
}

func RegisterLiveness(name string) *AggregatingMetricMonitor {
	return LIVENESS_ARBITER.CreateMonitor(name)
}

func RegisterReadiness(name string) *AggregatingMetricMonitor {
	return READINESS_ARBITER.CreateMonitor(name)
}

type HealthMetrics struct {
	LivenessMonitor  *AggregatingMetricMonitor
	ReadinessMonitor *AggregatingMetricMonitor
}

func NewHealthMetrics(obj interface{}) *HealthMetrics {
	return &HealthMetrics{
		LivenessMonitor:  RegisterLiveness(fmt.Sprintf("%v_liveness", reflect.TypeOf(obj).Name())),
		ReadinessMonitor: RegisterReadiness(fmt.Sprintf("%v_readiness", reflect.TypeOf(obj).Name())),
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
