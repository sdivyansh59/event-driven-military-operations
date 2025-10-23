package routes

import (
	"context"
	"net/http"

	"commander-service/app/setup"

	"github.com/danielgtaylor/huma/v2"
)

type HealthRequest struct{}
type HealthResponse struct {
	Body struct {
		Status  string `json:"status"`
		Service string `json:"service,omitempty"`
	}
}

// RegisterRoutes registers all mission routes to the API
func RegisterRoutes(api *huma.API, c *setup.Controllers) {
	// Health check route
	huma.Register(*api, huma.Operation{
		OperationID:   "health-check",
		Method:        http.MethodGet,
		Path:          "/health",
		Summary:       "Health check",
		Description:   "Check if the service is running.",
		Tags:          []string{"Health"},
		DefaultStatus: http.StatusOK,
	}, func(ctx context.Context, input *HealthRequest) (*HealthResponse, error) {
		return &HealthResponse{
			Body: struct {
				Status  string `json:"status"`
				Service string `json:"service,omitempty"`
			}{
				Status:  "ok",
				Service: "commander-service",
			},
		}, nil
	})

	// Mission routes
	huma.Register(*api, huma.Operation{
		OperationID:   "create-mission",
		Method:        http.MethodPost,
		Path:          "/missions",
		Summary:       "Create a new mission",
		Description:   "Create a new mission with a name, optional description, and creator information.",
		Tags:          []string{"Missions"},
		DefaultStatus: http.StatusAccepted,
	}, c.Mission.CreateMission)

	huma.Register(*api, huma.Operation{
		OperationID:   "get-mission-by-id",
		Method:        http.MethodGet,
		Path:          "/missions/{id}",
		Summary:       "Get mission by ID",
		Description:   "Retrieve a mission by its unique identifier.",
		Tags:          []string{"Missions"},
		DefaultStatus: http.StatusOK,
	}, c.Mission.GetMissionByID)
}
