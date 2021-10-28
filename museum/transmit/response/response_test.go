package response

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/diagnostic/log"
)

func TestSend(t *testing.T) {
	t.Run("sends no content response", func(t *testing.T) {
		w := httptest.NewRecorder()
		Send(w, nil, nil)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("raises encode errors", func(t *testing.T) {
		log.Writer(io.Discard)
		defer log.Reset()
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		Send(w, req, New(func() {}))
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("raises MIME errors", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "application/js")
		Send(w, req, New("test"))
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("can send", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		now := time.Now()
		cursor, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.99999-05:00")
		builder := New("test").Cached(now).Cursor(&cursor)
		Send(w, req, builder)
		assert.Equal(t, http.StatusOK, w.Code)
		body := fmt.Sprintf(`{
				"data": "test",
				"cachedAt":"%s",
				"cursor":"MjAwNi0wMS0wMlQxNTowNDowNS45OTk5OS0wNTowMA=="
			}`,
			now.Format(time.RFC3339Nano),
		)
		assert.JSONEq(t, body, w.Body.String())
	})

	t.Run("can send raw", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		builder := NewRaw(http.Header{"key": []string{"value"}}, 200, []byte("ok"))
		Send(w, req, builder)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
		assert.Equal(t, http.Header{"key": []string{"value"}}, w.Header())
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
		now := time.Now()
		builder = builder.Cursor(&now)
		assert.NotNil(t, builder)
		assert.Equal(t, &now, builder.cursor)
	})
}

func TestNew(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		builder := New("test")
		assert.NotNil(t, builder)
	})
}
