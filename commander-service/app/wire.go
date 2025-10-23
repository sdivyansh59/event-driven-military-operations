//go:build wireinject
// +build wireinject

package app

import (
	"commander-service/app/consumer"
	"commander-service/app/mission"
	"commander-service/app/producer"
	"commander-service/app/setup"
	"commander-service/app/setup/dbconfig"
	"commander-service/internal-lib/utils"

	"github.com/google/wire"
)

// InitializeApp wires up all dependencies and returns the application/service instance
func InitializeApp() (*App, error) {
	wire.Build(
		// Core configuration and logger
		utils.ProvideDefaultConfig,
		utils.InitGlobalLogger,
		utils.NewWithLogger,

		// Database initialization with migrations
		dbconfig.ProvideCommandersCampDB,

		// Infrastructure
		setup.ProvideSingletonChiRouter,
		setup.ProvideSingletonHuma,
		setup.ProvideSnowflakeGenerator,

		// Controllers
		setup.ProvideControllers,

		// Mission module
		mission.NewController,
		mission.NewRepository,
		mission.NewConverter,

		// Main application
		newApp,
		consumer.NewConsumer,
		producer.NewProducer,
	)
	return &App{}, nil
}
