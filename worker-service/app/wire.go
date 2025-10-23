//go:build wireinject
// +build wireinject

package app

import (
	"worker-service/app/setup"
	"worker-service/internal-lib/utils"

	"github.com/google/wire"
)

// InitializeApp wires up all dependencies and returns the application/service instance
func InitializeApp() (*App, error) {
	wire.Build(
		// Core configuration and logger
		utils.ProvideDefaultConfig,
		utils.InitGlobalLogger,
		utils.NewWithLogger,
		setup.NewConfig,

		// Main application
		newApp,

		// producer
		//producer.NewProducer,
	)
	return &App{}, nil
}
