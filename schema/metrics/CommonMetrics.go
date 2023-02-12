package metrics

import (
	"fmt"
	"reflect"
)

type HealthMetrics struct {
	LivenessMonitor  *Monitor
	ReadinessMonitor *Monitor
}

func NewHealthMetrics(obj interface{}, liveness *FlagArbiter, readiness *FlagArbiter) *HealthMetrics {
	return &HealthMetrics{
		LivenessMonitor:  liveness.RegisterMonitor(fmt.Sprintf("%v_liveness", reflect.TypeOf(obj).Name())),
		ReadinessMonitor: readiness.RegisterMonitor(fmt.Sprintf("%v_readiness", reflect.TypeOf(obj).Name())),
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
