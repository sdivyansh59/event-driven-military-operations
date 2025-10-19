package setup

import (
	"commander-service/app/shared"
	"commander-service/internal-lib/messaging"
	"commander-service/internal-lib/utils"
	"fmt"
)

// ProvideMessageProducer creates and configures the RabbitMQ message producer
func ProvideMessageProducer(logger *utils.WithLogger) (*messaging.MessageProducer, error) {
	// Get RabbitMQ URL from environment, with fallback
	rabbitMQURL := utils.GetEnvOr("RABBITMQ_URL", "amqp://admin:password@localhost:5672/")

	// Configure the messaging config
	msgConfig := &messaging.Config{
		URL:          rabbitMQURL,
		ExchangeName: "", // Use default exchange for simplicity
		QueueName:    shared.QueueNameOrder,
		RoutingKey:   shared.QueueNameOrder, // Same as queue name for default exchange
	}

	producer, err := messaging.NewMessageProducer(logger, msgConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create message producer: %w", err)
	}

	return producer, nil
}
