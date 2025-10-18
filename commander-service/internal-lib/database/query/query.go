package query

import (
	"fmt"

	"github.com/uptrace/bun"
)

type SearchOption func(*bun.SelectQuery) *bun.SelectQuery

func Where[T any](attr string, v T) SearchOption {
	return func(query *bun.SelectQuery) *bun.SelectQuery {
		return query.Where(fmt.Sprintf("%s = ?", attr), v)
	}
}
