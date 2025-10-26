package shared

const (
	MissionStatusCreated    MissionStatus = "CREATED"
	MissionStatusQueued     MissionStatus = "QUEUED"
	MissionStatusInProgress MissionStatus = "IN_PROGRESS"
	MissionStatusCompleted  MissionStatus = "COMPLETED"
	MissionStatusFailed     MissionStatus = "FAILED"
)

const OrderQueueName = "order_queue"
const StatusQueueName = "status_queue"
const TokenQueueName = "token_queue"

// MissionStatus represents the possible states of a job
type MissionStatus string

type OrderMessage struct {
	MissionID string `json:"missionID"`
	Status    string `json:"status"`
}

type StatusMessage struct {
	MissionID string `json:"mission_id"`
	Status    string `json:"status"`
	Token     string `json:"token"`
}

type TokenMessage struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}
