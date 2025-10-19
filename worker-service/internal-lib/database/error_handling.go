package database

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
)

// AppError represents an error with an HTTP status code
type AppError struct {
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	// ErrNotFound is returned when a database query returns no rows
	ErrNotFound = &AppError{
		Message:    "resource not found",
		StatusCode: http.StatusNotFound,
	}
)

// WrapError checks for knows database errors and wraps them.
func WrapError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	return err
}

// ErrContainsUniqueConstraintViolation checks if the given error includes 'SQLSTATE 23505' indicating a unique constraint violation.
func ErrContainsUniqueConstraintViolation(err error) bool {
	if err != nil && strings.Contains(err.Error(), "SQLSTATE 23505") {
		return true
	}

	return false
}
