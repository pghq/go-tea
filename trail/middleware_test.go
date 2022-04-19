package trail

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTraceMiddleware(t *testing.T) {
	t.Run("bad trail request header", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Request-Trail", "{{bad}}")
		w := httptest.NewRecorder()
		m := NewTraceMiddleware("1.0.0")
		m(http.Handler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {}))).ServeHTTP(w, r)
	})

	t.Run("panic", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		m := NewTraceMiddleware("1.0.0")
		m(http.Handler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) { panic("not impl") }))).ServeHTTP(w, r)
	})

	t.Run("ok", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		m := NewTraceMiddleware("1.0.0")
		m(http.Handler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) { writer.WriteHeader(200) }))).ServeHTTP(w, r)
	})
}
