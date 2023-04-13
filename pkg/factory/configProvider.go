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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/th2-net/th2-common-go/pkg/common"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
)

var (
	ResourceNotFound = errors.New("resource not found")
)

func newFileProvider(configPath string, extension string, args []string, logger zerolog.Logger) common.ConfigProvider {
	return &fileConfigProvider{
		configurationPath: configPath,
		fileExtension:     extension,
		files:             args,
		zLogger:           &logger,
	}
}

type fileConfigProvider struct {
	configurationPath string
	fileExtension     string
	files             []string
	zLogger           *zerolog.Logger
}

func (cfd *fileConfigProvider) getPath(resourceName string) (string, error) {
	if len(cfd.files) == 0 {
		path := filepath.Join(cfd.configurationPath, fmt.Sprint(resourceName, cfd.fileExtension))
		return path, nil
	}
	for _, filePath := range cfd.files {
		directory, file := filepath.Split(filePath)
		if directory != "" {
			fileName := file
			if strings.Contains(file, ".") {
				fileName = strings.Split(file, ".")[0]
			}
			if fileName == resourceName {
				path, err := filepath.Abs(fmt.Sprint(filePath, cfd.fileExtension))
				if err != nil {
					return "", err
				}
				return path, nil
			}
		} else {
			if filePath == resourceName {
				path := filepath.Join(cfd.configurationPath, fmt.Sprint(resourceName, cfd.fileExtension))
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("%s %w", resourceName, ResourceNotFound)
}

func (cfd *fileConfigProvider) GetConfig(resourceName string, target interface{}) error {
	path, err := cfd.getPath(resourceName)
	if err != nil {
		return err
	}
	cfd.zLogger.Debug().
		Str("resource", resourceName).
		Str("file", path).
		Msg("reading from file")
	fileContentBytes, fileReadErr := os.ReadFile(path)
	if fileReadErr != nil {
		cfd.zLogger.Error().Err(fileReadErr).
			Str("resource", resourceName).
			Str("file", path).
			Msg("file couldn't be read")
		return fileReadErr
	}

	content := string(fileContentBytes)
	cfd.zLogger.Info().
		Str("resource", resourceName).
		Str("file", path).
		Msg(content)
	content = os.ExpandEnv(content)
	err = json.Unmarshal([]byte(content), target)

	if err != nil {
		cfd.zLogger.Error().
			Err(err).
			Str("resource", resourceName).
			Str("file", path).
			Msg("deserialization error")
		return err
	}

	return nil
}
