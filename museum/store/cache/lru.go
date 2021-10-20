package cache

import (
	"fmt"
	"time"

	"github.com/hashicorp/golang-lru"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/internal"
	"github.com/pghq/go-museum/museum/internal/clock"
)

// LRU is an instance of a lru based Cache.
type LRU struct {
	lru   *lru.Cache
	clock internal.Clock
}

// Insert adds a value to the cache
func (c *LRU) Insert(k interface{}, v interface{}, ttl time.Duration) error {
	key, err := Key(k)
	if err != nil {
		return errors.Wrap(err)
	}

	c.lru.Add(key, &Item{
		value:    v,
		cachedAt: c.clock.Now(),
		ttl:      ttl,
	})

	return nil
}

// Remove deletes a value from the cache
func (c *LRU) Remove(k interface{}) error {
	key, err := Key(k)
	if err != nil {
		return errors.Wrap(err)
	}

	c.lru.Remove(key)

	return nil
}

// Get attempts to retrieve a value from the cache
func (c *LRU) Get(k interface{}) (*Item, error) {
	key, err := Key(k)
	if err != nil {
		return nil, errors.Wrap(err)
	}

	v, ok := c.lru.Get(key)
	if !ok {
		return nil, errors.NewNoContent("key not found in cache")
	}

	item, ok := v.(*Item)
	if !ok {
		return nil, errors.New("unexpected value in Cache")
	}

	if time.Since(item.cachedAt) > item.ttl {
		c.lru.Remove(key)
		err := fmt.Errorf("key expired at %s", item.cachedAt.Add(item.ttl).Format(time.RFC3339Nano))
		return nil, errors.NoContent(err)
	}

	return item, nil
}

// Len gets the number of items in the cache
func (c *LRU) Len() int {
	return c.lru.Len()
}

// SetCapacity sets the maximum number of items in the cache
func (c *LRU) SetCapacity(capacity int) {
	c.lru.Resize(capacity)
}

func (c *LRU) setClock(clock internal.Clock) *LRU {
	c.clock = clock

	return c
}

// NewLRU constructs a new lru Cache instance.
func NewLRU() *LRU {
	// we know at this point the Cache size is positive
	c, _ := lru.New(int(DefaultCacheSize))
	return &LRU{lru: c, clock: clock.New(time.Now())}
}
