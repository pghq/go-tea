package tea

import (
	"net/http"
	"strings"

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
