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

// Package cors provides resources for handling cors requests.
package cors

import (
	"net/http"
	"strings"

	"github.com/rs/cors"
)

// Middleware is an implementation of the CORS middleware
// providing method, origin, and credential allowance
type Middleware struct {
	cors *cors.Cors
}

// Handle provides an http handler for handling CORS
func (m *Middleware) Handle(next http.Handler) http.Handler {
	return m.cors.Handler(next)
}

// New constructs a new middleware that handles CORS
func New(origins ...string) *Middleware {
	ac, origins := credentials(origins)
	m := &Middleware{
		cors: cors.New(cors.Options{
			AllowedOrigins:   origins,
			AllowCredentials: ac,
			AllowedMethods:   []string{
				http.MethodOptions,
				http.MethodHead,
				http.MethodGet,
				http.MethodPost,
				http.MethodDelete,
				http.MethodPut,
			},
			AllowedHeaders:   []string{"*"},
		}),
	}

	return m
}

// credentials is a helper function to determine if credentials are allowed
func credentials(origins []string) (bool, []string ){
	var aos []string
	credentials := true
	for _, ao := range origins{
		origin := strings.Trim(ao, " ")
		if strings.Contains(origin, "*"){
			credentials = false
		}

		if origin != ""{
			aos = append(aos, origin)
		}
	}
	credentials = credentials && len(aos) > 0
	return credentials, aos
}
