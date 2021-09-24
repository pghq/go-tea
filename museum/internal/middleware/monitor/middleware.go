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

// Package monitor provides a middleware for monitoring errors.
package monitor

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

// New constructs a new middleware that handles exceptions
func New() *Middleware {
	m := &Middleware{
		sentry: sentryhttp.New(sentryhttp.Options{}),
	}

	return m
}