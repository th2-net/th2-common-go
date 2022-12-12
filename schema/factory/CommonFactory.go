package th2

import (
	"io"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	message "github.com/th2-net/th2-common-go/schema/message"
	cfg "github.com/th2-net/th2-common-go/schema/message/configuration"
	rmq_cfg "github.com/th2-net/th2-common-go/schema/message/impl/rabbitmq/configuration"
	rmq_connection "github.com/th2-net/th2-common-go/schema/message/impl/rabbitmq/connection"
	rmq_raw "github.com/th2-net/th2-common-go/schema/message/impl/rabbitmq/raw"
)

const (
	DEFAULT_CRADLE_INSTANCE_NAME   = "infra"
	EXACTPRO_IMPLEMENTATION_VENDOR = "Exactpro Systems LLC"
)

const (
	CONFIG_DEFAULT_PATH = "/var/th2/config/"

	RABBIT_MQ_FILE_NAME               = "rabbitMQ.json"
	ROUTER_MQ_FILE_NAME               = "mq.json"
	GRPC_FILE_NAME                    = "grpc.json"
	ROUTER_GRPC_FILE_NAME             = "grpc_router.json"
	CRADLE_CONFIDENTIAL_FILE_NAME     = "cradle.json"
	PROMETHEUS_FILE_NAME              = "prometheus.json"
	CUSTOM_FILE_NAME                  = "custom.json"
	BOX_FILE_NAME                     = "box.json"
	CONNECTION_MANAGER_CONF_FILE_NAME = "mq_router.json"
	CRADLE_NON_CONFIDENTIAL_FILE_NAME = "cradle_manager.json"

	DICTIONARY_DIR_NAME = "dictionary"

	RABBITMQ_SECRET_NAME   = "rabbitmq"
	CASSANDRA_SECRET_NAME  = "cassandra"
	RABBITMQ_PASSWORD_KEY  = "rabbitmq-password"
	CASSANDRA_PASSWORD_KEY = "cassandra-password"

	KEY_RABBITMQ_PASS  = "RABBITMQ_PASS"
	KEY_CASSANDRA_PASS = "CASSANDRA_PASS"

	GENERATED_CONFIG_DIR_NAME = "generated_configs"
)

type CommonFactory struct {
	rabbitMQ string
	routerMQ string

	rabbitMQCfg      *rmq_cfg.RabbitMQConfiguration
	messageRouterCfg *cfg.MessageRouterConfiguration

	messageRouterRawBatch     message.MessageRouter
	rabbitMQConnectionManager *rmq_connection.ConnectionManager

	logger zerolog.Logger
}

// NewCommonFactory establishes that both rabbitmq + routermq configs
// can be found and saves their paths
func NewCommonFactory() (*CommonFactory, error) {

	cf := CommonFactory{}

	cf.logger = log.Logger.With().Str("component", "CommonFactory").Logger()

	rabbitmq := filepath.Join(CONFIG_DEFAULT_PATH, RABBIT_MQ_FILE_NAME)
	if _, err := os.Stat(rabbitmq); err != nil {
		cf.logger.Error().Err(err).Str("filename", rabbitmq).Msg("rabbitmq config not found")
		return nil, err
	} else {
		cf.logger.Debug().Str("filename", rabbitmq).Msg("rabbitmq config found")
	}

	routermq := filepath.Join(CONFIG_DEFAULT_PATH, ROUTER_MQ_FILE_NAME)
	if _, err := os.Stat(routermq); err != nil {
		cf.logger.Error().Err(err).Str("filename", routermq).Msg("routermq config not found")
		return nil, err
	} else {
		cf.logger.Debug().Str("filename", routermq).Msg("routermq config found")
	}

	cf.rabbitMQ = rabbitmq
	cf.routerMQ = routermq

	return &cf, nil
}

// NewCommonFactoryFromArgs Creates CommonFactory from arguments
// --rabbitmq - path to json file with RabbitMQ configuration
// --routermq - path to json file with configuration for MessageRouter
// rabbitMq.json - configuration for RabbitMQ
// mq.json - configuration for MessageRouter
func NewCommonFactoryFromArgs(rabbitmq string, routermq string) (*CommonFactory, error) {

	cf := CommonFactory{}

	cf.logger = log.Logger

	if _, err := os.Stat(rabbitmq); err != nil {
		cf.logger.Error().Err(err).Str("filename", rabbitmq).Msg("rabbitmq config not found")
		return nil, err
	} else {
		cf.logger.Debug().Str("filename", rabbitmq).Msg("rabbitmq config found")
	}

	if _, err := os.Stat(routermq); err != nil {
		cf.logger.Error().Err(err).Str("filename", routermq).Msg("routermq config not found")
		return nil, err
	} else {
		cf.logger.Debug().Str("filename", routermq).Msg("routermq config found")
	}

	cf.rabbitMQ = rabbitmq
	cf.routerMQ = routermq

	return &cf, nil
}

func (cf *CommonFactory) Init() error {

	var err error

	cf.rabbitMQCfg, err = cf.loadRabbitmqConfiguration()
	if err != nil {
		cf.logger.Error().Err(err).Msg("load RabbitMQ configuration error")
		return err
	}

	cf.messageRouterCfg, err = cf.loadMessageRouterConfiguration()
	if err != nil {
		cf.logger.Error().Err(err).Msg("load MessageRouter configuration error")
		return err
	}

	return nil
}

func (cf *CommonFactory) GetMessageRouterRawBatch() (*message.MessageRouter, error) {

	if cf.messageRouterRawBatch == nil {

		//TBI
		cf.messageRouterRawBatch = &rmq_raw.RabbitRawBatchRouter{Logger: cf.logger}
		err := cf.messageRouterRawBatch.Init(cf.getRabbitMqConnectionManager(), cf.messageRouterCfg)

		if err != nil {
			cf.logger.Error().Err(err).Msg("cannot create raw message router")
			return nil, err
		}
	}

	return &cf.messageRouterRawBatch, nil
}

func (cf *CommonFactory) getRabbitMqConnectionManager() *rmq_connection.ConnectionManager {

	if cf.rabbitMQConnectionManager == nil {

		cf.rabbitMQConnectionManager = cf.createRabbitMQConnectionManager()
	}

	return cf.rabbitMQConnectionManager
}

func (cf *CommonFactory) createRabbitMQConnectionManager() *rmq_connection.ConnectionManager {

	rmq_cm := rmq_connection.NewConnectionManager(cf.rabbitMQCfg)
	rmq_cm.SetLogger(cf.logger)
	rmq_cm.Init()

	return rmq_cm
}

func (cf *CommonFactory) loadRabbitmqConfiguration() (*rmq_cfg.RabbitMQConfiguration, error) {

	jsonCfg, err := os.Open(cf.rabbitMQ)
	if err != nil {
		cf.logger.Error().Err(err).Msg("rabbitMQ config open error")
		return nil, err
	}

	raw_cfg, err := io.ReadAll(jsonCfg)
	if err != nil {
		cf.logger.Error().Err(err).Msg("rabbitMQ config read error")
		return nil, err
	}

	rc := rmq_cfg.RabbitMQConfiguration{Logger: cf.logger}

	rc.Init(raw_cfg)

	return &rc, nil
}

func (cf *CommonFactory) loadMessageRouterConfiguration() (*cfg.MessageRouterConfiguration, error) {
	rawCfg, err := os.ReadFile(cf.routerMQ)
	if err != nil {
		cf.logger.Error().Err(err).Msg("message router config read error")
	}

	mc := cfg.MessageRouterConfiguration{Logger: cf.logger}

	msgInitErr := mc.Init(string(rawCfg))
	if msgInitErr != nil {
		return nil, msgInitErr
	}

	return &mc, nil
}

func (cf *CommonFactory) SetLogger(l zerolog.Logger) {
	cf.logger = l.With().Str("component", "CommonFactory").Logger()
}
