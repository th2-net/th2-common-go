package th2

import (
	"encoding/json"
	"strconv"

	"github.com/rs/zerolog"
)

type ConnectionManagerConfiguration struct {
	subscriberName               string
	connectionTimeout            int
	connectionCloseTimeout       int
	maxRecoveryAttempts          int
	minConnectionRecoveryTimeout int
	maxConnectionRecoveryTimeout int
	prefetchCount                int
	messageRecursionLimit        int
}

type RabbitMQConfiguration struct {
	host          string
	vHost         string
	port          string
	username      string
	password      string
	exchangeName  string
	configuration ConnectionManagerConfiguration

	Logger zerolog.Logger
}

func (rc *RabbitMQConfiguration) Init(raw_cfg []byte) error {

	jsonMap := make(map[string]json.RawMessage)

	err := json.Unmarshal(raw_cfg, &jsonMap)
	if err != nil {
		rc.Logger.Error().Err(err).Msg("deserialization error")
		return err
	}

	err = json.Unmarshal(jsonMap["host"], &rc.host)
	if err != nil {
		rc.Logger.Error().Err(err).Msg("deserialization error")
		return err
	}

	err = json.Unmarshal(jsonMap["vHost"], &rc.vHost)
	if err != nil {
		rc.Logger.Error().Err(err).Msg("deserialization error")
		return err
	}

	err = json.Unmarshal(jsonMap["port"], &rc.port)
	if err != nil {
		rc.port = "5672"
	}

	err = json.Unmarshal(jsonMap["username"], &rc.username)
	if err != nil {
		rc.Logger.Error().Err(err).Msg("deserialization error")
		return err
	}

	err = json.Unmarshal(jsonMap["password"], &rc.password)
	if err != nil {
		rc.Logger.Error().Err(err).Msg("deserialization error")
		return err
	}

	err = json.Unmarshal(jsonMap["exchangeName"], &rc.exchangeName)
	if err != nil {
		rc.exchangeName = ""
	}

	rc.configuration = ConnectionManagerConfiguration{"", -1, 10000, 5, 10000, 60000, 10, 100}

	json.Unmarshal(jsonMap["connectionTimeout"], &rc.configuration.connectionTimeout)
	json.Unmarshal(jsonMap["connectionCloseTimeout"], &rc.configuration.connectionCloseTimeout)
	json.Unmarshal(jsonMap["maxRecoveryAttempts"], &rc.configuration.maxRecoveryAttempts)
	json.Unmarshal(jsonMap["minConnectionRecoveryTimeout"], &rc.configuration.minConnectionRecoveryTimeout)
	json.Unmarshal(jsonMap["maxConnectionRecoveryTimeout"], &rc.configuration.maxConnectionRecoveryTimeout)
	json.Unmarshal(jsonMap["prefetchCount"], &rc.configuration.prefetchCount)
	json.Unmarshal(jsonMap["messageRecursionLimit"], &rc.configuration.messageRecursionLimit)

	return nil
}

func (rc *RabbitMQConfiguration) GetHost() string {
	return rc.host
}

func (rc *RabbitMQConfiguration) GetVHost() string {
	return rc.vHost
}

func (rc *RabbitMQConfiguration) GetPort() int {

	p, err := strconv.Atoi(rc.port)
	if err != nil {
		p = 5672
	}
	return p
}

func (rc *RabbitMQConfiguration) GetUsername() string {
	return rc.username
}

func (rc *RabbitMQConfiguration) GetPassword() string {
	return rc.password
}

func (rc *RabbitMQConfiguration) GetExchangeName() string {
	return rc.exchangeName
}

func (rc *RabbitMQConfiguration) GetSubscriberName() string {
	return rc.configuration.subscriberName
}

func (rc *RabbitMQConfiguration) GetConnectionTimeout() int {
	return rc.configuration.connectionTimeout
}

func (rc *RabbitMQConfiguration) GetConnectionCloseTimeout() int {
	return rc.configuration.connectionCloseTimeout
}

func (rc *RabbitMQConfiguration) GetMaxRecoveryAttempts() int {
	return rc.configuration.maxRecoveryAttempts
}

func (rc *RabbitMQConfiguration) GetMinConnectionRecoveryTimeout() int {
	return rc.configuration.minConnectionRecoveryTimeout
}

func (rc *RabbitMQConfiguration) GetMaxConnectionRecoveryTimeout() int {
	return rc.configuration.maxConnectionRecoveryTimeout
}

func (rc *RabbitMQConfiguration) GetPrefetchCount() int {
	return rc.configuration.prefetchCount
}

func (rc *RabbitMQConfiguration) GetmessageRecursionLimit() int {
	return rc.configuration.messageRecursionLimit
}
