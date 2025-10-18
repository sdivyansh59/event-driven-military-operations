package shared

// MissionStatus represents the possible states of a job
type MissionStatus string

const (
	MissionQueued           MissionStatus = "QUEUED"
	MissionStatusInProgress MissionStatus = "IN_PROGRESS"
	MissionStatusCompleted  MissionStatus = "COMPLETED"
	MissionStatusFailed     MissionStatus = "FAILED"
)
