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
	"github.com/th2-net/th2-common-go/schema/common"
	"github.com/th2-net/th2-common-go/schema/factory"
	"github.com/th2-net/th2-common-go/schema/grpc/config"
	"github.com/th2-net/th2-common-go/schema/grpc/impl"
	"log"
	"reflect"
)

type GrpcModule struct {
	GrpcRouter impl.CommonGrpcRouter
}

func (m *GrpcModule) GetKey() common.ModuleKey {
	return grpcModuleKey
}

//func (m *GrpcModule) Close() {
//	//close router
//}

var grpcModuleKey = common.ModuleKey("grpc")

func NewGrpcModule(provider factory.ConfigProvider) common.Module {

	grpcConfiguration := config.GrpcConfig{}
	err := provider.GetConfig("grpc", &grpcConfiguration)
	if err != nil {
		log.Fatalln(err)
	}
	router := impl.CommonGrpcRouter{Config: grpcConfiguration}
	return &GrpcModule{GrpcRouter: router}
}

type Identity struct{}

func (id *Identity) GetModule(factory *factory.CommonFactory) (*GrpcModule, error) {
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
