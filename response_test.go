package tea

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSend(t *testing.T) {
	t.Parallel()

	t.Run("sends no content response", func(t *testing.T) {
		w := httptest.NewRecorder()
		Send(w, nil, nil)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("raises encode errors", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		Send(w, req, func() {})
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("raises MIME errors", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "application/js")
		Send(w, req, "test")
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("can send", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		Send(w, req, "test")
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "test", w.Body.String())
	})

	t.Run("can send json", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Content-Type", "application/json")
		Send(w, req, map[string]interface{}{"key": "value"})
		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{"key": "value"}`, w.Body.String())
	})
}

func TestSendError(t *testing.T) {
	t.Parallel()

	t.Run("no content", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		SendError(w, req, nil)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestBody(t *testing.T) {
	t.Parallel()

	t.Run("byte slice", func(t *testing.T) {
		_, _, err := Body(httptest.NewRequest("", "/test", nil), []byte{})
		assert.Nil(t, err)
	})
}

func TestSendNotAuthorized(t *testing.T) {
	t.Parallel()

	t.Run("not unauthorized", func(t *testing.T) {
		SendNotAuthorized(httptest.NewRecorder(), httptest.NewRequest("", "/test", nil), Err())
	})

	t.Run("unauthorized", func(t *testing.T) {
		SendNotAuthorized(httptest.NewRecorder(), httptest.NewRequest("", "/test", nil), ErrBadRequest())
	})
}
