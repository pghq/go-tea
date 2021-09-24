package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/internal/test"
)

func TestNewRouter(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		r := NewRouter(0)
		assert.NotNil(t, r)
	})
}

func TestRouter_Get(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		r := NewRouter(0)
		req := test.NewRequest(t).
			Method("GET").
			Path("/v0/tests").
			ExpectRoute("/tests").
			ExpectResponse("ok")

		RequestTest(t, r, req)
	})
}

func TestRouter_Post(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		r := NewRouter(0)
		req := test.NewRequest(t).
			Method("POST").
			Path("/v0/tests").
			Body("test").
			ExpectRoute("/tests").
			ExpectResponse("ok")

		RequestTest(t, r, req)
	})
}

func TestRouter_Put(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		r := NewRouter(0)
		req := test.NewRequest(t).
			Method("PUT").
			Path("/v0/tests/test").
			Body("test").
			ExpectRoute("/tests/test").
			ExpectResponse("ok")

		RequestTest(t, r, req)
	})
}

func TestRouter_Delete(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		r := NewRouter(0)
		req := test.NewRequest(t).
			Method("DELETE").
			Path("/v0/tests/test").
			ExpectRoute( "/tests/test").
			ExpectResponse("ok")

		RequestTest(t, r, req)
	})
}

func TestRouter_At(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		r := NewRouter(0).
			At("/tests")
		req := test.NewRequest(t).
			Method("GET").
			Path("/v0/tests/test").
			ExpectRoute("/test").
			ExpectResponse("ok")

		RequestTest(t, r, req)
	})
}

func TestRouter_Middleware(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		r := NewRouter(0).
			Middleware(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})
			})
		req := test.NewRequest(t).
			Method("GET").
			Path("/v0/tests").
			ExpectRoute("/tests").
			ExpectStatus(http.StatusNoContent)

		RequestTest(t, r, req)
	})
}

func TestNotFoundHandler(t *testing.T) {
	t.Run("Status", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/tests", nil)
		w := httptest.NewRecorder()

		NotFoundHandler(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, http.StatusText(http.StatusNotFound), w.Body.String())
	})

	t.Run("NotFoundError", func(t *testing.T) {
		r := NewRouter(0)
		req := test.NewRequest(t).
			Method("GET").
			Path("/v0/tests/test").
			ExpectRoute("/tests").
			ExpectStatus(http.StatusNotFound).
			ExpectResponse(http.StatusText(http.StatusNotFound))

		RequestTest(t, r, req)
	})
}

func TestMethodNotAllowedHandler(t *testing.T) {
	t.Run("Status", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/tests", nil)
		w := httptest.NewRecorder()

		MethodNotAllowedHandler(w, r)
		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		assert.Equal(t, http.StatusText(http.StatusMethodNotAllowed), w.Body.String())
	})

	t.Run("MethodNotAllowedError", func(t *testing.T) {
		r := NewRouter(0)
		req := test.NewRequest(t).
			Method("GET").
			Path("/v0/tests").
			ExpectRoute("/tests").
			ExpectMethod("POST").
			ExpectStatus(http.StatusMethodNotAllowed).
			ExpectResponse(http.StatusText(http.StatusMethodNotAllowed))

		RequestTest(t, r, req)
	})
}

func RequestTest(t *testing.T, r *Router, b *test.RequestBuilder){
	t.Helper()
	want := b.Response()

	var expected []byte
	if want.Body != nil{
		bytes, err := io.ReadAll(want.Body)
		assert.Nil(t, err)
		expected = bytes
		err = want.Body.Close()
		assert.Nil(t, err)
	}

	handlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(want.StatusCode)

		if len(expected) > 0{
			_, err := w.Write(expected)
			assert.Nil(t, err)
		}
	}

	switch strings.ToUpper(b.ExpectedMethod()) {
	case "GET":
		r = r.Get(b.ExpectedRoute(), handlerFunc)
	case "PUT":
		r = r.Put(b.ExpectedRoute(), handlerFunc)
	case "POST":
		r = r.Post(b.ExpectedRoute(), handlerFunc)
	case "DELETE":
		r = r.Delete(b.ExpectedRoute(), handlerFunc)
	default:
		t.Fatalf("unhandled method %s", b.ExpectedMethod())
	}
	assert.NotNil(t, r)

	s := httptest.NewServer(r)
	defer s.Close()
	req := b.Request(s.URL)
	got, err := s.Client().Do(req)
	assert.Nil(t, err)

	body, err := io.ReadAll(got.Body)
	assert.Nil(t, err)
	err = got.Body.Close()
	assert.Nil(t, err)

	assert.Equal(t, want.StatusCode, got.StatusCode)
	assert.Equal(t, len(expected), len(body))
}
