package response

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	t.Run("sends no content response", func(t *testing.T) {
		res := httptest.NewRecorder()
		Send(res, nil, nil)
		assert.Equal(t, http.StatusNoContent, res.Code)
	})

	t.Run("raises encode errors", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		Send(res, req, New(func() {}))
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})

	t.Run("raises MIME errors", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "application/js")
		Send(res, req, New("test"))
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	t.Run("can send", func(t *testing.T) {
		res := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		now := time.Now()
		builder := New("test").Cached(now).Cursor("test")
		Send(res, req, builder)
		assert.Equal(t, http.StatusOK, res.Code)
		body := fmt.Sprintf(`{
				"data": "test",
				"cachedAt":"%s",
				"cursor":"dGVzdA=="
			}`,
			now.Format(time.RFC3339Nano),
		)
		assert.JSONEq(t, body, res.Body.String())
	})
}

func TestBuilder_Cached(t *testing.T) {
	t.Run("can mark response as cached", func(t *testing.T) {
		builder := New("test")
		now := time.Now()
		builder = builder.Cached(now)
		assert.NotNil(t, builder)
		assert.Equal(t, now, builder.cachedAt)
	})
}

func TestBuilder_Cursor(t *testing.T) {
	t.Run("can set value", func(t *testing.T) {
		builder := New("test")
		builder = builder.Cursor("Astro")
		assert.NotNil(t, builder)
		assert.Equal(t, "Astro", builder.cursor)
	})
}

func TestNew(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		builder := New("test")
		assert.NotNil(t, builder)
	})
}
