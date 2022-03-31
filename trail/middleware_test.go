package trail

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompressHeader(t *testing.T) {
	t.Parallel()

	data := map[string][]string{"permissions": {"read:foo"}}
	t.Run("encodes", func(t *testing.T) {
		w := httptest.NewRecorder()
		assert.Nil(t, compressHeader(w, "data", &data))
	})

	t.Run("bad data", func(t *testing.T) {
		w := httptest.NewRecorder()
		assert.NotNil(t, compressHeader(w, "data", func() {}))
	})
}

func TestDecompressHeader(t *testing.T) {
	t.Parallel()

	data := map[string][]string{"permissions": {"read:foo"}}
	t.Run("encodes", func(t *testing.T) {
		w := httptest.NewRecorder()
		assert.Nil(t, compressHeader(w, "data", &data))

		var value map[string][]string
		err := decompressHeader(w, "data", &value)
		assert.Nil(t, err)
		assert.NotNil(t, value)
		assert.Equal(t, &value, &data)
	})

	t.Run("missing header", func(t *testing.T) {
		w := httptest.NewRecorder()
		assert.NotNil(t, decompressHeader(w, "data", nil))
	})

	t.Run("bad key", func(t *testing.T) {
		w := httptest.NewRecorder()
		w.Header().Set("data", fmt.Sprintf("%d", ^uint(0)))
		assert.NotNil(t, decompressHeader(w, "data", nil))
	})
}

func TestTraceMiddleware_Handle(t *testing.T) {
	t.Parallel()

	t.Run("recovers from panic", func(t *testing.T) {
		m := NewTraceMiddleware("", func(bundle []Fiber) {})
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("panic")
		})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/test", nil))
	})

	t.Run("writes fiber header", func(t *testing.T) {
		m := NewTraceMiddleware("", nil)
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/test", nil))
	})

	t.Run("collects fibers", func(t *testing.T) {
		m := NewTraceMiddleware("", func(bundle []Fiber) {})
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = compressHeader(w, "Trail-Fiber", Fiber{})
			_ = compressHeader(w, "Trail-Fiber", Fiber{})
			w.WriteHeader(http.StatusNoContent)
		})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/test", nil))
	})
}
