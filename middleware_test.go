package tea

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-tea/internal"
)

func TestMiddleware_Handle(t *testing.T) {
	t.Run("can create handler instance", func(t *testing.T) {
		m := MiddlewareFunc(func(h http.Handler) http.Handler {
			return internal.NoopHandler
		})

		assert.NotNil(t, m)
		h := m.Handle(nil)
		assert.NotNil(t, h)
	})
}

func TestNewCORSMiddleware(t *testing.T) {
	t.Run("can create wildcard instance", func(t *testing.T) {
		m := NewCORSMiddleware("*")
		assert.NotNil(t, m)
	})

	t.Run("can create instance with single origin", func(t *testing.T) {
		m := NewCORSMiddleware("https://test.domain.tld")
		assert.NotNil(t, m)
	})

	t.Run("can create instance with no origin", func(t *testing.T) {
		m := NewCORSMiddleware()
		assert.NotNil(t, m)
	})
}

func TestCORSMiddleware_Handle(t *testing.T) {
	t.Run("handles cors for request with no origin", func(t *testing.T) {
		m := NewCORSMiddleware()
		r := httptest.NewRequest("OPTIONS", "/tests", nil)

		w := httptest.NewRecorder()
		m.Handle(internal.NoopHandler).ServeHTTP(w, r)

		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
		assert.Contains(t, w.Header().Values("Vary"), "Origin")
	})

	t.Run("handles cors for request with no matching origin", func(t *testing.T) {
		m := NewCORSMiddleware("https://test.siteb.tld")
		r := httptest.NewRequest("OPTIONS", "/tests", nil)
		r.Header.Set("Origin", "https://test.site.tld")
		r.Header.Set("Access-Control-Request-Method", "GET")
		r.Header.Set("Access-Control-Request-Headers", "Content-Type")

		w := httptest.NewRecorder()
		m.Handle(internal.NoopHandler).ServeHTTP(w, r)

		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
		assert.Contains(t, w.Header().Values("Vary"), "Origin")
	})

	t.Run("handles cors for request with wildcard origin", func(t *testing.T) {
		m := NewCORSMiddleware("*")
		r := httptest.NewRequest("OPTIONS", "/tests", nil)
		r.Header.Set("Origin", "https://test.site.tld")
		r.Header.Set("Access-Control-Request-Method", "GET")
		r.Header.Set("Access-Control-Request-Headers", "Content-Type")

		w := httptest.NewRecorder()
		m.Handle(internal.NoopHandler).ServeHTTP(w, r)

		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Content-Type", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "GET", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Contains(t, w.Header().Values("Vary"), "Origin")
		assert.Contains(t, w.Header().Values("Vary"), "Access-Control-Request-Method")
		assert.Contains(t, w.Header().Values("Vary"), "Access-Control-Request-Headers")
	})

	t.Run("handles cors for request with matching origin", func(t *testing.T) {
		m := NewCORSMiddleware("https://test.site.tld")
		r := httptest.NewRequest("OPTIONS", "/tests", nil)
		r.Header.Set("Origin", "https://test.site.tld")
		r.Header.Set("Access-Control-Request-Method", "GET")
		r.Header.Set("Access-Control-Request-Headers", "Content-Type")

		w := httptest.NewRecorder()
		m.Handle(internal.NoopHandler).ServeHTTP(w, r)

		assert.Equal(t, "https://test.site.tld", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "Content-Type", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "GET", w.Header().Get("Access-Control-Allow-Methods"))
		assert.Contains(t, w.Header().Values("Vary"), "Origin")
		assert.Contains(t, w.Header().Values("Vary"), "Access-Control-Request-Method")
		assert.Contains(t, w.Header().Values("Vary"), "Access-Control-Request-Headers")
	})
}

func TestNewSentryMiddleware(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		m := NewSentryMiddleware()
		assert.NotNil(t, m)
	})
}

func TestSentryMiddleware_Handle(t *testing.T) {
	t.Run("handles panics", func(t *testing.T) {
		defer func() {
			err := recover()
			if err != nil {
				t.Fatalf("panic not expected: %+v", err)
			}
		}()

		m := NewSentryMiddleware()
		r := httptest.NewRequest("OPTIONS", "/tests", nil)
		w := httptest.NewRecorder()

		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("an error has occurred")
		})

		m.Handle(panicHandler).ServeHTTP(w, r)
	})
}
