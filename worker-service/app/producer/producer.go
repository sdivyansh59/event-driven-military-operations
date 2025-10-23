package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"worker-service/app/shared"
	"worker-service/internal-lib/utils"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Producer struct {
	*utils.WithLogger
	conn            *amqp.Connection
	channel         *amqp.Channel
	statusQueueName string
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
		shared.StatusQueueName, // name
		true,                   // durable
		false,                  // delete when unused
		false,                  // exclusive
		false,                  // no-wait
		nil,                    // arguments
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
	}, nil
}

func (p *Producer) PublishStatus(ctx context.Context, status MissionStatus) error {
	body, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	err = p.channel.PublishWithContext(
		ctx,
		"",                // exchange
		p.statusQueueName, // routing key
		false,             // mandatory
		false,             // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published status for mission %s: %s", status.MissionID, status.Status)
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
