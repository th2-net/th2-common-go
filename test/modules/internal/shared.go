/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
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

package internal

import (
	"encoding/json"
	"errors"
	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/pkg/common"
	"io/fs"
	"os"
)

func CreateTestFactory(fileSystem fs.FS) common.Factory {
	return &dummyFactory{
		store:    make(map[common.ModuleKey]common.Module),
		provider: testProvider{fs: fileSystem},
	}
}

type testProvider struct {
	fs fs.FS
}

func (p testProvider) GetConfig(resource string, target any) error {
	data, err := fs.ReadFile(p.fs, resource)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

type dummyFactory struct {
	provider common.ConfigProvider
	store    map[common.ModuleKey]common.Module
}

func (d *dummyFactory) GetBoxConfig() common.BoxConfig {
	return common.BoxConfig{}
}

func (d *dummyFactory) Register(factories ...func(common.ConfigProvider) (common.Module, error)) error {
	for _, f := range factories {
		mod, err := f(d.provider)
		if err != nil {
			return err
		}
		_, exist := d.store[mod.GetKey()]
		if exist {
			return errors.New("module exists")
		}
		d.store[mod.GetKey()] = mod
	}
	return nil
}

func (d *dummyFactory) Get(key common.ModuleKey) (common.Module, error) {
	return d.store[key], nil
}

func (d *dummyFactory) GetLogger(name string) zerolog.Logger {
	return zerolog.New(os.Stdout).With().Str("name", name).Logger()
}

func (d *dummyFactory) GetCustomConfiguration(any any) error {
	return nil
}

func (d *dummyFactory) Close() error {
	return nil
}
