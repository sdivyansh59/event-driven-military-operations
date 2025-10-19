package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"worker-service/internal-lib/utils"

	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageProducer handles RabbitMQ message publishing
type MessageProducer struct {
	*utils.WithLogger
	connection *amqp.Connection
	channel    *amqp.Channel
	config     *Config
}

// Config holds RabbitMQ configuration
type Config struct {
	URL          string
	ExchangeName string
	QueueName    string
	RoutingKey   string
}

// MissionCreatedMessage represents the message sent when a mission is created
type MissionCreatedMessage struct {
	MissionID string    `json:"mission_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy *string   `json:"created_by,omitempty"`
}

// NewMessageProducer creates a new RabbitMQ message producer
func NewMessageProducer(logger *utils.WithLogger, config *Config) (*MessageProducer, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	producer := &MessageProducer{
		WithLogger: logger,
		connection: conn,
		channel:    ch,
		config:     config,
	}

	// Initialize queue and exchange
	if err := producer.setupQueue(); err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to setup queue: %w", err)
	}

	return producer, nil
}

// setupQueue declares the queue and exchange
func (p *MessageProducer) setupQueue() error {
	// Declare exchange (optional, using default exchange for simplicity)
	if p.config.ExchangeName != "" {
		err := p.channel.ExchangeDeclare(
			p.config.ExchangeName, // name
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
	_, err := p.channel.QueueDeclare(
		p.config.QueueName, // name
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
	if p.config.ExchangeName != "" {
		err = p.channel.QueueBind(
			p.config.QueueName,    // queue name
			p.config.RoutingKey,   // routing key
			p.config.ExchangeName, // exchange
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue: %w", err)
		}
	}

	return nil
}

// PublishMissionCreated publishes a mission created event to the order_queue
func (p *MessageProducer) PublishMissionCreated(ctx context.Context, message *MissionCreatedMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	exchangeName := p.config.ExchangeName
	routingKey := p.config.RoutingKey

	// If no exchange specified, use default exchange with queue name as routing key
	if exchangeName == "" {
		routingKey = p.config.QueueName
	}

	err = p.channel.PublishWithContext(
		ctx,
		exchangeName, // exchange
		routingKey,   // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // make message persistent
			Timestamp:    time.Now(),
			Body:         body,
			MessageId:    message.MissionID,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.Logger.Info().
		Str("mission_id", message.MissionID).
		Str("queue", p.config.QueueName).
		Msg("Mission created message published successfully")

	return nil
}

// Close closes the RabbitMQ connection and channel
func (p *MessageProducer) Close() error {
	if p.channel != nil {
		if err := p.channel.Close(); err != nil {
			p.Logger.Error().Err(err).Msg("Failed to close RabbitMQ channel")
		}
	}
	if p.connection != nil {
		if err := p.connection.Close(); err != nil {
			p.Logger.Error().Err(err).Msg("Failed to close RabbitMQ connection")
			return err
		}
	}
	return nil
}

// IsConnected checks if the connection is still alive
func (p *MessageProducer) IsConnected() bool {
	return p.connection != nil && !p.connection.IsClosed()
}
