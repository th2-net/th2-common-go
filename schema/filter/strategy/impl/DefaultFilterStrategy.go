package impl

import (
	mqFilter "github.com/th2-net/th2-common-go/schema/queue/configuration"
	p_buff "th2-grpc/th2_grpc_common"
)

type DefaultFilterStrategy struct {
	extractFields Th2MsgFieldExtraction
}

func (fs *DefaultFilterStrategy) Construct() {
	fs.extractFields = Th2MsgFieldExtraction{}
}

func (fs *DefaultFilterStrategy) Verify(messages *p_buff.MessageGroup, filters []mqFilter.MqRouterFilterConfiguration) bool {
	// checking
	for _, filter := range filters {
		for _, anyMessage := range messages.Messages {
			if fs.CheckValues(fs.extractFields.GetFields(anyMessage), filter.Metadata.([]mqFilter.FilterFieldsConfiguration)) {
				return true
			}
		}
	}
	return false

}

func (fs *DefaultFilterStrategy) CheckValues(messageFields map[string]string, filters []mqFilter.FilterFieldsConfiguration) bool {
	res := []bool{}
	for _, filter := range filters {
		for _, mField := range messageFields {
			if filter.FieldName == mField {
				res = append(res, checkValue(messageFields[filter.FieldName], filter))
			}
		}
	}
	for _, r := range res {
		if r == true {
			return true
		}
	}
	return false

}

func checkValue(value string, filter mqFilter.FilterFieldsConfiguration) bool {
	expected := filter.ExpectedValue
	switch filter.Operation {
	case mqFilter.EQUAL:
		return value == expected
	case mqFilter.NOT_EQUAL:
		return value != expected
	case mqFilter.EMPTY:
		return len(value) == 0
	case mqFilter.NOT_EMPTY:
		//return fnmatch(v, expected)
	case mqFilter.WILDCARD:
		return value == expected
	default:
		return false
	}
}
