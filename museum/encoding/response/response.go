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

// Package response provides resources for replying to requests
package response

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/encoding/request"
)

// Send sends an HTTP response based on content type and body
func Send(w http.ResponseWriter, r *http.Request, body *Builder) {
	if body == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var encoder interface{ Encode(v interface{}) error }
	var contentType string

	switch {
	case request.Accepts(r, "application/json"):
		contentType = "application/json"
		encoder = json.NewEncoder(w)
	default:
		errors.EmitHTTP(w, r, errors.BadRequest(errors.New("unsupported MIME type")))
		return
	}

	w.Header().Set("Content-Type", contentType)
	content := body.response()
	if err := encoder.Encode(&content); err != nil {
		errors.EmitHTTP(w, r, err)
		return
	}
}

// response is the expected body contents for app requests.
type response struct {
	Data interface{} `json:"data"`
	CachedAt *time.Time `json:"cachedAt,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}

// Builder is an instance of a response builder
type Builder struct{
	data interface{}
	cachedAt time.Time
	cursor string
}

// Cached adds a cache time to the response
func (b *Builder) Cached(at time.Time) *Builder {
	b.cachedAt = at

	return b
}

// Cursor adds a cursor to the response
func (b *Builder) Cursor(cursor string) *Builder {
	b.cursor = cursor

	return b
}

func (b *Builder) response() *response {
	r := response{
		Data: b.data,
	}

	if b.cursor != ""{
		r.Cursor = base64.StdEncoding.EncodeToString([]byte(b.cursor))
	}

	if !b.cachedAt.IsZero(){
		r.CachedAt = &b.cachedAt
	}

	return &r
}

// New creates a new response to be sent
func New(data interface{}) *Builder {
	return &Builder{
		data: data,
	}
}
