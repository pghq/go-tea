package middleware

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/pilot"
)

func TestMiddleware_Handle(t *testing.T) {
	t.Run("can create handler instance", func(t *testing.T) {
		m := Func(func(h http.Handler) http.Handler {
			return pilot.NoopHandler
		})

		assert.NotNil(t, m)
		h := m.Handle(nil)
		assert.NotNil(t, h)
	})
}
