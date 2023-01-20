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
	"flag"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/th2-net/th2-common-go/schema/common"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	configurationPath = "/var/th2/config/"
	jsonExtension     = ".json"
)

type CommonFactory struct {
	modules     map[common.ModuleKey]common.Module
	cfgProvider ConfigProvider
	zLogger     zerolog.Logger
}

type ZerologConfig struct {
	Level      string `properties:"global_level,default=info"`
	Sampling   bool   `properties:"disable_sampling,default=false"`
	TimeField  string `properties:"time_field,default=time"`
	TimeFormat string `properties:"time_format, default=2006-01-02 15:04:05.000"`
	LevelField string `properties:"level_field, default=level"`
	MsgField   string `properties:"message_field, default=message"`
	ErrorField string `properties:"error_field, default=error"`
}

func configureZerolog(cfg *ZerologConfig) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if strings.ToLower(cfg.Level) == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else if strings.ToLower(cfg.Level) == "warn" {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	} else if strings.ToLower(cfg.Level) == "error" {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	} else if strings.ToLower(cfg.Level) == "fatal" {
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	}
	zerolog.TimeFieldFormat = cfg.TimeFormat
	zerolog.TimestampFieldName = cfg.TimeField
	zerolog.LevelFieldName = cfg.LevelField
	zerolog.MessageFieldName = cfg.MsgField
	zerolog.ErrorFieldName = cfg.ErrorField
	zerolog.DisableSampling(cfg.Sampling)

}

func newProvider(configPath string, extension string, args []string) ConfigProvider {
	return &ConfigProviderFromFile{configurationPath: configPath, fileExtension: extension,
		files: args, zLogger: zerolog.New(os.Stdout).With().Timestamp().Logger()}
}

func NewFactory(args ...string) *CommonFactory {
	configPath := flag.String("config-file-path", configurationPath, "pass path tp=o config file")
	extension := flag.String("config-file-extension", jsonExtension, "file extension")
	flag.Parse()
	p, pErr := properties.LoadFile(filepath.Join(*configPath, "zerolog.properties"), properties.UTF8)
	if pErr != nil {
		log.Info().Err(pErr).Msg("Can't get properties for zerolog")
	}
	var cfg ZerologConfig
	if err := p.Decode(&cfg); err != nil {
		log.Info().Err(pErr).Msg("Can't decode properties into zerolog configuration structure")
	}

	configureZerolog(&cfg)
	provider := newProvider(*configPath, *extension, args)
	return &CommonFactory{
		modules:     make(map[common.ModuleKey]common.Module),
		cfgProvider: provider,
		zLogger:     zerolog.New(os.Stdout).With().Timestamp().Logger(),
	}
}

func (cf *CommonFactory) Register(factories ...func(ConfigProvider) common.Module) error {
	for _, factory := range factories {
		module := factory(cf.cfgProvider)
		if oldModule, exist := cf.modules[module.GetKey()]; exist {
			return fmt.Errorf("module %s with key %s already registered", reflect.TypeOf(oldModule), module.GetKey())
		}
		cf.modules[module.GetKey()] = module
		cf.zLogger.Info().Msgf("Registered new %v module", module.GetKey())
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

func (cf *CommonFactory) Close() {
	for moduleKey, module := range cf.modules {
		module.Close()
		cf.zLogger.Info().Msgf("Module %v closed. \n", moduleKey)
	}
}

func (cf *CommonFactory) GetCustomConfiguration(any interface{}) error {
	err := cf.cfgProvider.GetConfig("custom", any)
	return err
}
