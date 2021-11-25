package tea

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Router is an instance of a mux based Router
type Router struct {
	mux   *mux.Router
}

// Get adds a handler for the path using the GET http method
func (r *Router) Get(path string, handlerFunc http.HandlerFunc) *Router {
	r.mux.HandleFunc(path, handlerFunc).
		Methods("GET", "OPTIONS")

	return r
}

// Put adds a handler for the path using the PUT http method
func (r *Router) Put(path string, handlerFunc http.HandlerFunc) *Router {
	r.mux.HandleFunc(path, handlerFunc).Methods("PUT", "OPTIONS")

	return r
}

// Post adds a handler for the path using the POST http method
func (r *Router) Post(path string, handlerFunc http.HandlerFunc) *Router {
	r.mux.HandleFunc(path, handlerFunc).Methods("POST", "OPTIONS")

	return r
}

// Patch adds a handler for the path using the PATCH http method
func (r *Router) Patch(path string, handlerFunc http.HandlerFunc) *Router {
	r.mux.HandleFunc(path, handlerFunc).Methods("PATCH", "OPTIONS")

	return r
}

// Delete adds a handler for the path using the DELETE http method
func (r *Router) Delete(path string, handlerFunc http.HandlerFunc) *Router {
	r.mux.HandleFunc(path, handlerFunc).Methods("DELETE", "OPTIONS")

	return r
}

// At creates a sub-router for handling requests at a sub-path
func (r *Router) At(path string) *Router {
	return &Router{
		mux: r.mux.PathPrefix(path).Subrouter(),
	}
}

// Middleware adds a handler to execute before/after the principle request handler
func (r *Router) Middleware(middlewares ...Middleware) *Router {
	for _, m := range middlewares {
		r.mux.Use(m.Handle)
	}

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
