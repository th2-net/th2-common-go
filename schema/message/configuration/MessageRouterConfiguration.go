package th2

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/rs/zerolog"
)

type MessageRouterConfiguration struct {
	Queues map[string]QueueConfiguration `json:"queues"`

	Logger zerolog.Logger
}

func (rc *MessageRouterConfiguration) Init(jsonStr string) error {

	reader := strings.NewReader(jsonStr)

	dec := json.NewDecoder(reader)

	dec.DisallowUnknownFields()

	err := dec.Decode(&rc)

	if err != nil {
		rc.Logger.Error().Err(err).Msg("message router config deserialization error")
		return err
	}

	return nil
}

func (rc *MessageRouterConfiguration) GetQueueByAlias(queueAlias string) (*QueueConfiguration, error) {

	if queue, present := rc.Queues[queueAlias]; present {
		return &queue, nil
	}

	return nil, errors.New("queue couldn't be found")
}

func (rc *MessageRouterConfiguration) GetQueues() *QueuesMap {
	return &rc.Queues
}

func (rc *MessageRouterConfiguration) FindQueuesByAttr(attr QueueAttribute) *QueuesMap {

	queuesMap := QueuesMap{}

	for name, queue := range rc.Queues {

		if queue.ContainsAllAttributes(attr) {
			queuesMap[name] = queue
		}
	}

	return &queuesMap
}

type MqRouterFilterConfiguration struct {
	Metadata []FieldFilterConfiguration `json:"metadata"`
	Message  []FieldFilterConfiguration `json:"message"`
}

type FieldFilterConfiguration struct {
	FieldName string `json:"fieldName"`
	Value     string `json:"value"`
	Operation string `json:"operation"`
}

type FieldFilterOperation int

const (
	EQUAL     FieldFilterOperation = iota // 0
	NOT_EQUAL                             // 1
	EMPTY                                 // 2
	NOT_EMPTY                             // 3
	WILDCARD                              // 4
)
