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
	"log"
	"os"
	"path/filepath"
	"strings"
)

type ConfigProvider interface {
	GetConfig(resourceName string, target interface{}) error
}

type ConfigProviderFromFile struct {
	configurationPath string
	fileExtension     string
	files             []string
}

func (cfd *ConfigProviderFromFile) getPath(resourceName string) string {
	if len(cfd.files) == 0 {
		path := filepath.Join(cfd.configurationPath, fmt.Sprint(resourceName, cfd.fileExtension))
		return path
	} else {
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
						log.Fatalln(err)
					}
					return path
				}
			} else {
				if filePath == resourceName {
					path := filepath.Join(cfd.configurationPath, fmt.Sprint(resourceName, cfd.fileExtension))
					return path
				}
			}
		}
	}
	return ""
}

func (cfd *ConfigProviderFromFile) GetConfig(resourceName string, target interface{}) error {

	path := cfd.getPath(resourceName)
	fileContentBytes, fileReadErr := os.ReadFile(path)
	if fileReadErr != nil {
		return errors.New(fmt.Sprintf("file with path %s couldn't be read", path))
	}

	err := json.Unmarshal(fileContentBytes, target)

	if err != nil {
		return errors.New(fmt.Sprintf("deserialization error: %s ", err.Error()))
	}

	return nil
}