/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package prometheus

import (
	"errors"
	"fmt"
	"github.com/th2-net/th2-common-go/pkg/metrics/prometheus"
	"reflect"

	"github.com/th2-net/th2-common-go/pkg/common"
	"github.com/th2-net/th2-common-go/pkg/metrics"
)

const (
	moduleKey      = "prometheus"
	configFileName = "prometheus"
)

var prometheusModuleKey = common.ModuleKey(moduleKey)

type Module interface {
	common.Module
	GetLivenessArbiter() *metrics.FlagArbiter
	GetReadinessArbiter() *metrics.FlagArbiter
}

type module struct {
	prometheus *prometheus.Server

	livenessArbiter  *metrics.FlagArbiter
	readinessArbiter *metrics.FlagArbiter
}

func (p *module) GetLivenessArbiter() *metrics.FlagArbiter {
	return p.livenessArbiter
}

func (p *module) GetReadinessArbiter() *metrics.FlagArbiter {
	return p.readinessArbiter
}

func (p *module) GetKey() common.ModuleKey {
	return prometheusModuleKey
}

func (p *module) Close() error {
	return p.prometheus.Stop()
}

func NewModule(provider common.ConfigProvider) (common.Module, error) {
	promConfig := prometheus.Configuration{Host: "0.0.0.0", Port: 9752}
	err := provider.GetConfig(configFileName, &promConfig)
	if err != nil {
		return nil, err
	}
	return New(promConfig)
}

func New(config prometheus.Configuration) (Module, error) {
	if config.Port == 0 {
		return nil, errors.New("port is not set")
	}
	if config.Host == "" {
		return nil, errors.New("host is not set")
	}

	serv := prometheus.NewServer(config.Host, config.Port)

	livenessArbiter := metrics.NewFlagArbiter(
		metrics.NewMetricFlag("th2_liveness", "Service liveness"),
		metrics.NewFileFlag("healthy"),
	)

	readinessArbiter := metrics.NewFlagArbiter(
		metrics.NewMetricFlag("th2_readiness", "Service readiness"),
		metrics.NewFileFlag("ready"),
	)

	if config.Enabled {
		serv.Run()
	}

	return &module{
		prometheus:       serv,
		livenessArbiter:  livenessArbiter,
		readinessArbiter: readinessArbiter,
	}, nil
}

type Identity struct{}

func (id *Identity) GetModule(factory common.Factory) (Module, error) {
	module, err := factory.Get(prometheusModuleKey)
	if err != nil {
		return nil, err
	}
	casted, success := module.(Module)
	if !success {
		return nil, fmt.Errorf("module with key %s is a %s", prometheusModuleKey, reflect.TypeOf(module))
	}
	return casted, nil
}

var ModuleID = &Identity{}
