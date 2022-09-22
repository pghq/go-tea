package tea

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

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

	t.Run("can encode headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Content-Type", "application/json")

		type Nested struct {
			UUID uuid.UUID `header:"uuid"`
		}

		type headerResponse struct {
			Nested
			RequestId string   `header:"request-id"`
			Names     []string `header:"names"`
			Empty     string   `header:"empty,omitempty"`
		}

		response := headerResponse{
			Nested: Nested{
				UUID: uuid.New(),
			},
			RequestId: "foo",
			Names:     []string{"bar"},
		}

		Send(w, req, response)
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "foo", w.Header().Get("request-id"))
		assert.Equal(t, []string{"bar"}, w.Header().Values("names"))
		assert.Empty(t, w.Header().Get("empty"))
		assert.Equal(t, response.UUID.String(), w.Header().Get("uuid"))
	})
}

func TestBody(t *testing.T) {
	t.Parallel()

	t.Run("byte slice", func(t *testing.T) {
		_, _, err := body(httptest.NewRequest("", "/test", nil), []byte{})
		assert.Nil(t, err)
	})
}
