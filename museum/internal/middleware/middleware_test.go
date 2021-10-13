package middleware

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/internal/test"
)

func TestMiddleware_Handle(t *testing.T) {
	t.Run("NotNil", func(t *testing.T) {
		m := Middleware(func(h http.Handler) http.Handler {
			return test.NoopHandler
		})

		assert.NotNil(t, m)
		h := m.Handle(nil)
		assert.NotNil(t, h)
	})
}
