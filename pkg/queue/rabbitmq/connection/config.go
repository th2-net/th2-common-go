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

package connection

type Config struct {
	Host                         string `json:"host"`
	VHost                        string `json:"VHost"`
	Port                         int    `json:"port,string"`
	Username                     string `json:"username"`
	Password                     string `json:"password"`
	ExchangeName                 string `json:"exchangeName"`
	ConnectionTimeout            int    `json:"connectionTimeout,omitempty"`
	ConnectionCloseTimeout       int    `json:"connectionCloseTimeout,omitempty"`
	MaxRecoveryAttempts          int    `json:"maxRecoveryAttempts,omitempty"`
	MinConnectionRecoveryTimeout int    `json:"minConnectionRecoveryTimeout,omitempty"`
	MaxConnectionRecoveryTimeout int    `json:"maxConnectionRecoveryTimeout,omitempty"`
	PrefetchCount                int    `json:"prefetchCount,omitempty"`
	MessageRecursionLimit        int    `json:"messageRecursionLimit,omitempty"`
}
