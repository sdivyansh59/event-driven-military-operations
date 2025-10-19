package database

import (
	"context"
	"time"

	"commander-service/internal-lib/utils"
	"github.com/uptrace/bun"
)

type DBLogger struct {
	*utils.WithLogger
}

var _ bun.QueryHook = &DBLogger{}

func NewDBLogger(withLogger *utils.WithLogger) *DBLogger {
	return &DBLogger{WithLogger: withLogger}
}

func (d DBLogger) BeforeQuery(c context.Context, _ *bun.QueryEvent) context.Context {
	return c
}

func (d DBLogger) AfterQuery(_ context.Context, q *bun.QueryEvent) {
	duration := time.Since(q.StartTime)

	if q.Err != nil {
		d.Logger.Error().
			Err(q.Err).
			Str("query", q.Query).
			Dur("duration", duration).
			Msg("Database query failed")
	} else {
		d.Logger.Debug().
			Str("query", q.Query).
			Dur("duration", duration).
			Msg("Database query executed")
	}
}
