/*
 * Copyright 2022-2023 Exactpro (Exactpro Systems Limited)
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

package common

import (
	"io"

	"github.com/rs/zerolog"
)

const ComponentLoggerKey = "component"

type BoxConfig struct {
	Name string `json:"boxName"`
	Book string `json:"bookName"`
}

type Module interface {
	GetKey() ModuleKey
	io.Closer
}

type ModuleKey string

type ConfigProvider interface {
	GetBoxConfig() BoxConfig
	GetConfig(resourceName string, target any) error
}

type Factory interface {
	GetBoxConfig() BoxConfig
	Register(factories ...func(ConfigProvider) (Module, error)) error
	Get(key ModuleKey) (Module, error)
	GetLogger(name string) zerolog.Logger
	GetCustomConfiguration(any any) error
	io.Closer
}
