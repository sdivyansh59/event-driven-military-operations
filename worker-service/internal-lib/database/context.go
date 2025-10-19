package database

import (
	"context"

	"github.com/uptrace/bun"
)

type key string

const (
	dbContextKey key = "db_context"
)

type DBContext struct {
	Tx *bun.Tx // Holds the database transaction
}

func GetDBContext(ctx context.Context) *DBContext {
	v := ctx.Value(dbContextKey)
	if r, ok := v.(*DBContext); ok {
		return r
	}

	return nil
}

func GetIDBFromContext(ctx context.Context, db *bun.DB) bun.IDB {
	dBContext := GetDBContext(ctx)

	if dBContext == nil || dBContext.Tx == nil {
		return db
	}

	return dBContext.Tx
}
