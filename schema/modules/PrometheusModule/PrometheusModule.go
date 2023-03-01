/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

	LivenessArbiter  *metrics.FlagArbiter
	ReadinessArbiter *metrics.FlagArbiter
}

func (p *PrometheusModule) GetKey() common.ModuleKey {
	return prometheusModuleKey
}

func (p *PrometheusModule) Close() error {
	p.prometheus.Stop()
	return nil
}

func NewPrometheusModule(provider common.ConfigProvider) common.Module {
	promConfig := metrics.PrometheusConfiguration{Host: "0.0.0.0", Port: "9752"}
	provider.GetConfig(PROMETHEUS_FILE_NAME, &promConfig)
	serv := metrics.NewPrometheusServer(promConfig.Host, promConfig.Port)

	LivenessArbiter := metrics.NewFlagArbiter(
		metrics.NewMetricFlag("th2_liveness", "Service liveness"),
		metrics.NewFileFlag("healthy"),
	)

	ReadinessArbiter := metrics.NewFlagArbiter(
		metrics.NewMetricFlag("th2_readiness", "Service readiness"),
		metrics.NewFileFlag("ready"),
	)

	if promConfig.Enabled {
		serv.Run()
	}

	return &PrometheusModule{
		prometheus:       serv,
		LivenessArbiter:  LivenessArbiter,
		ReadinessArbiter: ReadinessArbiter,
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
