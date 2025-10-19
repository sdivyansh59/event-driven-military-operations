package routes

import (
	"net/http"

	"commander-service/app/setup"

	"github.com/danielgtaylor/huma/v2"
)

// RegisterRoutes registers all mission routes to the API
func RegisterRoutes(api *huma.API, c *setup.Controllers) {
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
