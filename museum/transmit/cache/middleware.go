package cache

import (
	"net/http"
	"time"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/transmit/response"
)

// Middleware is a middleware for caching responses
type Middleware struct {
	cache *LRU
	opts  []Option
}

// With sets new cache options for the middleware
func (m *Middleware) With(opts ...Option) *Middleware {
	m.opts = append(m.opts, opts...)
	return m
}

// Handle provides an http handler for handling CORS
func (m *Middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.cache == nil {
			next.ServeHTTP(w, r)
			return
		}

		conf := Config{
			PositiveTTL: DefaultPositiveTTL,
			NegativeTTL: DefaultNegativeTTL,
		}

		for _, opt := range m.opts {
			opt.Apply(&conf)
		}

		key := RequestKey(r, conf.Queries...)
		i, err := m.cache.Get(key)
		if err != nil {
			w := NewResponseWatcher(m.cache, &conf, w, key)
			if errors.IsFatal(err) {
				errors.SendHTTP(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		resp, err := NewCachedResponse(i)
		if err != nil {
			errors.SendHTTP(w, r, err)
			return
		}

		resp.Send(w, r)
	})
}

// NewMiddleware creates a new lru middleware instance
func NewMiddleware(cache *LRU) *Middleware {
	m := &Middleware{
		cache: cache,
	}

	return m
}

// ResponseWatcher is a cached instance of the http.ResponseWriter
type ResponseWatcher struct {
	status   int
	key      string
	cache    *LRU
	positive time.Duration
	negative time.Duration
	w        http.ResponseWriter
}

func (r *ResponseWatcher) WriteHeader(statusCode int) {
	r.status = statusCode
	r.w.WriteHeader(statusCode)
}

func (r *ResponseWatcher) Header() http.Header {
	return r.w.Header()
}

func (r *ResponseWatcher) Write(data []byte) (int, error) {
	ttl := r.positive
	if r.status >= http.StatusBadRequest {
		ttl = r.negative
	}

	var v struct {
		Data   []byte
		Header http.Header
		Status int
	}

	v.Data = data
	v.Header = r.Header()
	v.Status = r.status

	_ = r.cache.Insert(r.key, v, ttl)

	return r.w.Write(data)
}

// NewResponseWatcher creates a new cached response watcher
func NewResponseWatcher(cache *LRU, conf *Config, w http.ResponseWriter, key string) *ResponseWatcher {
	r := ResponseWatcher{
		status: http.StatusOK,
		key:    key,
		cache:  cache,
		w:      w,
	}

	if conf != nil {
		r.positive = conf.PositiveTTL
		r.negative = conf.NegativeTTL
	}

	return &r
}

// NewCachedResponse creates a new response.Builder from the cache item
func NewCachedResponse(i *Item) (*response.Response, error) {
	v, ok := i.Value().(struct {
		Data   []byte
		Header http.Header
		Status int
	})

	if !ok {
		return nil, errors.New("unexpected cache value")
	}

	r := response.New().
		Header(v.Header).
		Status(v.Status).
		Body(v.Data).
		Cached(i.CachedAt())

	return r, nil
}
