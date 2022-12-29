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
