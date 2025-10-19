package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"worker-service/internal-lib/messaging"
	"worker-service/internal-lib/utils"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer handles RabbitMQ message consumption
type Consumer struct {
	*utils.WithLogger
	connection *amqp.Connection
	channel    *amqp.Channel
	config     *messaging.Config
	done       chan bool
}

// NewConsumer creates a new RabbitMQ message consumer
func NewConsumer(logger *utils.WithLogger, config) (*Consumer, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	consumer := &Consumer{
		WithLogger: logger,
		connection: conn,
		channel:    ch,
		config:     config,
		done:       make(chan bool),
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
	// Declare exchange (optional, using default exchange for simplicity)
	if c.config.ExchangeName != "" {
		err := c.channel.ExchangeDeclare(
			c.config.ExchangeName, // name
			"direct",              // type
			true,                  // durable
			false,                 // auto-deleted
			false,                 // internal
			false,                 // no-wait
			nil,                   // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange: %w", err)
		}
	}

	// Declare queue
	_, err := c.channel.QueueDeclare(
		c.config.QueueName, // name
		true,               // durable
		false,              // delete when unused
		false,              // exclusive
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange (if using custom exchange)
	if c.config.ExchangeName != "" {
		err = c.channel.QueueBind(
			c.config.QueueName,    // queue name
			c.config.RoutingKey,   // routing key
			c.config.ExchangeName, // exchange
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue: %w", err)
		}
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
		Str("queue", c.config.QueueName).
		Msg("Starting to consume messages")

	// Register consumer
	msgs, err := c.channel.Consume(
		c.config.QueueName, // queue
		"",                 // consumer tag (empty for auto-generated)
		false,              // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
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
	c.Logger.Info().
		Str("message_id", delivery.MessageId).
		Str("routing_key", delivery.RoutingKey).
		Msg("Processing message")

	// Parse the mission created message
	var message messaging.MissionCreatedMessage
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		c.Logger.Error().
			Err(err).
			Str("message_id", delivery.MessageId).
			Msg("Failed to unmarshal message")

		// Reject the message without requeuing for malformed messages
		delivery.Nack(false, false)
		return
	}

	// Process the message based on its content
	if err := c.handleMissionCreated(&message); err != nil {
		c.Logger.Error().
			Err(err).
			Str("mission_id", message.MissionID).
			Str("message_id", delivery.MessageId).
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
		Str("message_id", delivery.MessageId).
		Msg("Message processed successfully")
}

// handleMissionCreated processes mission created events
func (c *Consumer) handleMissionCreated(message *messaging.MissionCreatedMessage) error {
	c.Logger.Info().
		Str("mission_id", message.MissionID).
		Str("name", message.Name).
		Str("status", message.Status).
		Time("created_at", message.CreatedAt).
		Msg("Processing mission created event")

	// TODO: Add your business logic here
	// Examples of what you might do:
	// 1. Update mission status in database
	// 2. Send notifications
	// 3. Trigger other workflows
	// 4. Call external APIs
	// 5. Generate reports

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	// For now, just log the processing
	c.Logger.Info().
		Str("mission_id", message.MissionID).
		Msg("Mission created event processed successfully")

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

	// Establish new connection
	conn, err := amqp.Dial(c.config.URL)
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
