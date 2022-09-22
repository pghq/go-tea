package tea

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hashicorp/go-version"

	"github.com/pghq/go-tea/health"
	"github.com/pghq/go-tea/trail"
)

// Router is an instance of a mux based Router
type Router struct {
	mux           *mux.Router
	routes        *mux.Router
	middlewares   []Middleware
	servicePrefix string
}

// Route adds a handler for the http method and endpoint
func (r *Router) Route(method, endpoint string, handlerFunc http.HandlerFunc, middlewares ...Middleware) {
	s := r.routes.Methods(method, "OPTIONS").Subrouter()
	s.HandleFunc(endpoint, handlerFunc)
	for _, m := range append(r.middlewares, middlewares...) {
		s.Use(m.Handle)
	}
}

// Middleware adds a handler to execute before/after the principle request handler
func (r *Router) Middleware(middlewares ...Middleware) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// NewRouter constructs a new mux based Router
func NewRouter(semver string, opts ...RouterOption) *Router {
	r := Router{mux: mux.NewRouter().StrictSlash(true)}
	for _, opt := range opts {
		opt(&r)
	}

	r.routes = r.mux
	if r.servicePrefix != "" {
		r.routes = r.routes.PathPrefix(r.servicePrefix).Subrouter()
	}

	hc := health.NewService(semver)
	r.Route("GET", "/health/status", func(w http.ResponseWriter, r *http.Request) {
		Send(w, r, hc.Status())
	})

	v, _ := version.NewVersion(semver)
	if v != nil {
		r.routes = r.routes.PathPrefix(fmt.Sprintf("/v%d", v.Segments()[0])).Subrouter()
	}
	r.mux.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
	r.mux.MethodNotAllowedHandler = http.HandlerFunc(MethodNotAllowedHandler)
	r.Middleware(MiddlewareFunc(trail.NewTraceMiddleware(v.String(), true)))
	return &r
}

// RouterOption is a handler for configuring the router
type RouterOption func(r *Router)

// WithServicePrefix creates a service prefix option
func WithServicePrefix(prefix string) RouterOption {
	return func(r *Router) {
		r.servicePrefix = prefix
	}
}

// NotFoundHandler is a custom handler for not found requests
func NotFoundHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(http.StatusText(http.StatusNotFound)))
}

// MethodNotAllowedHandler is a custom handler for method not allowed requests
func MethodNotAllowedHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	_, _ = w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
}

// Rte a request
func Rte[Q any, R any](h RouteHandler, method, endpoint string, fn func(context.Context, Q) (R, error), middlewares ...Middleware) {
	h.Route(method, endpoint, func(w http.ResponseWriter, r *http.Request) {
		var req Q
		if err := Parse(w, r, &req); err != nil {
			Send(w, r, err)
			return
		}

		resp, err := fn(r.Context(), req)
		if err != nil {
			Send(w, r, err)
			return
		}

		Send(w, r, resp)
	}, middlewares...)
}

// RouteHandler a handler for routing http requests
type RouteHandler interface {
	Route(method, endpoint string, handlerFunc http.HandlerFunc, middlewares ...Middleware)
}
