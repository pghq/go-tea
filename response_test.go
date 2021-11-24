package tea

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	t.Run("sends no content response", func(t *testing.T) {
		w := httptest.NewRecorder()
		NewResponse().Send(w, nil)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("raises encode errors", func(t *testing.T) {
		SetGlobalLogWriter(io.Discard)
		defer ResetGlobalLogger()
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		NewResponse().Body(func() {}).Send(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("raises MIME errors", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "application/js")
		NewResponse().Body("test").Send(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("can send", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		now := time.Now()
		cursor, _ := time.Parse(time.RFC3339Nano, "2006-01-02T15:04:05.99999-05:00")
		NewResponse().Body("test").Cached(now).Cursor(&cursor).Status(205).Send(w, req)
		assert.Equal(t, 205, w.Code)
		assert.Equal(t, now.Format(time.RFC3339Nano), w.Header().Get("Cached-At"))
		assert.Equal(t, "</tests?after=MjAwNi0wMS0wMlQxNTowNDowNS45OTk5OS0wNTowMA%3D%3D>", w.Header().Get("Link"))
		assert.Equal(t, "test", w.Body.String())
	})

	t.Run("can send json", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Content-Type", "application/json")
		NewResponse().Body(map[string]interface{}{"key": "value"}).Send(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"key": "value"}`, w.Body.String())
	})

	t.Run("can send raw", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		NewResponse().Body([]byte("ok")).Header(http.Header{"key": []string{"value"}}).Send(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
		assert.Equal(t, http.Header{"key": []string{"value"}}, w.Header())
	})
}

func TestBuilder_Cached(t *testing.T) {
	t.Run("can mark response as cached", func(t *testing.T) {
		builder := NewResponse()
		now := time.Now()
		builder = builder.Cached(now)
		assert.NotNil(t, builder)
		assert.Equal(t, now, builder.cachedAt)
	})
}

func TestBuilder_Cursor(t *testing.T) {
	t.Run("can set value", func(t *testing.T) {
		builder := NewResponse()
		now := time.Now()
		builder = builder.Cursor(&now)
		assert.NotNil(t, builder)
		assert.Equal(t, &now, builder.cursor)
	})
}

func TestNewResponse(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		builder := NewResponse()
		assert.NotNil(t, builder)
	})
}
