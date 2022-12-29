package dbcontext

import (
	"context"
	"database/sql"

	"github.com/quells/mastobot/internal/di"
)

type contextKeyType struct{}

var contextKey = contextKeyType{}

func Set(ctx context.Context, db *sql.DB) context.Context {
	return di.Set(ctx, contextKey, db)
}

func From(ctx context.Context) (*sql.DB, error) {
	return di.Get[contextKeyType, *sql.DB](ctx, contextKey, "*sql.DB")
}
