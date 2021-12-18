package tea

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Router is an instance of a mux based Router
type Router struct {
	mux         *mux.Router
	middlewares []Middleware
}

// Route adds a handler for the path
func (r *Router) Route(method, endpoint string, handlerFunc http.HandlerFunc, middlewares ...Middleware) *Router {
	s := r.mux.Methods(method, "OPTIONS").Subrouter()
	s.HandleFunc(endpoint, handlerFunc)
	for _, m := range append(r.middlewares, middlewares...) {
		s.Use(m.Handle)
	}

	return r
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
func (r *Router) Middleware(middlewares ...Middleware) *Router {
	for _, m := range middlewares {
		r.mux.Use(m.Handle)
	}

	r.middlewares = append(r.middlewares, middlewares...)
	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// NewRouter constructs a new mux based Router
func NewRouter(version int) *Router {
	r := Router{
		mux: mux.NewRouter().
			StrictSlash(true).
			PathPrefix(fmt.Sprintf("/v%d", version)).
			Subrouter(),
	}

	r.mux.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
	r.mux.MethodNotAllowedHandler = http.HandlerFunc(MethodNotAllowedHandler)

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
