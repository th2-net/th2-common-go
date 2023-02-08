package factory

import (
	"fmt"
	"reflect"

	"github.com/th2-net/th2-common-go/schema/common"
	"github.com/th2-net/th2-common-go/schema/metrics"
)

const (
	PROMETHEUS_MODULE_KEY = "prometheus"
)

var prometheusModuleKey = common.ModuleKey(PROMETHEUS_MODULE_KEY)

type PrometheusModule struct {
	livenessMonitor *metrics.AggregatingMetricMonitor
	prometheus      *metrics.PrometheusServer
	livenessArbiter *metrics.AggregatingMetric
}

func (p *PrometheusModule) GetKey() common.ModuleKey {
	return prometheusModuleKey
}

func (p *PrometheusModule) Close() {
	p.prometheus.Stop()
}

func NewPrometheusModule(provider ConfigProvider) common.Module {
	promConfig := metrics.PrometheusConfiguration{}
	provider.GetConfig(PROMETHEUS_FILE_NAME, &promConfig)
	serv := metrics.NewPrometheusServer(promConfig.Host, promConfig.Port)

	LivenessArbiter := metrics.NewAggregatingMetric(
		metrics.NewPrometheusMetric("th2_liveness", "Service liveness"),
	)

	livenessMonitor := metrics.RegisterMonitor("common_factory_liveness", LivenessArbiter)
	if promConfig.Enabled {
		serv.Run()
		livenessMonitor.Enable()
	}

	return &PrometheusModule{
		prometheus:      serv,
		livenessMonitor: livenessMonitor,
		livenessArbiter: LivenessArbiter,
	}
}

func (p *PrometheusModule) RegisterMonitor(name string) *metrics.AggregatingMetricMonitor {
	return metrics.RegisterMonitor(name, p.livenessArbiter)
}

type Identity struct{}

func (id *Identity) GetModule(factory *CommonFactory) (*PrometheusModule, error) {
	module, err := factory.Get(prometheusModuleKey)
	if err != nil {
		return nil, err
	}
	casted, success := module.(*PrometheusModule)
	if !success {
		return nil, fmt.Errorf("module with key %s is a %s", prometheusModuleKey, reflect.TypeOf(module))
	}
	return casted, nil
}

var ModuleID = &Identity{}
