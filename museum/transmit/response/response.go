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
	"github.com/pghq/go-museum/museum/transmit/request"
)

// Send sends an HTTP response based on content type and body
func Send(w http.ResponseWriter, r *http.Request, body *Builder) {
	if body == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if !request.Accepts(r, "application/json") {
		errors.SendHTTP(w, r, errors.BadRequest(errors.New("unsupported MIME type")))
		return
	}

	if body.IsRaw() {
		header := w.Header()
		for k, v := range body.raw.header {
			header[k] = v
		}
		w.WriteHeader(body.raw.status)
		_, _ = w.Write(body.raw.data)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(body.Response()); err != nil {
		errors.SendHTTP(w, r, err)
		return
	}
}

// Builder is an instance of a response builder
type Builder struct {
	data interface{}
	raw  struct {
		data   []byte
		header http.Header
		status int
	}
	cachedAt time.Time
	cursor   *time.Time
}

// Cached adds a cache time to the response
func (b *Builder) Cached(at time.Time) *Builder {
	b.cachedAt = at

	return b
}

// Cursor adds a cursor to the response
func (b *Builder) Cursor(cursor *time.Time) *Builder {
	b.cursor = cursor

	return b
}

// Response is the body to be sent
func (b *Builder) Response() interface{} {
	var r struct {
		Data     interface{} `json:"data"`
		CachedAt *time.Time  `json:"cachedAt,omitempty"`
		Cursor   string      `json:"cursor,omitempty"`
	}

	r.Data = b.data
	if b.cursor != nil && !b.cursor.IsZero() {
		ds := b.cursor.Format(time.RFC3339Nano)
		r.Cursor = base64.StdEncoding.EncodeToString([]byte(ds))
	}

	if !b.cachedAt.IsZero() {
		r.CachedAt = &b.cachedAt
	}

	return &r
}

// IsRaw checks if the response is raw
func (b *Builder) IsRaw() bool {
	return len(b.raw.data) > 0
}

// New creates a new response to be sent
func New(data interface{}) *Builder {
	b := Builder{
		data: data,
	}

	b.raw.status = http.StatusOK
	return &b
}

// NewRaw creates a new raw response to be sent
func NewRaw(header http.Header, status int, data []byte) *Builder {
	b := Builder{
		data: data,
	}

	b.raw.data = data
	b.raw.header = header
	b.raw.status = status
	return &b
}
