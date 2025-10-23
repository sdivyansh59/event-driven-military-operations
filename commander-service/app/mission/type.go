package mission

import (
	"commander-service/app/shared"
	"commander-service/internal-lib/snowflake"
	"time"

	"github.com/uptrace/bun"
)

type MissionEntity struct {
	bun.BaseModel `bun:"mission,alias:mission"`

	ID          snowflake.ID         `bun:"id,pk"`
	Name        string               `bun:"name"`
	Description *string              `bun:"description"`
	Status      shared.MissionStatus `bun:"status"`
	CreatedBy   *string              `bun:"created_by"`
	CreatedAt   time.Time            `bun:"created_at"`
	UpdatedAt   time.Time            `bun:"updated_at"`
}

type MissionDTO struct {
	ID          string
	Name        string
	Status      string
	Description *string
	CreatedBy   *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateMissionInput struct {
	Body struct {
		Name        string  `json:"name" validate:"required"`
		Description *string `json:"description,omitempty"`
		CreatedBy   *string `json:"created_by,omitempty"`
	}
}

type CreateMissionResponse struct {
	Body MissionDTO
}

type GetMissionByIDInput struct {
	ID string `path:"id" validate:"required,numeric,min=1" doc:"Unique identifier of the mission"`
}

type GetMissionByIDResponse struct {
	Body MissionDTO
}
