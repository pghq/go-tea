package tea

import (
	"net/http"
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
