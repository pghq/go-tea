package tea

import (
	"net/http"
	"strings"
	"time"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/rs/cors"
)

// Middleware represents a handler that is called to for example,
// modify request/response before or after the principal handler is called.
// Each Middleware is responsible for calling the next middleware in
// the chain (or not if continued execution is not desired)
type Middleware interface {
	Handle(http.Handler) http.Handler
}

// MiddlewareFunc handles http middleware requests
type MiddlewareFunc func(http.Handler) http.Handler

// Handle creates a http handler from the middleware
func (m MiddlewareFunc) Handle(h http.Handler) http.Handler {
	return m(h)
}

// SentryMiddleware is an implementation of the sentry middleware
type SentryMiddleware struct {
	sentry *sentryhttp.Handler
}

// Handle provides an http handler for handling exceptions
func (m *SentryMiddleware) Handle(next http.Handler) http.Handler {
	return m.sentry.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// NewSentryMiddleware constructs a new middleware that handles exceptions
func NewSentryMiddleware() *SentryMiddleware {
	m := &SentryMiddleware{
		sentry: sentryhttp.New(sentryhttp.Options{}),
	}

	return m
}

// CORSMiddleware is an implementation of the CORS middleware
// providing method, origin, and credential allowance
type CORSMiddleware struct {
	cors *cors.Cors
}

// Handle provides an http handler for handling CORS
func (m *CORSMiddleware) Handle(next http.Handler) http.Handler {
	return m.cors.Handler(next)
}

// NewCORSMiddleware constructs a new middleware that handles CORS
func NewCORSMiddleware(origins ...string) *CORSMiddleware {
	ac, origins := credentials(origins)
	m := &CORSMiddleware{
		cors: cors.New(cors.Options{
			AllowedOrigins:   origins,
			AllowCredentials: ac,
			AllowedMethods: []string{
				http.MethodOptions,
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodDelete,
				http.MethodPut,
			},
			AllowedHeaders: []string{"*"},
		}),
	}

	return m
}

// CacheMiddleware is a middleware for caching responses
type CacheMiddleware struct {
	cache *LRU
	opts  []CacheOption
}

// With sets new cache options for the middleware
func (m *CacheMiddleware) With(opts ...CacheOption) *CacheMiddleware {
	m.opts = append(m.opts, opts...)
	return m
}

// Handle provides an http handler for handling CORS
func (m *CacheMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.cache == nil {
			next.ServeHTTP(w, r)
			return
		}

		conf := CacheConfig{
			PositiveTTL: DefaultPositiveCacheTTL,
			NegativeTTL: DefaultNegativeCacheTTL,
		}

		for _, opt := range m.opts {
			opt.Apply(&conf)
		}

		key := CacheRequestKey(r, conf.Queries...)
		i, err := m.cache.Get(key)
		if err != nil {
			w := NewCacheResponseWatcher(m.cache, &conf, w, key)
			if IsFatal(err) {
				SendHTTP(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		resp, err := NewCachedResponse(i)
		if err != nil {
			SendHTTP(w, r, err)
			return
		}

		resp.Send(w, r)
	})
}

// NewCacheMiddleware creates a new lru middleware instance
func NewCacheMiddleware(cache *LRU) *CacheMiddleware {
	m := &CacheMiddleware{
		cache: cache,
	}

	return m
}

// CacheResponseWatcher is a cached instance of the http.ResponseWriter
type CacheResponseWatcher struct {
	status   int
	key      string
	cache    *LRU
	positive time.Duration
	negative time.Duration
	w        http.ResponseWriter
}

func (r *CacheResponseWatcher) WriteHeader(statusCode int) {
	r.status = statusCode
	r.w.WriteHeader(statusCode)
}

func (r *CacheResponseWatcher) Header() http.Header {
	return r.w.Header()
}

func (r *CacheResponseWatcher) Write(data []byte) (int, error) {
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

// NewCacheResponseWatcher creates a new cached response watcher
func NewCacheResponseWatcher(cache *LRU, conf *CacheConfig, w http.ResponseWriter, key string) *CacheResponseWatcher {
	r := CacheResponseWatcher{
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
func NewCachedResponse(i *CacheItem) (*Response, error) {
	v, ok := i.Value().(struct {
		Data   []byte
		Header http.Header
		Status int
	})

	if !ok {
		return nil, NewError("unexpected cache value")
	}

	r := NewResponse().
		Header(v.Header).
		Status(v.Status).
		Body(v.Data).
		Cached(i.CachedAt())

	return r, nil
}

// credentials is a helper function to determine if credentials are allowed
func credentials(origins []string) (bool, []string) {
	var aos []string
	credentials := true
	for _, ao := range origins {
		origin := strings.Trim(ao, " ")
		if strings.Contains(origin, "*") {
			credentials = false
		}

		if origin != "" {
			aos = append(aos, origin)
		}
	}
	credentials = credentials && len(aos) > 0
	return credentials, aos
}
