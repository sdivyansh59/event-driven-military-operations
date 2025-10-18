//go:build wireinject
// +build wireinject

package app

import (
	"commander-service/app/mission"
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
		setup.ProvideControllers,

		// Application
		newApp,

		// Initialize application controllers, converter and repositories
		// mission
		mission.NewController,
		mission.NewConverter,
		mission.NewRepository,
	)
	return nil, nil
}
