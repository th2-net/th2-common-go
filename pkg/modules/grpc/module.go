/*
 * Copyright 2022-2023 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package grpc

import (
	"fmt"
	"github.com/th2-net/th2-common-go/pkg/grpc"
	"os"
	"reflect"

	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/pkg/common"
)

const (
	configFilename = "grpc"
	moduleKey      = "grpc"
)

type Module interface {
	common.Module
	GetRouter() grpc.Router
}

type impl struct {
	router grpc.Router
}

func (m *impl) GetRouter() grpc.Router {
	return m.router
}

func (m *impl) GetKey() common.ModuleKey {
	return grpcModuleKey
}

func (m *impl) Close() error {
	return m.router.Close()
}

var grpcModuleKey = common.ModuleKey(moduleKey)

func NewModule(provider common.ConfigProvider) (common.Module, error) {

	grpcConfiguration := grpc.Config{ZLogger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
	err := provider.GetConfig(configFilename, &grpcConfiguration)
	if err != nil {
		return nil, err
	}
	return New(grpcConfiguration)
}

func New(config grpc.Config) (Module, error) {
	router := grpc.NewRouter(
		config,
		zerolog.New(os.Stdout).With().Timestamp().Logger(),
	)
	return &impl{router: router}, nil
}

type Identity struct{}

func (id *Identity) GetModule(factory common.Factory) (Module, error) {
	module, err := factory.Get(grpcModuleKey)
	if err != nil {
		return nil, err
	}
	casted, success := module.(Module)
	if !success {
		return nil, fmt.Errorf("module with key %s is a %s", grpcModuleKey, reflect.TypeOf(module))
	}
	return casted, nil
}

var ModuleID = &Identity{}
