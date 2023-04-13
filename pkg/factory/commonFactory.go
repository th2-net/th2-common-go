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

	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/magiconair/properties"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/th2-net/th2-common-go/pkg/common"
	"github.com/th2-net/th2-common-go/pkg/modules/prometheus"
)

const (
	configurationPath = "/var/th2/config/"
	jsonExtension     = ".json"
	customFileName    = "custom"

	ComponentLoggerKey = "component"
)

type Config struct {
	ConfigurationsDir string
	FileExtension     string
}

type commonFactory struct {
	modules     map[common.ModuleKey]common.Module
	cfgProvider common.ConfigProvider
	zLogger     zerolog.Logger
	boxConfig   common.BoxConfig
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

func New() common.Factory {
	configPath := flag.String("config-file-path", configurationPath, "pass path to config files")
	extension := flag.String("config-file-extension", jsonExtension, "file extension")
	flag.Parse()
	factory, err := NewFromConfig(Config{
		ConfigurationsDir: *configPath,
		FileExtension:     *extension,
	})
	if err != nil {
		panic(err)
	}
	return factory
}

func NewFromConfig(config Config) (common.Factory, error) {
	if config.ConfigurationsDir == "" {
		return nil, fmt.Errorf("configuration directory is empty")
	}
	if config.FileExtension == "" {
		return nil, fmt.Errorf("configurations file extension is empty")
	}
	loadZeroLogConfig(config)

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	provider := NewFileProvider(
		config.ConfigurationsDir,
		config.FileExtension,
		logger.With().Str(ComponentLoggerKey, "file_provider").Logger(),
	)
	var boxConfig common.BoxConfig
	if err := provider.GetConfig("box", &boxConfig); err != nil {
		log.Warn().Err(err).Msg("cannot read box configuration")
	}
	cf := &commonFactory{
		modules:     make(map[common.ModuleKey]common.Module),
		cfgProvider: provider,
		zLogger:     logger,
		boxConfig:   boxConfig,
	}
	err := cf.Register(prometheus.NewModule)
	if err != nil {
		return nil, err
	}

	return cf, nil
}

func loadZeroLogConfig(config Config) {
	var cfg ZerologConfig
	p, pErr := properties.LoadFile(filepath.Join(config.ConfigurationsDir, "zerolog.properties"), properties.UTF8)
	if pErr != nil {
		log.Error().Err(pErr).Msg("Can't get properties for zerolog")
		return
	}
	if err := p.Decode(&cfg); err != nil {
		log.Error().Err(pErr).Msg("Can't decode properties into zerolog configuration structure")
		return
	}
	log.Info().Msg("Loggers will be configured via zerolog.properties file")

	configureZerolog(&cfg)
}

func (cf *commonFactory) Register(factories ...func(common.ConfigProvider) (common.Module, error)) error {
	for _, factory := range factories {
		module, err := factory(cf.cfgProvider)
		if err != nil {
			return err
		}
		if oldModule, exist := cf.modules[module.GetKey()]; exist {
			return fmt.Errorf("module %s with key %s already registered", reflect.TypeOf(oldModule), module.GetKey())
		}
		cf.modules[module.GetKey()] = module
		cf.zLogger.Info().Msgf("Registered new %v module", module.GetKey())
	}
	return nil
}

func (cf *commonFactory) Get(key common.ModuleKey) (common.Module, error) {
	if module, exist := cf.modules[key]; !exist {
		return nil, errors.New("module " + string(key) + " does not exist")
	} else {
		return module, nil
	}
}

func (cf *commonFactory) GetLogger(name string) zerolog.Logger {
	return cf.zLogger.With().Str("component", name).Logger()
}

func (cf *commonFactory) Close() error {
	var err error
	for moduleKey, module := range cf.modules {
		if err = module.Close(); err == nil {
			cf.zLogger.Info().Msgf("Module %v closed", moduleKey)
		} else {
			cf.zLogger.Error().Err(err).Msgf("Module %v raised error", moduleKey)
		}
	}
	return nil
}

func (cf *commonFactory) GetCustomConfiguration(any any) error {
	return cf.cfgProvider.GetConfig(customFileName, any)
}

func (cf *commonFactory) GetBoxConfig() common.BoxConfig {
	return cf.boxConfig
}
