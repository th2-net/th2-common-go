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

package factory

import (
	"errors"
	"fmt"
	"github.com/th2-net/th2-common-go/schema/common"
	"reflect"
)

const (
	sourceDirectory = "../resources"
	jsonExtension   = ".json"
)

type CommonFactory struct {
	modules     map[common.ModuleKey]common.Module
	cfgProvider ConfigProvider
}

func newProvider(args []string) ConfigProvider {
	return &ConfigProviderFromFile{configurationPath: sourceDirectory, fileExtension: jsonExtension, files: args}
}

func NewFactory(args []string) *CommonFactory {
	provider := newProvider(args[1:])
	return &CommonFactory{
		modules:     make(map[common.ModuleKey]common.Module),
		cfgProvider: provider,
	}
}

func (cf *CommonFactory) Register(factories ...func(ConfigProvider) common.Module) error {
	for _, factory := range factories {
		module := factory(cf.cfgProvider)
		if oldModule, exist := cf.modules[module.GetKey()]; exist {
			return fmt.Errorf("module %s with key %s already registered", reflect.TypeOf(oldModule), module.GetKey())
		}
		cf.modules[module.GetKey()] = module
	}
	return nil
}

func (cf *CommonFactory) Get(key common.ModuleKey) (common.Module, error) {
	if module, exist := cf.modules[key]; !exist {
		return nil, errors.New("module " + string(key) + " does not exist")
	} else {
		return module, nil
	}
}
