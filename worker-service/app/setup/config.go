package setup

import "worker-service/internal-lib/utils"

type MessageQueueConfig struct {
	URL          string
	ExchangeName string
}

type Config struct {
	*utils.DefaultConfig
	*MessageQueueConfig
}

func NewConfig(defaultConfig *utils.DefaultConfig) *Config {
	return &Config{
		DefaultConfig: defaultConfig,
		MessageQueueConfig: &MessageQueueConfig{
			URL:          utils.GetEnvOr("RABBITMQ_URL", "amqp://admin:password@localhost:5672/"),
			ExchangeName: utils.GetEnvOr("RABBITMQ_ORDER_QUEUE_EXCHANGE", ""),
		},
	}
}
