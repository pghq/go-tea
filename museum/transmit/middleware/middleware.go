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

// Package middleware provides common request/response middlewares.
package middleware

import "net/http"

// Middleware represents a handler that is called to for example,
// modify request/response before or after the principal handler is called.
// Each Middleware is responsible for calling the next middleware in
// the chain (or not if continued execution is not desired)
type Middleware interface {
	Handle(http.Handler) http.Handler
}

// Func handles http middleware requests
type Func func(http.Handler) http.Handler

// Handle creates a http handler from the middleware
func (m Func) Handle(h http.Handler) http.Handler {
	return m(h)
}
