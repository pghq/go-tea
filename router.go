package tea

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hashicorp/go-version"

	"github.com/pghq/go-tea/health"
)

// Router is an instance of a mux based Router
type Router struct {
	version     *version.Version
	mux         *mux.Router
	middlewares []Middleware
}

// Route adds a handler for the path
func (r *Router) Route(method, endpoint string, handlerFunc http.HandlerFunc, middlewares ...Middleware) {
	s := r.mux.Methods(method, "OPTIONS").Subrouter()
	router := r
	s.HandleFunc(endpoint, func(w http.ResponseWriter, r *http.Request) {
		span := Start(r.Context(), "http")
		defer span.End()
		span.SetRequest(r)
		span.SetVersion(router.version)
		handlerFunc(w, r.WithContext(span))
	})
	for _, m := range append(r.middlewares, middlewares...) {
		s.Use(m.Handle)
	}
}

// At creates a sub-router for handling requests at a sub-path
func (r *Router) At(path string) *Router {
	s := r.mux.PathPrefix(path).Subrouter()
	for _, m := range r.middlewares {
		s.Use(m.Handle)
	}

	return &Router{
		mux: r.mux.PathPrefix(path).Subrouter(),
	}
}

// Middleware adds a handler to execute before/after the principle request handler
func (r *Router) Middleware(middlewares ...Middleware) {
	for _, m := range middlewares {
		r.mux.Use(m.Handle)
	}

	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// NewRouter constructs a new mux based Router
func NewRouter(semver string) *Router {
	r := Router{mux: mux.NewRouter().StrictSlash(true)}
	r.version, _ = version.NewVersion(semver)
	if r.version != nil {
		r.mux = r.mux.PathPrefix(fmt.Sprintf("/v%d", r.version.Segments()[0])).Subrouter()
	}
	r.mux.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
	r.mux.MethodNotAllowedHandler = http.HandlerFunc(MethodNotAllowedHandler)
	r.Route("GET", "/health/status", func(w http.ResponseWriter, r *http.Request) {
		Send(w, r, health.NewService(semver).Status())
	})
	r.Middleware(Trace())
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
