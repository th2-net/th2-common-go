package th2

import (
	"github.com/rs/zerolog"
)

type QueueAttribute = []string
type QueuesMap = map[string]QueueConfiguration

type QueueConfiguration struct {
	RoutingKey string                        `json:"routingKey"`
	Queue      string                        `json:"queue"`
	Exchange   string                        `json:"exchange"`
	Attributes QueueAttribute                `json:"attributes"`
	Filters    []MqRouterFilterConfiguration `json:"filters"`
	isReadable bool
	isWritable bool

	logger zerolog.Logger
}

func (qc *QueueConfiguration) GetFilters() *[]MqRouterFilterConfiguration {
	return &qc.Filters
}

func (qc *QueueConfiguration) GetExchange() string {
	return qc.Exchange
}

func (qc *QueueConfiguration) GetRoutingKey() string {
	return qc.RoutingKey
}

func (qc *QueueConfiguration) ContainsAllAttributes(attrs QueueAttribute) bool {

	if len(attrs) != len(qc.Attributes) {
		return false
	}

	var res bool
	for _, attr := range attrs {

		res = false
		for _, q_attr := range qc.Attributes {
			if q_attr == attr {
				res = true
				break
			}
		}

		if !res {
			return false
		}

	}
	return res
}

func (qc *QueueConfiguration) IsReadable() bool {
	return qc.isReadable
}

func (qc *QueueConfiguration) IsWritable() bool {
	return qc.isWritable
}
