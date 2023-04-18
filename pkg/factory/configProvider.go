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

package factory

import (
	"encoding/json"
	"errors"
	"github.com/rs/zerolog"
	"github.com/th2-net/th2-common-go/pkg/common"
	"io/fs"
	"os"
)

var (
	ResourceNotFound = errors.New("resource not found")
)

func NewFileProvider(configPath string, extension string, logger zerolog.Logger) common.ConfigProvider {
	return NewFileProviderForFS(
		os.DirFS(configPath),
		extension,
		logger,
	)
}

func NewFileProviderForFS(fs fs.FS, extension string, logger zerolog.Logger) common.ConfigProvider {
	return &fileConfigProvider{
		configFS:      fs,
		fileExtension: extension,
		zLogger:       &logger,
	}
}

type fileConfigProvider struct {
	configFS      fs.FS
	fileExtension string
	zLogger       *zerolog.Logger
}

func (cfd *fileConfigProvider) GetConfig(resourceName string, target any) error {
	stat, err := fs.Stat(cfd.configFS, resourceName+cfd.fileExtension)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ResourceNotFound
		} else {
			return err
		}
	}
	fileName := stat.Name()
	cfd.zLogger.Debug().
		Str("resource", resourceName).
		Str("file", fileName).
		Msg("reading from file")
	fileContentBytes, fileReadErr := fs.ReadFile(cfd.configFS, fileName)
	if fileReadErr != nil {
		cfd.zLogger.Error().Err(fileReadErr).
			Str("resource", resourceName).
			Str("file", fileName).
			Msg("file couldn't be read")
		return fileReadErr
	}

	content := string(fileContentBytes)
	cfd.zLogger.Info().
		Str("resource", resourceName).
		Str("file", fileName).
		Msg(content)
	content = os.ExpandEnv(content)
	err = json.Unmarshal([]byte(content), target)

	if err != nil {
		cfd.zLogger.Error().
			Err(err).
			Str("resource", resourceName).
			Str("file", fileName).
			Msg("deserialization error")
		return err
	}

	return nil
}
