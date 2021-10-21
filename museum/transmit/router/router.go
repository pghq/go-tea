// Copyright 2021 PGHQ. All Rights Reserved.
//
// Licensed under the GNU General Public License, Version 3 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package router provides a http router for handling requests.
package router

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/pghq/go-museum/museum/store/cache"
	"github.com/pghq/go-museum/museum/transmit/middleware"
)

const (
	// DefaultPositiveCacheTTL is the default positive cache time
	DefaultPositiveCacheTTL = 1 * time.Second

	// DefaultNegativeCacheTTL is the default negative cache time
	DefaultNegativeCacheTTL = 15 * time.Second
)

// Router is an instance of a mux based Router
type Router struct {
	mux   *mux.Router
	cache *cache.LRU
}

// Get adds a handler for the path using the GET http method
func (r *Router) Get(path string, handlerFunc http.HandlerFunc, opts ...Option) *Router {
	conf := Config{
		PositiveCacheTTL: DefaultPositiveCacheTTL,
		NegativeCacheTTL: DefaultNegativeCacheTTL,
	}

	for _, opt := range opts {
		opt.Apply(&conf)
	}

	if conf.PositiveCacheTTL != 0 || conf.NegativeCacheTTL != 0 {
		handlerFunc = cache.NewMiddleware(r.cache).Handle(handlerFunc).ServeHTTP
	}

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
func (r *Router) Middleware(middlewares ...middleware.Middleware) *Router {
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
		cache: cache.NewLRU(),
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

// Config for router
type Config struct {
	PositiveCacheTTL time.Duration
	NegativeCacheTTL time.Duration
}

// Option for router
type Option interface {
	Apply(conf *Config)
}

// cacheOption is an option for caching for get requests.
type cacheOption struct {
	positive time.Duration
	negative time.Duration
}

func (o cacheOption) Apply(conf *Config) {
	if conf != nil {
		conf.PositiveCacheTTL = o.positive
		conf.NegativeCacheTTL = o.negative
	}
}

// CacheFor creates a new router option for caching.
func CacheFor(positive, negative time.Duration) Option {
	return cacheOption{
		positive: positive,
		negative: negative,
	}
}
