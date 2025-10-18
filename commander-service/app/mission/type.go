package mission

import (
	"commander-service/internal-lib/snowflake"
	"time"
)

type Mission struct {
	ID          snowflake.ID
	Name        string
	Description string
	Status      string
	CreatedBy   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type MissionDTO struct {
	ID          string
	Name        string
	Description string
	Status      string
	CreatedBy   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateMissionInput struct {
}

type CreateMissionResponse struct {
}

type GetMissionByIDInput struct {
}

type GetMissionByIDResponse struct {
}
