package tea

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"net/http"
	"net/url"
	"time"

	lru "github.com/hashicorp/golang-lru"

	"github.com/pghq/go-tea/internal"
	"github.com/pghq/go-tea/internal/clock"
)

const (
	// DefaultCacheSize is the default size of the Cache
	DefaultCacheSize int = 1024

	// DefaultPositiveCacheTTL is the default positive cache time
	DefaultPositiveCacheTTL = 1 * time.Second

	// DefaultNegativeCacheTTL is the default negative cache time
	DefaultNegativeCacheTTL = 15 * time.Second
)

// CacheKey encodes the key into a format consistent and compatible with the Cache.
func CacheKey(key interface{}) (string, error) {
	if k, ok := key.(string); ok {
		return k, nil
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return "", Error(err)
	}

	return fmt.Sprintf("%x", sha1.Sum(buf.Bytes())), nil
}

// CacheRequestKey encodes an http request to a key
func CacheRequestKey(r *http.Request, queries ...string) string {
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

	k, _ := CacheKey(key)

	return k
}

// CacheItem is an instance of a value in the lru Cache.
type CacheItem struct {
	value    interface{}
	cachedAt time.Time
	ttl      time.Duration
}

// CachedAt gets the time the item was added to the cache
func (i *CacheItem) CachedAt() time.Time {
	return i.cachedAt
}

// Value gets the raw object
func (i *CacheItem) Value() interface{} {
	return i.value
}

// CacheConfig for router
type CacheConfig struct {
	PositiveTTL time.Duration
	NegativeTTL time.Duration
	Queries     []string
}

// CacheOption for router
type CacheOption interface {
	Apply(conf *CacheConfig)
}

// cacheOption is an option for caching for get requests.
type cacheOption struct {
	positive time.Duration
	negative time.Duration
	queries  []string
}

func (o cacheOption) Apply(conf *CacheConfig) {
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

// PositiveCacheFor creates a new router option for positive ttl for caching.
func PositiveCacheFor(ttl time.Duration) CacheOption {
	return cacheOption{
		positive: ttl,
	}
}

// NegativeCacheFor creates a new router option for negative ttl for caching.
func NegativeCacheFor(ttl time.Duration) CacheOption {
	return cacheOption{
		negative: ttl,
	}
}

// UseCacheQuery creates a new router option for which queries are used for cache key.
func UseCacheQuery(queries ...string) CacheOption {
	return cacheOption{
		queries: queries,
	}
}

// LRU is an instance of a lru based Cache.
type LRU struct {
	lru   *lru.Cache
	clock internal.Clock
}

// Insert adds a value to the cache
func (c *LRU) Insert(k interface{}, v interface{}, ttl time.Duration) error {
	key, err := CacheKey(k)
	if err != nil {
		return Error(err)
	}

	c.lru.Add(key, &CacheItem{
		value:    v,
		cachedAt: c.clock.Now(),
		ttl:      ttl,
	})

	return nil
}

// Remove deletes a value from the cache
func (c *LRU) Remove(k interface{}) error {
	key, err := CacheKey(k)
	if err != nil {
		return Error(err)
	}

	c.lru.Remove(key)

	return nil
}

// Get attempts to retrieve a value from the cache
func (c *LRU) Get(k interface{}) (*CacheItem, error) {
	key, err := CacheKey(k)
	if err != nil {
		return nil, Error(err)
	}

	v, ok := c.lru.Get(key)
	if !ok {
		return nil, NewNoContent("key not found in cache")
	}

	item, ok := v.(*CacheItem)
	if !ok {
		return nil, NewError("unexpected value in Cache")
	}

	if time.Since(item.cachedAt) > item.ttl {
		c.lru.Remove(key)
		err := fmt.Errorf("key expired at %s", item.cachedAt.Add(item.ttl).Format(time.RFC3339Nano))
		return nil, NoContent(err)
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
	c, _ := lru.New(DefaultCacheSize)
	return &LRU{lru: c, clock: clock.New(time.Now())}
}
