package shared

import "math/rand"

const (
	StatusQueueName = "status_queue"
	OrderQueueName  = "order_queue"
	TokenQueueName  = "token_queue"
)

// MissionStatus represents the possible states of a job
type MissionStatus string

var Token string

const (
	MissionStatusCreated    MissionStatus = "CREATED"
	MissionStatusQueued     MissionStatus = "QUEUED"
	MissionStatusInProgress MissionStatus = "IN_PROGRESS"
	MissionStatusCompleted  MissionStatus = "COMPLETED"
	MissionStatusFailed     MissionStatus = "FAILED"
)

type OrderMessage struct {
	MissionID string
	Status    string
}

// GenerateRandomNumber generate random number between min and max
func GenerateRandomNumber(min, max int) int {
	return min + rand.Intn(max-min+1)
}
