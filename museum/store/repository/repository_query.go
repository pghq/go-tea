package repository

import (
	"context"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/internal/database"
)

// Query retrieves items from the repository matching criteria.
func (r *Repository) Query(ctx context.Context, query database.Query) (database.Cursor, error) {
	if err := r.client.Query().Decode(query); err != nil {
		return nil, errors.BadRequest(err)
	}

	return query.Execute(ctx)
}
