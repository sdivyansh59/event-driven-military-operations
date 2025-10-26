package auth

import (
	"commander-service/app/shared"
	"context"
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"commander-service/internal-lib/utils"

	amqp "github.com/rabbitmq/amqp091-go"
)

const tokenLifeSpan = time.Second * 30

type Service struct {
	*utils.WithLogger
	channel        *amqp.Channel
	secretKey      string
	currentToken   string
	tokenExpiry    time.Time
	tokenLifespan  time.Duration
	tokenQueueName string
	mu             sync.RWMutex
}

func NewService(logger *utils.WithLogger) (*Service, error) {
	secret := utils.GetEnvOr("SECRET_KEY", "default-secret-key-change-me")
	// Todo: move RabbitMQ connection logic to a shared location
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

	// Declare token queue
	_, err = channel.QueueDeclare(
		shared.TokenQueueName, // name
		true,                  // durable
		false,                 // delete when unused
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare token queue: %w", err)
	}

	return &Service{
		channel:        channel,
		WithLogger:     logger,
		secretKey:      secret,
		tokenLifespan:  tokenLifeSpan,
		tokenQueueName: shared.TokenQueueName,
	}, nil
}

// generateToken creates a new token with HMAC signature
// Token format: token:expiry_unix:signature
// The signature is computed over (token + expiry) to ensure integrity
func (s *Service) generateToken() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate random token using crypto/rand
	tokenBytes := make([]byte, 32)
	if _, err := crand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(tokenBytes)
	expiryTime := time.Now().Add(s.tokenLifespan)
	expiryUnix := expiryTime.Unix()

	// Create HMAC signature: HMAC(token + expiry, secretKey)
	payload := fmt.Sprintf("%s:%d", token, expiryUnix)
	signature := s.sign(payload)

	s.tokenExpiry = expiryTime
	fullToken := fmt.Sprintf("%s:%d:%s", token, expiryUnix, signature)
	s.currentToken = fullToken

	return fullToken, nil
}

// ValidateToken verifies the token signature and expiry
// Expected token format: token:expiry_unix:signature
func (s *Service) ValidateToken(token string) bool {
	// Parse token in the form token:expiry:signature safely
	parts := strings.SplitN(token, ":", 3)
	if len(parts) != 3 {
		return false
	}
	tokenPart := parts[0]
	expiryStr := parts[1]
	signature := parts[2]

	expiryUnix, err := strconv.ParseInt(expiryStr, 10, 64)
	if err != nil {
		return false
	}

	// Verify signature (must match what we signed during generation)
	payload := fmt.Sprintf("%s:%d", tokenPart, expiryUnix)
	expectedSig := s.sign(payload)
	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return false
	}

	// Check expiry - token is expired if current time is past the expiry time
	if time.Now().Unix() > expiryUnix {
		return false
	}

	return true
}

func (s *Service) sign(payload string) string {
	h := hmac.New(sha256.New, []byte(s.secretKey))
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Service) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return time.Now().After(s.tokenExpiry)
}

// PublishToken generates and publishes tokens to token_queue every 30 seconds
func (s *Service) PublishToken(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Publish initial token immediately
	err := s.publishSingleToken(ctx)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Failed to publish initial token")
		return err
	}

	for {
		select {
		case <-ticker.C:
			err = s.publishSingleToken(ctx)
			if err != nil {
				s.Logger.Error().Err(err).Msg("Failed to publish token")
				return err
			}

		case <-ctx.Done():
			s.Logger.Info().Msg("Stopping token publisher")
			return err
		}
	}
}

// publishSingleToken generates and publishes a single token to the token queue
func (s *Service) publishSingleToken(ctx context.Context) error {
	// Generate token using auth service
	token, err := s.generateToken()
	if err != nil {
		return err
	}

	tokenMsg := token

	body, err := json.Marshal(tokenMsg)
	if err != nil {
		return err
	}

	err = s.channel.PublishWithContext(
		ctx,
		"",               // exchange
		s.tokenQueueName, // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return err
	}

	s.Logger.Info().Str("token", token).Msg("Published updated token to token_queue")
	return nil
}
