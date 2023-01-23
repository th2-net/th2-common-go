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

package config

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Address struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (addr *Address) AsColonSeparatedString() string {
	return fmt.Sprint(addr.Host, ":", addr.Port)
}

type Endpoint struct {
	Attributes []string `json:"attributes"`
	Address
}

type Service struct {
	Endpoints map[string]Endpoint `json:"endpoints"`
	//others if needed..
}

type Services map[string]Service

type GrpcConfig struct {
	ServerConfig Address  `json:"server"`
	ServicesMap  Services `json:"services"`
}

func (gc *GrpcConfig) ValidatePins() error {
	for pinName, service := range gc.ServicesMap {
		if len(service.Endpoints) > 1 {
			return errors.New(fmt.Sprintf(
				`config is invalid. pin "%s" has more than 1 endpoint`, pinName))
		}
	}
	return nil
}

// Checks for the inclusion of target attributes
func (gc *GrpcConfig) FindEndpointAddrViaAttributes(targetAttributes []string) (Address, error) {
	targetCopy := make([]string, len(targetAttributes))
	copy(targetCopy, targetAttributes)
	sort.Strings(targetCopy)
	separator := ", "
	sortedTargetAttrsStr := strings.Join(targetCopy, separator)

	for _, service := range gc.ServicesMap {
		for _, endpoint := range service.Endpoints {
			endpointAttrs := endpoint.Attributes
			endpointAttrsCopy := make([]string, len(endpointAttrs))
			copy(endpointAttrsCopy, endpointAttrs)
			sort.Strings(endpointAttrsCopy)
			sortedEndpointAttrsStr := strings.Join(endpointAttrsCopy, separator)
			if strings.Contains(sortedEndpointAttrsStr, sortedTargetAttrsStr) {
				return endpoint.Address, nil
			}
		}
	}
	return Address{}, errors.New("endpoint with provided attributes does not exist")
}

func (gc *GrpcConfig) GetServerAddress() string {
	return gc.ServerConfig.AsColonSeparatedString()
}
