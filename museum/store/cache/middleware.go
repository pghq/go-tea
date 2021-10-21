package cache

import (
	"net/http"
	"time"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/transmit/response"
)

// Middleware is a middleware for caching responses
type Middleware struct {
	cache    *LRU
	positive time.Duration
	negative time.Duration
}

// Positive sets the positive ttl for cached responses
func (m *Middleware) Positive(ttl time.Duration) *Middleware {
	m.positive = ttl
	return m
}

// Negative sets the negative ttl for cached responses
func (m *Middleware) Negative(ttl time.Duration) *Middleware {
	m.negative = ttl
	return m
}

// Handle provides an http handler for handling CORS
func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.cache == nil {
			next.ServeHTTP(w, r)
			return
		}

		key := RequestKey(r)
		i, err := m.cache.Get(key)
		if err != nil {
			resp := NewCachedResponse(m.cache, key, w).
				Positive(m.positive).
				Negative(m.negative)

			if errors.IsFatal(err) {
				errors.SendHTTP(resp, r, err)
				return
			}

			next.ServeHTTP(resp, r)
			return
		}

		response.Send(w, r, response.New(i.Value()).Cached(i.CachedAt()))
	})
}

// NewMiddleware creates a new lru middleware instance
func NewMiddleware(cache *LRU) *Middleware {
	m := &Middleware{
		cache: cache,
	}

	return m
}

// CachedResponse is a cached instance of the http.ResponseWriter
type CachedResponse struct {
	key      string
	status   int
	cache    *LRU
	positive time.Duration
	negative time.Duration
	http.ResponseWriter
}

// Positive sets the positive ttl for cached responses
func (r *CachedResponse) Positive(ttl time.Duration) *CachedResponse {
	r.positive = ttl
	return r
}

// Negative sets the negative ttl for cached responses
func (r *CachedResponse) Negative(ttl time.Duration) *CachedResponse {
	r.negative = ttl
	return r
}

func (r *CachedResponse) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *CachedResponse) Write(data []byte) (int, error) {
	ttl := r.positive
	if r.status >= http.StatusBadRequest {
		ttl = r.negative
	}

	_ = r.cache.Insert(r.key, string(data), ttl)

	return r.ResponseWriter.Write(data)
}

// NewCachedResponse creates a new cached response
func NewCachedResponse(cache *LRU, key string, w http.ResponseWriter) *CachedResponse {
	r := CachedResponse{
		key:            key,
		cache:          cache,
		ResponseWriter: w,
	}

	return &r
}
