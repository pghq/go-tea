package repository

import (
	"context"

	"github.com/pghq/go-museum/museum/errors"
	"github.com/pghq/go-museum/museum/internal/database"
)

func (r *Repository) Query(ctx context.Context, query database.Query) (database.Cursor, error){
	if err := r.client.Query().Decode(query); err != nil{
		return nil, errors.BadRequest(err)
	}

	return query.Execute(ctx)
}
