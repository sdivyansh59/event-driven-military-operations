package producer

import (
	"commander-service/app/shared"
	"commander-service/internal-lib/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	*utils.WithLogger
	conn            *amqp.Connection
	channel         *amqp.Channel
	statusQueueName string
	orderQueueName  string
}

type MissionStatus struct {
	MissionID string `json:"mission_id"`
	Status    string `json:"status"`
}

func NewProducer(logger *utils.WithLogger) (*Producer, error) {
	rabbitmqURL := utils.GetEnvOr("RABBITMQ_URL", "amqp://admin:password@localhost:5672/")

	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	_, err = channel.QueueDeclare(
		shared.OrderQueueName, // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &Producer{
		WithLogger:      logger,
		conn:            conn,
		channel:         channel,
		statusQueueName: shared.StatusQueueName,
		orderQueueName:  shared.OrderQueueName,
	}, nil
}

func (p *Producer) PublishOrder(ctx context.Context, message *shared.OrderMessage) error {
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal order message: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		"",               // exchange
		p.orderQueueName, // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published order for mission %s: %s", message.MissionID, message.Status)
	return nil
}

func (p *Producer) Close() error {
	if p.channel != nil {
		p.channel.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}
