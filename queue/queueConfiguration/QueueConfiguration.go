package queueConfiguration

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"log"
	"os"
)

type MqRouterFilterConfiguration struct{}

type QueueConfig struct {
	RoutingKey string                        `json:"name"`
	QueueName  string                        `json:"queue"`
	Exchange   string                        `json:"exchange"`
	Attributes []string                      `json:"attributes"`
	Filters    []MqRouterFilterConfiguration `json:"filters"`
}

type MessageRouterConfiguration struct {
	Queues map[string]QueueConfig `json:"queues"`

	Logger zerolog.Logger
}

func (mrc *MessageRouterConfiguration) Init(path string) error {
	content, err := os.ReadFile(path) // Read json file
	if err != nil {
		log.Println("json file reading error: ")
		return err
	}
	fail := json.Unmarshal(content, mrc)
	if fail != nil {
		log.Printf("Deserialization error %v \n", fail)
		return err
	}
	return nil
}
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
func (mrc *MessageRouterConfiguration) FindQueuesByAttr(attrs []string) map[string]QueueConfig {
	result := make(map[string]QueueConfig)
	for alias, config := range mrc.Queues {
		var containsAttr []bool
		for _, attr := range attrs {
			containsAttr = append(containsAttr, contains(config.Attributes, attr))
		}
		for i, v := range containsAttr {
			if v == false {
				break
			}
			if i == (len(containsAttr) - 1) {
				result[alias] = config
			}
		}
	}
	return result
}
