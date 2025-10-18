package routes

import (
	"net/http"

	"commander-service/app/setup"

	"github.com/danielgtaylor/huma/v2"
)

// RegisterRoutes registers all job routes to the API
func RegisterRoutes(api *huma.API, c *setup.Controllers) {
	// Job routes
	huma.Register(*api, huma.Operation{
		OperationID: "create-mission",
		Method:      http.MethodPost,
		Path:        "/missions",
		Summary:     "Create a new job",
		Description: "Create a new job with a name, optional description, scheduled time, and creator email. " +
			"The scheduled time must be a future Unix timestamp.",
		Tags:          []string{"Jobs"},
		DefaultStatus: http.StatusCreated,
	}, c.Mission.CreateMission)

	huma.Register(*api, huma.Operation{
		OperationID: "get-job-by-id",
		Method:      http.MethodGet,
		Path:        "/jobs/{id}",
		Summary:     "Get job by ID",
		Description: "Retrieve a job by its unique identifier.",
		Tags:        []string{"Jobs"},
	}, c.Mission.GetMissionByID)

}
