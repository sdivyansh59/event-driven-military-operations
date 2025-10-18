package dbconfig

import (
	"context"
	"fmt"
	"time"

	"commander-service/internal-lib/database"
	"commander-service/internal-lib/utils"

	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
)

type CommandersCampDB struct {
	*bun.DB
}

func ProvideCommandersCampDB(logger *zerolog.Logger, withLogger *utils.WithLogger, config *utils.DefaultConfig) (*CommandersCampDB, error) {
	// Use the config as-is without changing ServicePrefix
	// This allows POSTGRES_DB_URL to be found correctly
	db := database.NewPostgres(config, logger)

	if utils.IsDebug() {
		db.AddQueryHook(database.NewDBLogger(withLogger))
	}

	// Verify database connection health
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Error().Err(err).Msg("Database health check failed")
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	logger.Info().Msg("Job Scheduler Database health check passed")

	jobSchedulerDB := &CommandersCampDB{DB: db}

	// Run migrations automatically as part of initialization
	if err := runMigrations(jobSchedulerDB, logger); err != nil {
		return nil, fmt.Errorf("failed to initialize database with migrations: %w", err)
	}

	return jobSchedulerDB, nil
}

// runMigrations is an internal helper to run migrations
func runMigrations(db *CommandersCampDB, logger *zerolog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get migrations path from environment or use default
	migrationsPath := utils.GetEnvOr("MIGRATIONS_PATH", "migrations")

	logger.Info().Msgf("Running migrations from path: %s", migrationsPath)

	// Run migrations
	if err := database.RunMigrationsFromPath(ctx, db.DB, migrationsPath, logger); err != nil {
		logger.Error().Err(err).Msg("Failed to run migrations")
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info().Msg("Migrations completed successfully")
	return nil
}

// AddJobSchedulerDBMigrationsHook is kept for backward compatibility or manual migration runs
// In normal operation, migrations are run automatically by ProvideJobSchedulerDB
func AddJobSchedulerDBMigrationsHook(db *CommandersCampDB, logger *zerolog.Logger) error {
	return runMigrations(db, logger)
}
