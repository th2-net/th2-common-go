/*
 * Copyright 2023 Exactpro (Exactpro Systems Limited)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metrics

import (
	"encoding/json"
	"fmt"
	"os"
)

type PrometheusConfiguration struct {
	Host    string `json:"host"`
	Port    string `json:"port"`
	Enabled bool   `json:"enabled"`
}

func (promConfig *PrometheusConfiguration) Init(path string) error {
	content, err := os.ReadFile(path) // Read json file
	if err != nil {
		fmt.Println("Error happened when reading")
		return err
	}
	if err := json.Unmarshal(content, promConfig); err != nil {
		fmt.Println("Error when unmarshaling")
		return err
	}
	return nil
}
