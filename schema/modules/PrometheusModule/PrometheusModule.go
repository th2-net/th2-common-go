package PrometheusModule

import (
	"fmt"
	"reflect"

	"github.com/th2-net/th2-common-go/schema/common"
	"github.com/th2-net/th2-common-go/schema/metrics"
)

const (
	PROMETHEUS_MODULE_KEY = "prometheus"
	PROMETHEUS_FILE_NAME  = "prometheus"
)

var prometheusModuleKey = common.ModuleKey(PROMETHEUS_MODULE_KEY)

type PrometheusModule struct {
	prometheus *metrics.PrometheusServer

	livenessArbiter  *metrics.FlagArbiter
	readinessArbiter *metrics.FlagArbiter
}

func (p *PrometheusModule) GetKey() common.ModuleKey {
	return prometheusModuleKey
}

func (p *PrometheusModule) Close() {
	p.prometheus.Stop()
}

func NewPrometheusModule(provider common.ConfigProvider) common.Module {
	promConfig := metrics.PrometheusConfiguration{}
	provider.GetConfig(PROMETHEUS_FILE_NAME, &promConfig)
	serv := metrics.NewPrometheusServer(promConfig.Host, promConfig.Port)

	LivenessArbiter := metrics.NewFlagArbiter(
		metrics.NewMetricFlag("th2_liveness", "Service liveness"),
	)

	ReadinessArbiter := metrics.NewFlagArbiter(
		metrics.NewMetricFlag("th2_readiness", "Service readiness"),
	)

	if promConfig.Enabled {
		serv.Run()
	}

	return &PrometheusModule{
		prometheus:       serv,
		livenessArbiter:  LivenessArbiter,
		readinessArbiter: ReadinessArbiter,
	}
}

type Identity struct{}

func (id *Identity) GetModule(factory common.CommonFactory) (*PrometheusModule, error) {
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
