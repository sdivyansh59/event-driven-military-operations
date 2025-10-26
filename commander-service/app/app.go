package app

import (
	"commander-service/app/auth"
	"commander-service/app/consumer"
	"context"
	"net/http"

	"commander-service/app/setup"
	"commander-service/app/setup/dbconfig"
	"commander-service/internal-lib/utils"
	"commander-service/routes"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun"
)

// App is the main application struct
type App struct {
	*utils.WithLogger
	router           *chi.Mux
	huma             *huma.API
	commandersCampDB *bun.DB
	controllers      *setup.Controllers
	config           *utils.DefaultConfig
	consumer         *consumer.Consumer
	authService      *auth.Service
}

func newApp(r *chi.Mux, h *huma.API, config *utils.DefaultConfig, c *setup.Controllers, logger *utils.WithLogger,
	commandersCampDB *dbconfig.CommandersCampDB, consumer *consumer.Consumer, auth *auth.Service) *App {

	config.HTTPAddress = utils.GetEnvOr("HTTP_PORT", ":8081")
	return &App{
		WithLogger:       logger,
		router:           r,
		huma:             h,
		commandersCampDB: commandersCampDB.DB,
		controllers:      c,
		config:           config,
		consumer:         consumer,
		authService:      auth,
	}
}

// Run starts the application server
func (a *App) Run() error {
	ctx := context.Background()
	// Configure routes
	a.registerRoutes()

	// start consumer
	go func() {
		err := a.consumer.StartConsuming(ctx)
		if err != nil {
			a.Logger.Fatal().Err(err).Msg("Failed to start consumer")
		}
	}()

	// start auth token publisher
	go func() {
		err := a.authService.PublishToken(ctx)
		if err != nil {
			a.Logger.Fatal().Err(err).Msg("Failed to start auth token refresher")
		}
	}()

	// Start the HTTP server
	a.Logger.Info().Msgf("Starting server on %s", a.config.HTTPAddress)
	return http.ListenAndServe(a.config.HTTPAddress, a.router)
}

// registerRoutes configures all API endpoints
func (a *App) registerRoutes() {
	if a.huma == nil {
		a.Logger.Fatal().Msgf("huma is nil")
	}

	routes.RegisterRoutes(a.huma, a.controllers)
}
