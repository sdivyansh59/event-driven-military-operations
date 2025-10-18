package crud

import (
	"context"

	"commander-service/internal-lib/database"
	"commander-service/internal-lib/database/query"
	"github.com/uptrace/bun"
)

type DB *bun.DB

// Handler provides CRUD operations for a database table.
// E is the Entity and ID is the primary key.
type Handler[E any, ID any] struct {
	db DB
}

func NewHandler[E any, ID any](db DB) *Handler[E, ID] {
	return &Handler[E, ID]{db: db}
}

// Search returns a list of entities.
// Use options to add additional conditions to the query.
func (h Handler[E, ID]) Search(ctx context.Context, options ...query.SearchOption) ([]E, error) {
	var result []E
	q := database.GetIDBFromContext(ctx, h.db).
		NewSelect().
		Model(&result)

	for _, option := range options {
		q = option(q)
	}

	err := q.Scan(ctx)

	return result, database.WrapError(err)
}

// GetByID returns a single entity by ID.
// Use options to add additional conditions to the query.
func (h Handler[E, ID]) GetByID(ctx context.Context, id ID, options ...query.SearchOption) (*E, error) {
	var result E

	q := database.GetIDBFromContext(ctx, h.db).NewSelect().Model(&result)
	if len(options) == 0 {
		q = q.Where("id = ?", id)
	}

	for _, option := range options {
		q = option(q)
	}

	err := q.Scan(ctx)

	return &result, database.WrapError(err)
}

// Create inserts a new entity.
func (h Handler[E, ID]) Create(ctx context.Context, entity *E) error {
	q := database.GetIDBFromContext(ctx, h.db).
		NewInsert().
		Model(entity)

	_, err := q.Exec(ctx)

	return database.WrapError(err)
}

// Update updates an existing entity.
func (h Handler[E, ID]) Update(ctx context.Context, entity *E) error {
	q := database.GetIDBFromContext(ctx, h.db).
		NewUpdate().
		Model(entity).
		WherePK()

	_, err := q.Exec(ctx)

	return database.WrapError(err)
}

// Delete deletes an existing entity.
func (h Handler[E, ID]) Delete(ctx context.Context, entity *E) error {
	q := database.GetIDBFromContext(ctx, h.db).
		NewDelete().
		Model(entity).
		WherePK()

	_, err := q.Exec(ctx)

	return database.WrapError(err)
}
