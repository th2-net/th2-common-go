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

package log

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/th2-net/th2-common-go/pkg/common"
	"os"
	"strings"
)

var logger zerolog.Logger = zerolog.New(os.Stdout).
	Level(zerolog.InfoLevel).
	With().Timestamp().Logger()

type ZerologConfig struct {
	Level      string `properties:"global_level,default=info"`
	Sampling   bool   `properties:"disable_sampling,default=false"`
	TimeField  string `properties:"time_field,default=time"`
	TimeFormat string `properties:"time_format, default=2006-01-02 15:04:05.000"`
	LevelField string `properties:"level_field, default=level"`
	MsgField   string `properties:"message_field, default=message"`
	ErrorField string `properties:"error_field, default=error"`
}

func ConfigureZerolog(cfg *ZerologConfig) {
	switch level := strings.ToLower(cfg.Level); level {
	case "trace":
		logger = logger.Level(zerolog.TraceLevel)
		break
	case "debug":
		logger = logger.Level(zerolog.DebugLevel)
		break
	case "info":
		logger = logger.Level(zerolog.InfoLevel)
		break
	case "warn":
		logger = logger.Level(zerolog.WarnLevel)
		break
	case "error":
		logger = logger.Level(zerolog.ErrorLevel)
		break
	case "fatal":
		logger = logger.Level(zerolog.FatalLevel)
		break
	default:
		logger = logger.Level(zerolog.InfoLevel)
		log.Warn().Msgf("'%s' log level is unknown. 'INFO' log level is used instead", level)
		break
	}

	zerolog.TimeFieldFormat = cfg.TimeFormat
	zerolog.TimestampFieldName = cfg.TimeField
	zerolog.LevelFieldName = cfg.LevelField
	zerolog.MessageFieldName = cfg.MsgField
	zerolog.ErrorFieldName = cfg.ErrorField
	zerolog.DisableSampling(cfg.Sampling)
}

func ForComponent(name string) zerolog.Logger {
	return logger.With().Str(common.ComponentLoggerKey, name).Logger()
}

func Global() *zerolog.Logger {
	return &logger
}
