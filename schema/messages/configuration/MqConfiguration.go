package configuration

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"os"
)

//type ConnectionManagerConfiguration struct {
//	subscriberName               string
//	ConnectionTimeout            int `json:"connectionTimeout"`
//	ConnectionCloseTimeout       int `json:"connectionCloseTimeout"`
//	MaxRecoveryAttempts          int `json:"maxRecoveryAttempts"`
//	MinConnectionRecoveryTimeout int `json:"minConnectionRecoveryTimeout"`
//	MaxConnectionRecoveryTimeout int `json:"maxConnectionRecoveryTimeout"`
//	PrefetchCount                int `json:"prefetchCount"`
//	MessageRecursionLimit        int `json:"messageRecursionLimit"`
//}

type RabbitMQConfiguration struct {
	Host                         string `json:"host"`
	VHost                        string `json:"VHost"`
	Port                         string `json:"port"`
	Username                     string `json:"username"`
	Password                     string `json:"password"`
	ExchangeName                 string `json:"exchangeName"`
	ConnectionTimeout            int    `json:"connectionTimeout"`
	ConnectionCloseTimeout       int    `json:"connectionCloseTimeout"`
	MaxRecoveryAttempts          int    `json:"maxRecoveryAttempts"`
	MinConnectionRecoveryTimeout int    `json:"minConnectionRecoveryTimeout"`
	MaxConnectionRecoveryTimeout int    `json:"maxConnectionRecoveryTimeout"`
	PrefetchCount                int    `json:"prefetchCount"`
	MessageRecursionLimit        int    `json:"messageRecursionLimit"`

	Logger zerolog.Logger
}

func (mq *RabbitMQConfiguration) Init(path string) error {
	content, err := os.ReadFile(path) // Read json file
	if err != nil {
		mq.Logger.Error().Err(err).Msg("json file reading error")
		return err
	}
	fail := json.Unmarshal(content, mq)
	if fail != nil {
		mq.Logger.Error().Err(err).Msg("Deserialization error")
		fmt.Println(err)
		return err
	}
	return nil
}

//
//func (mq *RabbitMQConfiguration) GetConnectionTimeout() int {
//	return mq.ConnectionTimeout
//}
//
//func (mq *RabbitMQConfiguration) GetConnectionCloseTimeout() int {
//	return mq.ConnectionCloseTimeout
//}
//
//func (mq *RabbitMQConfiguration) GetMaxRecoveryAttempts() int {
//	return mq.MaxRecoveryAttempts
//}
//
//func (mq *RabbitMQConfiguration) GetMinConnectionRecoveryTimeout() int {
//	return mq.MinConnectionRecoveryTimeout
//}
//
//func (mq *RabbitMQConfiguration) GetMaxConnectionRecoveryTimeout() int {
//		return mq.MaxConnectionRecoveryTimeout
//}
//
//func (mq *RabbitMQConfiguration) GetPrefetchCount() int {
//		return mq.PrefetchCount
//}
//
//func (mq *RabbitMQConfiguration) GetmessageRecursionLimit() int {
//	return mq.MessageRecursionLimit
//}
