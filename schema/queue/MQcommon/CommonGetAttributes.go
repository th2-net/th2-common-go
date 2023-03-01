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
package MQcommon

import "strings"

func GetSendAttributes(attrs []string) []string {
	res := []string{}
	if len(attrs) == 0 {
		return res
	} else {
		attrMap := make(map[string]bool)
		for _, attr := range attrs {
			attrMap[strings.ToLower(attr)] = true
		}
		attrMap["publish"] = true
		for k, _ := range attrMap {
			res = append(res, k)
		}
		return res
	}
}

func GetSubscribeAttributes(attrs []string) []string {
	res := []string{}
	if len(attrs) == 0 {
		return res
	} else {
		attrMap := make(map[string]bool)
		for _, attr := range attrs {
			attrMap[strings.ToLower(attr)] = true
		}
		attrMap["subscribe"] = true
		for k, _ := range attrMap {
			res = append(res, k)
		}
		return res
	}
}
