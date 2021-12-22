package tea

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddleware_Handle(t *testing.T) {
	t.Run("can create handler instance", func(t *testing.T) {
		m := MiddlewareFunc(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		})

		assert.NotNil(t, m)
		h := m.Handle(nil)
		assert.NotNil(t, h)
	})
}

func TestCORS(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		assert.NotNil(t, CORS())
	})
}

func TestCORSMiddleware_Handle(t *testing.T) {
	t.Run("handles cors for request with no origin", func(t *testing.T) {
		m := CORS()
		r := httptest.NewRequest("OPTIONS", "/tests", nil)
		r.Header.Set("Origin", "https://tea.pghq.app")
		w := httptest.NewRecorder()
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(w, r)

		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Headers"))
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Methods"))
		assert.Contains(t, w.Header().Values("Vary"), "Origin")
	})

	t.Run("handles cors for request with no matching origin", func(t *testing.T) {
		m := CORS()
		r := httptest.NewRequest("OPTIONS", "/tests", nil)
		r.Header.Set("Origin", "https://test.site.tld")
		r.Header.Set("Access-Control-Request-Method", "GET")
		r.Header.Set("Access-Control-Request-Headers", "Content-Type")

		w := httptest.NewRecorder()
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(w, r)

		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
		assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
		assert.Contains(t, w.Header().Values("Vary"), "Origin")
	})
}
