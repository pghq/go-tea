package repository

import (
	"context"

	"github.com/pghq/go-museum/museum/store"
)

// Query gets a new query for searching the repository.
func (r *Repository) Query() store.Query {
	return r.client.Query()
}

// Search retrieves items from the repository matching criteria.
func (r *Repository) Search(ctx context.Context, query store.Query) (store.Cursor, error) {
	return query.Execute(ctx)
}
