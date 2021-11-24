package tea

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Response is the http response
type Response struct {
	header   http.Header
	body     interface{}
	status   int
	cachedAt time.Time
	cursor   *time.Time
}

// Send sends an HTTP response based on content type and body
func (r *Response) Send(w http.ResponseWriter, req *http.Request) {
	if r == nil || r.body == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	headers := w.Header()
	for k, v := range r.Headers(w, req) {
		headers[k] = v
	}

	bytes, ct, err := r.Bytes(req)
	if err != nil {
		SendHTTP(w, req, err)
		return
	}

	if ct != "" {
		headers.Set("Content-Type", ct)
	}

	w.WriteHeader(r.status)
	_, _ = w.Write(bytes)
}

// Body sets the response body
func (r *Response) Body(body interface{}) *Response {
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
func (r *Response) Bytes(req *http.Request) ([]byte, string, error) {
	if Accepts(req, "*/*") {
		if body, ok := r.body.([]byte); ok {
			return body, "", nil
		}

		if body, ok := r.body.(string); ok {
			return []byte(body), "", nil
		}
	}

	switch {
	case Accepts(req, "application/json"):
		bytes, err := json.Marshal(r.body)
		if err != nil {
			return nil, "", Error(err)
		}

		return bytes, "application/json", nil
	}

	return nil, "", NewBadRequest("unsupported content type")
}

// NewResponse creates a new response to be sent
func NewResponse() *Response {
	r := Response{
		status: http.StatusOK,
	}

	return &r
}
