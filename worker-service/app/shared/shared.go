package shared

import "math/rand"

const (
	StatusQueueName = "status_queue"
	OrderQueueName  = "order_queue"
)

type OrderMessage struct {
	MissionID string
	Status    string
}

// GenerateRandomNumber generate random number between min and max
func GenerateRandomNumber(min, max int) int {
	return min + rand.Intn(max-min+1)
}
