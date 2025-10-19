package app

import (
	"context"
	"net/http"
	"os"
	"worker-service/app/consumer"
	"worker-service/app/setup"
	"worker-service/internal-lib/utils"

	"github.com/rs/zerolog/log"
)

// App is the main application struct for worker service
type App struct {
	*utils.WithLogger
	config     *setup.Config
	httpServer *http.Server
	consumer   *consumer.Consumer
}

// NewApp creates a new worker service app instance
func newApp(config *setup.Config, logger *utils.WithLogger) *App {
	// Setup simple HTTP server with just health check
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthCheckHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Worker Service - Only health check available"))
	})

	httpServer := &http.Server{
		Addr:    config.HTTPAddress,
		Handler: mux,
	}

	// Initialize the consumer
	messageConsumer, err := consumer.NewConsumer(logger, config.MessageQueueConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create message consumer")
		os.Exit(1)
	}

	return &App{
		WithLogger: logger,
		config:     config,
		httpServer: httpServer,
		consumer:   messageConsumer,
	}
}

// Run starts the worker service (both message consumer and health check server)
func (a *App) Run(ctx context.Context) error {
	// Start HTTP server for health checks in background
	go func() {
		log.Info().Msgf("Starting health check server on %s", a.config.HTTPAddress)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Health check server failed")
		}
	}()

	// Start consuming messages (this is the main worker functionality)
	log.Info().Msg("Starting message consumer")
	return a.consumer.StartConsuming(ctx)
}

// Stop gracefully stops the worker service
func (a *App) Stop(ctx context.Context) error {
	log.Info().Msg("Stopping worker service")

	// Stop HTTP server
	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown HTTP server")
	}

	// Stop message consumer
	if err := a.consumer.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close message consumer")
		return err
	}

	return nil
}

// healthCheckHandler provides a simple health check endpoint
func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"healthy","service":"worker-service"}`))
}
