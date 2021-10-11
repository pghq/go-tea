package repository

import (
	"context"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/internal/database"
)

// Remove removes items from the repository matching criteria.
func (r *Repository) Remove(ctx context.Context, command database.Remove) (uint, error){
	if err := r.client.Remove().Decode(command); err != nil{
		return 0, errors.BadRequest(err)
	}

	return command.Execute(ctx)
}
