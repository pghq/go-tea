package test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type RequestBuilder struct {
	t      *testing.T
	path   string
	method string
	body   string
	router struct {
		method string
		path   string
	}
	response struct {
		code int
		body string
	}
}

func (b *RequestBuilder) Method(method string) *RequestBuilder {
	b.method = method
	if b.router.method == "" {
		b.router.method = method
	}

	return b
}

func (b *RequestBuilder) Path(path string) *RequestBuilder {
	b.path = path

	return b
}

func (b *RequestBuilder) Body(body string) *RequestBuilder {
	b.body = body
	return b
}

func (b *RequestBuilder) ExpectStatus(code int) *RequestBuilder {
	b.response.code = code
	return b
}

func (b *RequestBuilder) ExpectRoute(path string) *RequestBuilder {
	b.router.path = path
	return b
}

func (b *RequestBuilder) ExpectedRoute() string {
	return b.router.path
}

func (b *RequestBuilder) ExpectMethod(method string) *RequestBuilder {
	b.router.method = method
	return b
}

func (b *RequestBuilder) ExpectedMethod() string {
	return b.router.method
}

func (b *RequestBuilder) ExpectResponse(body string) *RequestBuilder {
	b.response.body = body
	return b
}

func (b *RequestBuilder) Request(base string) *http.Request {
	b.t.Helper()
	var body io.Reader
	if b.body != "" {
		body = io.NopCloser(strings.NewReader(b.body))
	}

	u := fmt.Sprintf("%s/%s", base, strings.TrimPrefix(b.path, "/"))
	req, err := http.NewRequest(b.method, u, body)
	assert.Nil(b.t, err)

	return req
}

func (b *RequestBuilder) Response() *http.Response {
	w := http.Response{
		StatusCode: b.response.code,
	}

	if len(b.response.body) > 0 {
		w.Body = io.NopCloser(strings.NewReader(b.response.body))
	}

	return &w
}

func NewRequest(t *testing.T) *RequestBuilder {
	b := RequestBuilder{t: t}
	b.response.code = http.StatusOK
	return &b
}
