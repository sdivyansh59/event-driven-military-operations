package consumer

import (
	"commander-service/app/auth"
	"commander-service/app/mission"
	"commander-service/app/shared"
	"commander-service/internal-lib/snowflake"
	"commander-service/internal-lib/utils"
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer handles RabbitMQ message consumption
type Consumer struct {
	*utils.WithLogger
	connection      *amqp.Connection
	channel         *amqp.Channel
	orderQueueName  string
	statusQueueName string
	done            chan bool

	missionRepository mission.IRepository
	authService       *auth.Service
}

// NewConsumer creates a new RabbitMQ message consumer
func NewConsumer(logger *utils.WithLogger, missionRepo mission.IRepository, authService *auth.Service) (*Consumer, error) {
	rabbitMQURL := utils.GetEnvOr("RABBITMQ_URL", "amqp://admin:password@localhost:5672/")

	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	consumer := &Consumer{
		WithLogger:        logger,
		connection:        conn,
		channel:           ch,
		orderQueueName:    shared.OrderQueueName,
		statusQueueName:   shared.StatusQueueName,
		done:              make(chan bool),
		missionRepository: missionRepo,
		authService:       authService,
	}

	// Initialize queue and exchange
	if err := consumer.setupQueue(); err != nil {
		consumer.Close()
		return nil, fmt.Errorf("failed to setup queue: %w", err)
	}

	return consumer, nil
}

// setupQueue declares the queue and exchange for consumption
func (c *Consumer) setupQueue() error {
	// Declare queue
	_, err := c.channel.QueueDeclare(
		c.statusQueueName, // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Set QoS to control how many messages to prefetch
	err = c.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	return nil
}

// StartConsuming starts consuming messages from the queue
func (c *Consumer) StartConsuming(ctx context.Context) error {
	c.Logger.Info().
		Str("queue", c.statusQueueName).
		Msg("Starting to consume messages")

	// Register consumer
	msgs, err := c.channel.Consume(
		c.statusQueueName, // queue
		"",                // consumer tag (empty for auto-generated)
		false,             // auto-ack
		false,             // exclusive
		false,             // no-local
		false,             // no-wait
		nil,               // args
	)
	if err != nil {
		return fmt.Errorf("failed to register %s consumer: %w", c.statusQueueName, err)
	}

	// Process messages in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				c.Logger.Info().Msg("Context cancelled, stopping consumer")
				c.done <- true
				return
			case d, ok := <-msgs:
				if !ok {
					c.Logger.Warn().Msg("Message channel closed")
					c.done <- true
					return
				}
				c.processMessage(d)
			}
		}
	}()

	// Wait for done signal
	<-c.done
	c.Logger.Info().Msg("Consumer stopped")
	return nil
}

// processMessage handles individual message processing
func (c *Consumer) processMessage(delivery amqp.Delivery) {
	// Parse the status-update message
	var message shared.StatusMessage
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		c.Logger.Error().
			Err(err).
			Str("message_id", delivery.MessageId).
			Msg("Failed to unmarshal message")

		// Reject the message without requesting for malformed messages
		delivery.Nack(false, false)
		return
	}

	// validate message token
	if !c.authService.ValidateToken(message.Token) {
		c.Logger.Error().Str("token", message.Token).Msg("token validation failed")

		// Todo: send this message to retry-queue for token refresh and reprocessing
		// Reject the message without requesting for invalid token messages
		delivery.Nack(false, false)
		return
	}

	// Process the message based on its content
	if err := c.updateMissionStatus(context.Background(), &message); err != nil {
		c.Logger.Error().
			Err(err).
			Str("mission_id", message.MissionID).
			Str("status", message.Status).
			Msg("Failed to process mission created message")

		// Reject and requeue the message for processing errors
		delivery.Nack(false, true)
		return
	}

	// Acknowledge the message after successful processing
	if err := delivery.Ack(false); err != nil {
		c.Logger.Error().
			Err(err).
			Str("message_id", delivery.MessageId).
			Msg("Failed to acknowledge message")
	}

	c.Logger.Info().
		Str("mission_id", message.MissionID).
		Str("status", message.Status).
		Msg("Message status updated successfully")
}

// updateOrderStatus processes mission created events
func (c *Consumer) updateMissionStatus(ctx context.Context, message *shared.StatusMessage) error {
	id, err := snowflake.ConvertToSnowflake(message.MissionID)
	if err != nil {
		return fmt.Errorf("invalid mission ID: %w", err)
	}

	entity, err := c.missionRepository.GetMissionByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get mission by ID: %w", err)
	}

	entity.Status = shared.MissionStatus(message.Status)
	_, err = c.missionRepository.UpdateMission(ctx, entity)
	if err != nil {
		return fmt.Errorf("failed to update mission: %w", err)
	}

	return nil
}

// Stop gracefully stops the consumer
func (c *Consumer) Stop() {
	c.Logger.Info().Msg("Stopping consumer")
	close(c.done)
}

// Close closes the RabbitMQ connection and channel
func (c *Consumer) Close() error {
	if c.channel != nil {
		if err := c.channel.Close(); err != nil {
			c.Logger.Error().Err(err).Msg("Failed to close RabbitMQ channel")
		}
	}
	if c.connection != nil {
		if err := c.connection.Close(); err != nil {
			c.Logger.Error().Err(err).Msg("Failed to close RabbitMQ connection")
			return err
		}
	}
	return nil
}

// IsConnected checks if the connection is still alive
func (c *Consumer) IsConnected() bool {
	return c.connection != nil && !c.connection.IsClosed()
}

// Reconnect attempts to reconnect to RabbitMQ
func (c *Consumer) Reconnect() error {
	c.Logger.Info().Msg("Attempting to reconnect to RabbitMQ")

	// Close existing connections
	c.Close()

	rabbitMQURL := utils.GetEnvOr("RABBITMQ_URL", "amqp://admin:password@localhost:5672/")

	// Establish new connection
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel on reconnect: %w", err)
	}

	c.connection = conn
	c.channel = ch

	// Setup queue again
	if err := c.setupQueue(); err != nil {
		c.Close()
		return fmt.Errorf("failed to setup queue on reconnect: %w", err)
	}

	c.Logger.Info().Msg("Successfully reconnected to RabbitMQ")
	return nil
}
