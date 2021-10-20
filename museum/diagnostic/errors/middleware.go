package errors

import (
	"net/http"

	sentryhttp "github.com/getsentry/sentry-go/http"
)

// Middleware is an implementation of the sentry middleware
type Middleware struct {
	sentry *sentryhttp.Handler
}

// Handle provides an http handler for handling exceptions
func (m *Middleware) Handle(next http.Handler) http.Handler {
	return m.sentry.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// NewMiddleware constructs a new middleware that handles exceptions
func NewMiddleware() *Middleware {
	m := &Middleware{
		sentry: sentryhttp.New(sentryhttp.Options{}),
	}

	return m
}
