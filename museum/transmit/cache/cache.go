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
	"net/http"
	"net/url"
	"time"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

const (
	// DefaultSize is the default size of the Cache
	DefaultSize int = 1024

	// DefaultPositiveTTL is the default positive cache time
	DefaultPositiveTTL = 1 * time.Second

	// DefaultNegativeTTL is the default negative cache time
	DefaultNegativeTTL = 15 * time.Second
)

// Key encodes the key into a format consistent and compatible with the Cache.
func Key(key interface{}) (string, error) {
	if k, ok := key.(string); ok {
		return k, nil
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return "", errors.Wrap(err)
	}

	return fmt.Sprintf("%x", sha1.Sum(buf.Bytes())), nil
}

// RequestKey encodes an http request to a key
func RequestKey(r *http.Request, queries ...string) string {
	var key struct {
		Method string
		Header http.Header
		Scheme string
		Host   string
		Values url.Values
	}

	key.Method = r.Method
	key.Header = r.Header
	key.Scheme = r.URL.Scheme
	key.Host = r.URL.Host
	key.Values = r.URL.Query()

	if len(queries) != 0 {
		value := make(url.Values)
		for _, query := range queries {
			value[query] = key.Values[query]
		}
		key.Values = value
	}

	k, _ := Key(key)

	return k
}

// Item is an instance of a value in the lru Cache.
type Item struct {
	value    interface{}
	cachedAt time.Time
	ttl      time.Duration
}

// CachedAt gets the time the item was added to the cache
func (i *Item) CachedAt() time.Time {
	return i.cachedAt
}

// Value gets the raw object
func (i *Item) Value() interface{} {
	return i.value
}

// Config for router
type Config struct {
	PositiveTTL time.Duration
	NegativeTTL time.Duration
	Queries     []string
}

// Option for router
type Option interface {
	Apply(conf *Config)
}

// cacheOption is an option for caching for get requests.
type cacheOption struct {
	positive time.Duration
	negative time.Duration
	queries  []string
}

func (o cacheOption) Apply(conf *Config) {
	if conf != nil {
		if o.negative != 0 {
			conf.NegativeTTL = o.negative
		}

		if o.positive != 0 {
			conf.PositiveTTL = o.positive
		}

		if len(o.queries) > 0 {
			conf.Queries = o.queries
		}
	}
}

// PositiveFor creates a new router option for positive ttl for caching.
func PositiveFor(ttl time.Duration) Option {
	return cacheOption{
		positive: ttl,
	}
}

// NegativeFor creates a new router option for negative ttl for caching.
func NegativeFor(ttl time.Duration) Option {
	return cacheOption{
		negative: ttl,
	}
}

// Use creates a new router option for which queries are used for cache key.
func Use(queries ...string) Option {
	return cacheOption{
		queries: queries,
	}
}
