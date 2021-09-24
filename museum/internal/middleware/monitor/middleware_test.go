package monitor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		m := New()
		assert.NotNil(t, m)
	})
}

func TestMiddleware_Handle(t *testing.T) {
	t.Run("NoPanic", func(t *testing.T) {
		defer func() {
			err := recover()
			if err != nil{
				t.Fatalf("panic not expected: %+v", err)
			}
		}()

		m := New()
		r := httptest.NewRequest("OPTIONS", "/tests", nil)
		w := httptest.NewRecorder()

		panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("an error has occurred")
		})

		m.Handle(panicHandler).ServeHTTP(w, r)
	})
}
