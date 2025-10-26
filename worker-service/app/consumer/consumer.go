package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"worker-service/app/producer"
	"worker-service/app/shared"
	"worker-service/internal-lib/utils"

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

	producer *producer.Producer
}

// NewConsumer creates a new RabbitMQ message consumer
func NewConsumer(logger *utils.WithLogger, producer *producer.Producer) (*Consumer, error) {
	rabbitMqURL := utils.GetEnvOr("RABBITMQ_URL", "amqp://admin:password@localhost:5672/")

	conn, err := amqp.Dial(rabbitMqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	consumer := &Consumer{
		WithLogger: logger,
		connection: conn,
		channel:    ch,
		done:       make(chan bool),
		producer:   producer,
	}

	// Initialize queue and exchange
	if err := consumer.setupQueue(); err != nil {
		_ = consumer.Close()
		return nil, fmt.Errorf("failed to setup queue: %w", err)
	}

	return consumer, nil
}

// setupQueue declares the queue and exchange for consumption
func (c *Consumer) setupQueue() error {
	// Declare order queue
	_, err := c.channel.QueueDeclare(
		shared.OrderQueueName, // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Declare token queue (we'll receive tokens from here)
	_, err = c.channel.QueueDeclare(
		shared.TokenQueueName, // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare token queue: %w", err)
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
		Str("queue", shared.OrderQueueName).
		Msg("Starting to consume messages")

	// Register consumer for order queue
	msgs, err := c.channel.Consume(
		shared.OrderQueueName, // queue
		"",                    // consumer tag (empty for auto-generated)
		false,                 // auto-ack
		false,                 // exclusive
		false,                 // no-local
		false,                 // no-wait
		nil,                   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	// Register consumer for token queue
	tokenMsgs, err := c.channel.Consume(
		shared.TokenQueueName, // queue
		"token_consumer",      // consumer tag
		false,                 // auto-ack
		false,                 // exclusive
		false,                 // no-local
		false,                 // no-wait
		nil,                   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register token consumer: %w", err)
	}

	// Process order messages in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				c.Logger.Info().Msg("Context cancelled, stopping order consumer")
				c.done <- true
				return
			case d, ok := <-msgs:
				if !ok {
					c.Logger.Warn().Msg("Order message channel closed")
					c.done <- true
					return
				}
				c.processMessage(ctx, d)
			}
		}
	}()

	// Process token messages in a separate goroutine (doesn't signal c.done)
	go func() {
		for {
			select {
			case <-ctx.Done():
				c.Logger.Info().Msg("Context cancelled, stopping token consumer")
				return
			case d, ok := <-tokenMsgs:
				if !ok {
					c.Logger.Warn().Msg("Token message channel closed")
					return
				}
				c.processToken(ctx, d)
			}
		}
	}()

	// Wait for done signal from order consumer
	<-c.done
	c.Logger.Info().Msg("Consumer stopped")
	return nil
}

// processMessage handles individual message processing
func (c *Consumer) processMessage(ctx context.Context, delivery amqp.Delivery) {
	c.Logger.Info().
		Str("message_id", delivery.MessageId).
		Str("routing_key", delivery.RoutingKey).
		Msg("Processing message")

	// Parse the mission created message
	var message shared.OrderMessage
	if err := json.Unmarshal(delivery.Body, &message); err != nil {
		c.Logger.Error().
			Err(err).
			Str("message_id", delivery.MessageId).
			Msg("Failed to unmarshal message")

		// Reject the message without requeuing for malformed messages
		if err := delivery.Nack(false, false); err != nil {
			c.Logger.Error().Err(err).Msg("failed to nack malformed message")
		}
		return
	}

	// Process the message based on its content
	if err := c.executeMission(ctx, &message); err != nil {
		// update status to FAILED
		if err := c.producer.PublishStatus(ctx, producer.MissionStatus{
			MissionID: message.MissionID,
			Status:    string(shared.MissionStatusFailed),
			Token:     c.producer.GetToken(),
		}); err != nil {
			c.Logger.Error().Err(err).
				Str("mission_id", message.MissionID).
				Msg("failed to publish completed status")
		}

		c.Logger.Error().
			Err(err).
			Str("mission_id", message.MissionID).
			Str("status", message.Status).
			Msg("Failed to process mission created message")

		// Reject and requeue the message for processing errors
		if err := delivery.Nack(false, true); err != nil {
			c.Logger.Error().Err(err).Msg("failed to nack for requeue")
		}
		return
	}

	// Acknowledge the message after successful processing
	if err := delivery.Ack(false); err != nil {
		c.Logger.Error().
			Err(err).
			Str("message_id", delivery.MessageId).
			Msg("Failed to acknowledge message")
	}

	// update status to completed
	if err := c.producer.PublishStatus(ctx, producer.MissionStatus{
		MissionID: message.MissionID,
		Status:    string(shared.MissionStatusCompleted),
		Token:     c.producer.GetToken(),
	}); err != nil {
		c.Logger.Error().Err(err).
			Str("mission_id", message.MissionID).
			Msg("failed to publish completed status")
	}

	c.Logger.Info().
		Str("mission_id", message.MissionID).
		Str("status", string(shared.MissionStatusCompleted)).
		Str("token", c.producer.GetToken()).
		Msg("Status update published successfully")
}

// processToken handles token messages received on the token queue
func (c *Consumer) processToken(ctx context.Context, delivery amqp.Delivery) {
	var token string

	if err := json.Unmarshal(delivery.Body, &token); err != nil {
		c.Logger.Error().Err(err).Str("message_id", delivery.MessageId).Msg("Failed to unmarshal token message")
		// reject malformed token messages without requeue
		if err := delivery.Nack(false, false); err != nil {
			c.Logger.Error().Err(err).Msg("failed to nack malformed token message")
		}
		return
	}

	if err := delivery.Ack(false); err != nil {
		c.Logger.Error().Err(err).Str("message_id", delivery.MessageId).Msg("Failed to acknowledge token message")
	}

	// Thread-safe token update
	c.producer.SetToken(token)

	c.Logger.Info().Str("token", token).Msg("new token received and updated in producer")
}

// executeMission processes mission created events
func (c *Consumer) executeMission(ctx context.Context, message *shared.OrderMessage) error {
	c.Logger.Info().
		Str("mission_id", message.MissionID).
		Str("status", message.Status).
		Msg("Executing mission")

	// update status to in-progress
	if err := c.producer.PublishStatus(ctx, producer.MissionStatus{
		MissionID: message.MissionID,
		Status:    string(shared.MissionStatusInProgress),
		Token:     c.producer.GetToken(),
	}); err != nil {
		return fmt.Errorf("failed to publish in-progress status: %w", err)
	}

	// Simulate processing time
	t := shared.GenerateRandomNumber(5, 15)
	time.Sleep(time.Duration(t) * time.Second)

	// success rate 80% times and failed 20% times
	isSuccess := shared.GenerateRandomNumber(1, 10) <= 8
	if !isSuccess {
		return fmt.Errorf("mission execution failed due to simulated error")
	}

	// For now, just log the processing
	c.Logger.Info().
		Str("mission_id", message.MissionID).
		Msg("Mission executed successfully")

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
	_ = c.Close()

	rabbitMQURL := utils.GetEnvOr("RABBITMQ_URL", "amqp://admin:password@localhost:5672/")
	// Establish new connection
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to reconnect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to open channel on reconnect: %w", err)
	}

	c.connection = conn
	c.channel = ch

	// Setup queue again
	if err := c.setupQueue(); err != nil {
		_ = c.Close()
		return fmt.Errorf("failed to setup queue on reconnect: %w", err)
	}

	c.Logger.Info().Msg("Successfully reconnected to RabbitMQ")
	return nil
}
