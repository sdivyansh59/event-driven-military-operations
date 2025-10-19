package database

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

// RunMigrationsFromPath runs migrations from a directory path
func RunMigrationsFromPath(ctx context.Context, db *bun.DB, path string, logger *zerolog.Logger) error {
	migrations := migrate.NewMigrations()

	// Use os.DirFS to create a filesystem for the migrations directory
	fsys := os.DirFS(path)

	// Discover migrations automatically using bun's built-in method
	if err := migrations.Discover(fsys); err != nil {
		logger.Error().Err(err).Msgf("Failed to discover migrations from path: %s", path)
		return fmt.Errorf("failed to discover migrations from path %s: %w", path, err)
	}

	return RunMigrations(ctx, db, migrations, logger)
}

// RunMigrations runs database migrations
func RunMigrations(ctx context.Context, db *bun.DB, migrations *migrate.Migrations, logger *zerolog.Logger) error {
	migrator := migrate.NewMigrator(db, migrations)

	if err := migrator.Init(ctx); err != nil {
		logger.Error().Err(err).Msg("Failed to initialize migrator")
		return fmt.Errorf("failed to initialize migrator: %w", err)
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to run migrations")
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if group.IsZero() {
		logger.Info().Msg("No new migrations to run")
	} else {
		logger.Info().Msgf("Migrated to %s", group)
	}

	return nil
}
