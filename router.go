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
	mux         *mux.Router
	routes      *mux.Router
	middlewares []Middleware
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
	for _, m := range middlewares {
		r.routes.Use(m.Handle)
	}

	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// NewRouter constructs a new mux based Router
func NewRouter(semver string) *Router {
	r := Router{mux: mux.NewRouter().StrictSlash(true)}
	r.routes = r.mux

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
	r.Middleware(trail.NewTraceMiddleware(v.String(), nil))
	return &r
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

// HTTPCommand creates a http command handler from a command
func HTTPCommand[command any](fn func(context.Context, command) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req command
		if err := Parse(w, r, &req); err != nil {
			Send(w, r, err)
			return
		}

		if err := fn(r.Context(), req); err != nil {
			Send(w, r, err)
			return
		}

		Send(w, r, nil)
	}
}

// HTTPQuery creates a http query handler from a query
func HTTPQuery[query any, response any](fn func(context.Context, query) (response, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req query
		if err := parseURL(r, &req); err != nil {
			Send(w, r, err)
			return
		}

		resp, err := fn(r.Context(), req)
		if err != nil {
			Send(w, r, err)
			return
		}

		Send(w, r, resp)
	}
}
