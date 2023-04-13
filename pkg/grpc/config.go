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

package grpc

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
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

type Server struct {
	Endpoint
	Workers int `json:"workers"`
}
type Strategy struct {
	Endpoints []string `json:"endpoints"`
	Name      string   `json:"name"`
}

type Service struct {
	Endpoints    map[string]Endpoint `json:"endpoints"`
	ServiceClass string              `json:"service-class"`
	Strategy     `json:"strategy"`
}

type Services map[string]Service

type Config struct {
	ServerConfig Server   `json:"server"`
	ServicesMap  Services `json:"services"`
	ZLogger      zerolog.Logger
}

func (gc *Config) ValidatePins() error {
	for pinName, service := range gc.ServicesMap {
		if len(service.Endpoints) > 1 {
			return errors.New(fmt.Sprintf(
				`config is invalid. pin "%s" has more than 1 endpoint`, pinName))
		}
		gc.ZLogger.Info().Msg("Pins validated.")
	}
	return nil
}

func (gc *Config) findEndpointAddrViaServiceName(srvcName string) (Address, error) {
	for _, service := range gc.ServicesMap {
		serviceName := service.ServiceClass
		if strings.Contains(serviceName, ".") {
			serviceNameList := strings.Split(serviceName, ".")
			index := len(serviceNameList)
			serviceName = serviceNameList[index-1]
		}
		if serviceName == srvcName {
			if len(service.Endpoints) > 1 {
				return Address{}, fmt.Errorf("number of endpoints should equal to 1")
			} else {
				for _, endpoint := range service.Endpoints {
					gc.ZLogger.Debug().Msg("Endpoint was found")
					return endpoint.Address, nil
				}
			}
		}
	}
	gc.ZLogger.Error().Str("service", srvcName).Msg("No endpoint exists")
	return Address{}, errors.New("endpoint with provided service name does not exist")
}

func (gc *Config) getServerAddress() string {
	return gc.ServerConfig.AsColonSeparatedString()
}
