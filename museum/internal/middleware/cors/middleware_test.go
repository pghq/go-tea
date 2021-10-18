package cors

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/internal"
)

func TestNew(t *testing.T) {
	t.Run("can create wildcard instance", func(t *testing.T) {
		m := New("*")
		assert.NotNil(t, m)
	})

	t.Run("can create instance with single origin", func(t *testing.T) {
		m := New("https://test.domain.tld")
		assert.NotNil(t, m)
	})

	t.Run("can create instance with no origin", func(t *testing.T) {
		m := New()
		assert.NotNil(t, m)
	})
}

func TestMiddleware_Handle(t *testing.T) {
	t.Run("handles cors for request with no origin", func(t *testing.T) {
		m := New()
		r := httptest.NewRequest("OPTIONS", "/tests", nil)

		w := httptest.NewRecorder()
		m.Handle(internal.NoopHandler).ServeHTTP(w, r)

		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
		assert.Contains(t, w.Header().Values("Vary"), "Origin")
	})

	t.Run("handles cors for request with no matching origin", func(t *testing.T) {
		m := New("https://test.siteb.tld")
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
		m := New("*")
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
		m := New("https://test.site.tld")
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
