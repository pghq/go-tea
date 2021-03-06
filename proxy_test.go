package tea

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pghq/go-tea/trail"

	"github.com/stretchr/testify/assert"
)

func TestProxy_ServeHTTP(t *testing.T) {
	t.Parallel()

	t.Run("bad host", func(t *testing.T) {
		p := NewProxy("")
		err := p.Direct("", "")
		assert.NotNil(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		p := NewProxy("")
		r := httptest.NewRequest("", "/", nil)
		w := httptest.NewRecorder()
		p.ServeHTTP(w, r)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("health check", func(t *testing.T) {
		p := NewProxy("0.0.1")
		r := httptest.NewRequest("", "/health/status", nil)
		w := httptest.NewRecorder()
		p.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("director", func(t *testing.T) {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Request-Trail", "KLUv/UQAsAC1BgCijisfYEfcABUOJnFtuxCOKKEymXXHBCOIpJLlRk03FgQBiwkJz1yWqMZKFLE0p3Rz7DTNsGvO0TGZ5rwn9t007ugDm6lgyswCw0fPXJaY1MiEFEv30DWSkiRUiaqp37PBorRThVGosxy9HzfA9wj3HD2LLNL3TLM8xwhHmbp/T7d+R2/c2By+b4YPAY7fc3RGeCUHN8UGhWolgosThexICnAxJWCGxr4vCPTj54DdEwEOAH4DAkjGqD3tNkOpKnPgBZR0eDJc0wCQrLv6UOXy1bzcrxkFkVtadA==")
		}))
		defer s.Close()
		p := NewProxy("")
		p.Middleware(MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				span := trail.StartSpan(r.Context(), "test")
				assert.NotEmpty(t, span.SpanId)
				assert.NotEmpty(t, span.ParentId)

				r.Header.Add("Test", "1")
				next.ServeHTTP(w, r)
				for _, v := range r.Header.Values("Test") {
					w.Header().Add("Test", v)
				}
			})
		}))
		p.Middleware(MiddlewareFunc(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
				r.Header.Add("Test", "2")
			})
		}))
		err := p.Direct("test", s.URL)
		assert.Nil(t, err)
		r := httptest.NewRequest("", "/test/foo", nil)
		w := httptest.NewRecorder()
		p.ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, []string{"1", "2"}, w.Header().Values("Test"))
		assert.Empty(t, w.Header().Get("Request-Trail"))
	})
}
