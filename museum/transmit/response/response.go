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
	"fmt"
	"net/http"
	"time"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/transmit/request"
)

// Response is the http response
type Response struct {
	header http.Header
	body interface{}
	status int
	cachedAt time.Time
	cursor   *time.Time
}

// Send sends an HTTP response based on content type and body
func (r *Response) Send(w http.ResponseWriter, req *http.Request){
	if r == nil || r.body == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	headers := w.Header()
	for k, v := range r.Headers(w, req) {
		headers[k] = v
	}

	bytes, ct, err := r.Bytes(req)
	if err != nil{
		errors.SendHTTP(w, req, err)
		return
	}

	if ct != ""{
		headers.Set("Content-Type", ct)
	}

	w.WriteHeader(r.status)
	_, _ = w.Write(bytes)
}

// Body sets the response body
func (r *Response) Body(body interface{}) *Response{
	r.body = body
	return r
}

// Status sets the http status code
func (r *Response) Status(code int) *Response {
	r.status = code

	return r
}

// Cached adds a cache time to the response
func (r *Response) Cached(at time.Time) *Response {
	r.cachedAt = at

	return r
}

// Cursor adds a cursor to the response
func (r *Response) Cursor(cursor *time.Time) *Response {
	r.cursor = cursor

	return r
}

// Header sets http response headers
func (r *Response) Header(header http.Header) *Response {
	r.header = header

	return r
}

// Headers gets the http response headers
func (r *Response) Headers(w http.ResponseWriter, req *http.Request) http.Header {
	header := w.Header().Clone()
	for k, v := range r.header {
		header[k] = v
	}

	if r.cursor != nil && !r.cursor.IsZero() {
		ds := r.cursor.Format(time.RFC3339Nano)
		link := req.URL
		query := link.Query()
		query.Set("after", base64.StdEncoding.EncodeToString([]byte(ds)))
		link.RawQuery = query.Encode()
		header.Add("Link", fmt.Sprintf("<%s>", link.String()))
	}

	if !r.cachedAt.IsZero() {
		header.Add("Cached-At", r.cachedAt.Format(time.RFC3339Nano))
	}

	return header
}

// Bytes gets the response as bytes based on origin
func (r *Response) Bytes(req *http.Request) ([]byte, string, error){
	if request.Accepts(req, "*/*"){
		if body, ok := r.body.([]byte); ok{
			return body, "", nil
		}

		if body, ok := r.body.(string); ok{
			return []byte(body), "", nil
		}
	}

	switch {
	case request.Accepts(req, "application/json"):
		bytes, err := json.Marshal(r.body)
		if err != nil {
			return nil, "", errors.Wrap(err)
		}

		return bytes, "application/json", nil
	}

	return nil, "", errors.NewBadRequest("unsupported content type")
}

// New creates a new response to be sent
func New() *Response {
	r := Response{
		status: http.StatusOK,
	}

	return &r
}
