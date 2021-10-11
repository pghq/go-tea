// Copyright 2021 PGHQ. All Rights Reserved.
//
// Licensed under the GNU General Public License, Version 3 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package cache provides a caching for the app.
package cache

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"time"

	hashicorp "github.com/hashicorp/golang-lru"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/internal"
	"github.com/pghq/go-museum/museum/internal/clock"
)

const (
	// DefaultCacheSize is the default size of the Cache
	DefaultCacheSize uint = 1024
)

// encodeKey encodes the key into a format consistent and compatible with the Cache.
func encodeKey(key interface{}) (string, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return "", errors.Wrap(err)
	}

	return fmt.Sprintf("%x", sha1.Sum(buf.Bytes())), nil
}

// Cache is an instance of a lru based Cache.
type Cache struct{
	lru *hashicorp.Cache
	clock internal.Clock
}

// Insert adds a value to the cache
func (c *Cache) Insert(k interface{}, v interface{}, ttl time.Duration) error {
	key, err := encodeKey(k)
	if err != nil {
		return errors.Wrap(err)
	}

	c.lru.Add(key, &Item{
		value:     v,
		cachedAt: c.clock.Now(),
		ttl:      ttl,
	})

	return nil
}

// Remove deletes a value from the cache
func (c *Cache) Remove(k interface{}) error {
	key, err := encodeKey(k)
	if err != nil {
		return errors.Wrap(err)
	}

	c.lru.Remove(key)

	return nil
}

// Get attempts to retrieve a value from the cache
func (c *Cache) Get(k interface{}) (*Item, error) {
	key, err := encodeKey(k)
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
func (c *Cache) Len() int{
	return c.lru.Len()
}

// SetCapacity sets the maximum number of items in the cache
func (c *Cache) SetCapacity(capacity uint){
	c.lru.Resize(int(capacity))
}

func (c *Cache) setClock(clock internal.Clock) *Cache {
	c.clock = clock

	return c
}

// New constructs a new lru Cache instance.
func New() *Cache {
	// we know at this point the Cache size is positive
	c, _ := hashicorp.New(int(DefaultCacheSize))
	return &Cache{lru: c, clock: clock.New(time.Now())}
}

// Item is an instance of a value in the lru Cache.
type Item struct {
	value     interface{}
	cachedAt time.Time
	ttl      time.Duration
}

// CachedAt gets the time the item was added to the cace
func (i *Item) CachedAt() time.Time{
	return i.cachedAt
}

// Value gets the raw object
func (i *Item) Value() interface{}{
	return i.value
}