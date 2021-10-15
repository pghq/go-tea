package repository

import (
	"context"

	"github.com/pghq/go-museum/museum/store"
)

// Update updates an item matching a filter
func (r *Repository) Update(ctx context.Context, collection string, filter store.Filter, item store.Snapper) (int, error) {
	return r.client.Update().In(collection).Filter(filter).Item(item.Snapshot()).Execute(ctx)
}
