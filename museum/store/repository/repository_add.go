package repository

import (
	"context"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/store"
)

// Add adds items(s) to the repository
func (r *Repository) Add(ctx context.Context, collection string, items ...store.Item) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.client.Transaction(ctx)
	if err != nil {
		return errors.Wrap(err)
	}
	defer tx.Rollback()

	for _, item := range items {
		if err := item.Validate(); err != nil {
			return errors.BadRequest(err)
		}

		command := r.client.Add().To(collection).Item(item.Map())
		if _, err := tx.Execute(command); err != nil {
			return errors.Wrap(err)
		}
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err)
	}

	return nil
}
