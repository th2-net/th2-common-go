/*
 * Copyright 2022 Exactpro (Exactpro Systems Limited)
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package grpcModule

import (
	"fmt"
	"os"
	"reflect"

	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/schema/common"
	"github.com/th2-net/th2-common-go/schema/grpc/config"
	"github.com/th2-net/th2-common-go/schema/grpc/impl"
)

const (
	GRPC_CONFIG_FILENAME = "grpc"
	GRPC_MODULE_KEY      = "grpc"
)

type GrpcModule struct {
	GrpcRouter impl.CommonGrpcRouter
}

func (m *GrpcModule) GetKey() common.ModuleKey {
	return grpcModuleKey
}

func (m *GrpcModule) Close() error {
	m.GrpcRouter.Close()
	return nil
}

var grpcModuleKey = common.ModuleKey(GRPC_MODULE_KEY)

func NewGrpcModule(provider common.ConfigProvider) common.Module {

	grpcConfiguration := config.GrpcConfig{ZLogger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
	err := provider.GetConfig(GRPC_CONFIG_FILENAME, &grpcConfiguration)
	if err != nil {
		grpcConfiguration.ZLogger.Fatal().Err(err).Send()
	}
	router := impl.CommonGrpcRouter{Config: grpcConfiguration,
		ZLogger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
	return &GrpcModule{GrpcRouter: router}
}

type Identity struct{}

func (id *Identity) GetModule(factory common.CommonFactory) (*GrpcModule, error) {
	module, err := factory.Get(grpcModuleKey)
	if err != nil {
		return nil, err
	}
	casted, success := module.(*GrpcModule)
	if !success {
		return nil, fmt.Errorf("module with key %s is a %s", grpcModuleKey, reflect.TypeOf(module))
	}
	return casted, nil
}

var ModuleID = &Identity{}
