package database

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"commander-service/internal-lib/utils"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.elastic.co/apm/module/apmsql"
	_ "go.elastic.co/apm/module/apmsql/pgxv4"
)

const errorDetail = "See https://inheaden.atlassian.net/l/cp/nW1U21Fq for more information."
const pingTimeout = 5 * time.Second

var postgresDriverRegistered = false

type PostgresConfig struct {
	URL             string
	Host            string
	Port            uint16
	Username        string
	Password        string
	Database        string
	SSLMode         string
	SSLCertLocation string
}

// NewPostgres creates a database instance using environment variables.
//
// If POSTGRES_DB_URL is set, it will be used to connect to the database.
// Otherwise, the following environment variables are used:
//
//   - POSTGRES_DB_HOST
//   - POSTGRES_DB_PORT
//   - POSTGRES_DB_USERNAME
//   - POSTGRES_DB_PASSWORD
//   - POSTGRES_DB_DATABASE
//   - POSTGRES_DB_SSLMODE
//   - POSTGRES_DB_SSLROOT_LOCATION
//
// If POSTGRES_DB_URL is set, the other environment variables are ignored.
//
// See https://inheaden.atlassian.net/l/cp/nW1U21Fq for more information.
func NewPostgres(config *utils.DefaultConfig, log *zerolog.Logger) *bun.DB {
	postgresConfig := getPostgresConfig(config, log)

	if postgresConfig.URL != "" {
		parsed, err := pgx.ParseConfig(postgresConfig.URL)
		if err != nil {
			log.Fatal().Err(err).Msg("could not parse connection string")
		}

		postgresConfig.Host = parsed.Host
		postgresConfig.Port = parsed.Port
		postgresConfig.Username = parsed.User
		postgresConfig.Password = parsed.Password
		postgresConfig.Database = parsed.Database

		u := utils.MustParseURL(postgresConfig.URL)

		postgresConfig.SSLMode = u.Query().Get("sslmode")
		postgresConfig.SSLCertLocation = u.Query().Get("sslrootcert")
	}

	if !postgresDriverRegistered {
		apmsql.Register("postgres", &stdlib.Driver{})
		postgresDriverRegistered = true
	}

	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s sslrootcert=%s",
		postgresConfig.Host,
		postgresConfig.Port,
		postgresConfig.Username,
		postgresConfig.Password,
		postgresConfig.Database,
		postgresConfig.SSLMode,
		postgresConfig.SSLCertLocation,
	)

	sqlDB, err := apmsql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to database")
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)

	err = sqlDB.PingContext(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("could not connect to database")
	}

	log.Info().Msg("connected to database")
	cancel()

	return bun.NewDB(sqlDB, pgdialect.New())
}

func getPostgresConfig(config *utils.DefaultConfig, log *zerolog.Logger) *PostgresConfig {
	getEnv := func(key, or string) string {
		return utils.GetEnvOrPrefix(config.ServicePrefix, key, or)
	}

	databasePort, err := strconv.Atoi(utils.GetEnvOrPrefix(config.ServicePrefix, "POSTGRES_DB_PORT", "0"))
	if err != nil {
		panic(err)
	}

	result := &PostgresConfig{
		URL:             getEnv("POSTGRES_DB_URL", ""),
		Host:            getEnv("POSTGRES_DB_HOST", ""),
		Port:            uint16(databasePort),
		Username:        getEnv("POSTGRES_DB_USERNAME", ""),
		Password:        getEnv("POSTGRES_DB_PASSWORD", ""),
		Database:        getEnv("POSTGRES_DB_DATABASE", ""),
		SSLMode:         getEnv("POSTGRES_DB_SSLMODE", "require"),
		SSLCertLocation: getEnv("POSTGRES_DB_SSLROOT_LOCATION", ""),
	}

	if result.URL != "" {
		log.Info().Msg("using postgres url from environment variable")

		return result
	}

	log.Info().Msg("using individual postgres variables from environment")

	var missingVariables []string

	if result.Host == "" {
		missingVariables = append(missingVariables, "POSTGRES_DB_HOST")
	}

	if result.Port == 0 {
		missingVariables = append(missingVariables, "POSTGRES_DB_PORT")
	}

	if result.Username == "" {
		missingVariables = append(missingVariables, "POSTGRES_DB_USERNAME")
	}

	if result.Password == "" {
		missingVariables = append(missingVariables, "POSTGRES_DB_PASSWORD")
	}

	if result.Database == "" {
		missingVariables = append(missingVariables, "POSTGRES_DB_DATABASE")
	}

	if len(missingVariables) > 0 {
		log.Fatal().
			Str("missing_variables", strings.Join(missingVariables, ",")).
			Str("detail", errorDetail).
			Msg("This app is using a postgres instance. You have not specified POSTGRES_DB_URL but then you have to define all additional parameters")
	}

	return result
}
