package utils

import (
	"fmt"
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitGlobalLogger initializes the global zerolog logger and returns it
// This should be called once at application startup via Wire
func InitGlobalLogger(config *DefaultConfig) (*zerolog.Logger, error) {
	env := GetEnvOr("ENVIRONMENT", "production")
	isProd := env == "production"

	// Configure log level
	logLevel := zerolog.InfoLevel
	if config.IsDebug || os.Getenv("DEBUG") == "true" {
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	// Get hostname, with fallback
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
		// Log the error but don't fail initialization
	}

	// Configure output writer
	var output io.Writer = os.Stdout
	if isProd {
		// JSON format for production (better for log aggregators)
		output = os.Stdout
	} else {
		// Pretty console output for development
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006/01/02 15:04:05",
		}
	}

	// Set up the global logger
	logger := zerolog.New(output).
		With().
		Timestamp().
		Str("host", hostname).
		Str("env", env).
		Logger()

	// Set it as the global logger
	log.Logger = logger

	log.Info().
		Str("level", logLevel.String()).
		Str("mode", env).
		Msg("Global logger initialized successfully")

	return &logger, nil
}

// WithLogger can be used to compose a struct with a logger.
// It wraps a logger instance for dependency injection.
//
// Example:
//
//	type MyStruct struct {
//		*utils.WithLogger
//	}
type WithLogger struct {
	Logger *zerolog.Logger
}

// NewWithLogger creates a WithLogger wrapper around the provided logger
func NewWithLogger(logger *zerolog.Logger) (*WithLogger, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil, probably global loggrer not initialized yet")
	}
	return &WithLogger{Logger: logger}, nil
}

// NewTestWithLogger creates a test logger with debug mode enabled
func NewTestWithLogger() *WithLogger {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	}
	logger := zerolog.New(output).With().Timestamp().Caller().Logger()
	return &WithLogger{Logger: &logger}
}
