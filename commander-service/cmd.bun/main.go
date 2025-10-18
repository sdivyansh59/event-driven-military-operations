package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/migrate"
	"github.com/urfave/cli/v2"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	app := &cli.App{
		Name: "migration",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "create migration tables",
				Action: func(c *cli.Context) error {
					db := connectDB()
					defer db.Close()

					migrator := migrate.NewMigrator(db, migrate.NewMigrations())
					return migrator.Init(context.Background())
				},
			},
			{
				Name:  "migrate",
				Usage: "migrate database",
				Action: func(c *cli.Context) error {
					db := connectDB()
					defer db.Close()

					migrator := migrate.NewMigrator(db, migrate.NewMigrations(migrate.WithMigrationsDirectory("../migrations")))
					group, err := migrator.Migrate(context.Background())
					if err != nil {
						return err
					}

					if group.ID == 0 {
						fmt.Printf("there are no new migrations to run\n")
						return nil
					}

					fmt.Printf("migrated to %s\n", group)
					return nil
				},
			},
			{
				Name:  "rollback",
				Usage: "rollback the last migration group",
				Action: func(c *cli.Context) error {
					db := connectDB()
					defer db.Close()

					migrator := migrate.NewMigrator(db, migrate.NewMigrations(migrate.WithMigrationsDirectory("../migrations")))
					group, err := migrator.Rollback(context.Background())
					if err != nil {
						return err
					}

					if group.ID == 0 {
						fmt.Printf("there are no groups to roll back\n")
						return nil
					}

					fmt.Printf("rolled back %s\n", group)
					return nil
				},
			},
			{
				Name:  "status",
				Usage: "print migration status",
				Action: func(c *cli.Context) error {
					db := connectDB()
					defer db.Close()

					migrator := migrate.NewMigrator(db, migrate.NewMigrations(migrate.WithMigrationsDirectory("../migrations")))
					ms, err := migrator.MigrationsWithStatus(context.Background())
					if err != nil {
						return err
					}
					fmt.Printf("migrations: %s\n", ms)
					fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
					fmt.Printf("last migration group: %s\n", ms.LastGroup())
					return nil
				},
			},
			{
				Name:  "create_sql",
				Usage: "create a new SQL migration",
				Action: func(c *cli.Context) error {
					migrator := migrate.NewMigrator(nil, migrate.NewMigrations(migrate.WithMigrationsDirectory("../migrations")))
					name := c.Args().First()
					if name == "" {
						return fmt.Errorf("migration name is required")
					}
					files, err := migrator.CreateSQLMigrations(context.Background(), name)
					if err != nil {
						return err
					}
					for _, mf := range files {
						fmt.Printf("created migration %s (%s)\n", mf.Name, mf.Path)
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func connectDB() *bun.DB {
	// Get database URL from environment
	dbURL := os.Getenv("POSTGRES_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/commanders_camp_db?sslmode=disable"
	}

	// Connect to PostgreSQL
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dbURL)))
	db := bun.NewDB(sqldb, pgdialect.New())

	return db
}
