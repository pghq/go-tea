package request

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/iotest"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

func TestDecode(t *testing.T) {
	t.Run("BadQuery", func(t *testing.T) {
		var query struct {
			First int `json:"first"`
		}

		req := httptest.NewRequest("GET", "/tests?first=three", nil)
		err := Decode(httptest.NewRecorder(), req, &query)
		assert.Equal(t, http.StatusBadRequest, errors.StatusCode(err))
	})

	t.Run("BadPath", func(t *testing.T) {
		var query struct {
			Test int `json:"test"`
		}

		req := httptest.NewRequest("GET", "/tests/:test", nil)
		req = mux.SetURLVars(req, map[string]string{"test": "one"})
		err := Decode(httptest.NewRecorder(), req, &query)
		assert.Equal(t, http.StatusBadRequest, errors.StatusCode(err))
	})

	t.Run("NoBody", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		err := Decode(httptest.NewRecorder(), req, &struct{}{})
		assert.Nil(t, err)
	})

	t.Run("BodyError", func(t *testing.T) {
		body := iotest.ErrReader(errors.New("an error has occurred"))
		req := httptest.NewRequest("POST", "/tests", body)
		err := Decode(httptest.NewRecorder(), req, &struct{}{})
		assert.Equal(t, http.StatusBadRequest, errors.StatusCode(err))
	})

	t.Run("JSONError", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{
			"data": "test",
		}`))
		value := struct {}{}
		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", "application/json")
		err := Decode(httptest.NewRecorder(), req, &value)
		assert.Equal(t, http.StatusBadRequest, errors.StatusCode(err))
	})

	t.Run("ContentTypeError", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{
			"data": "test"
		}`))
		value := struct {}{}
		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", "application/bson")
		err := Decode(httptest.NewRecorder(), req, &value)
		assert.Equal(t, http.StatusBadRequest, errors.StatusCode(err))
	})

	t.Run("NoError", func(t *testing.T) {
		body := bytes.NewReader([]byte(`{
			"data": "test"
		}`))
		var value struct {Data string `json:"data"`}
		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", "application/json")
		err := Decode(httptest.NewRecorder(), req, &value)
		assert.Nil(t, err)
	})
}

func TestAuthorization(t *testing.T) {
	t.Run("NotPresent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		scheme, parameters := Authorization(req)
		assert.Empty(t, scheme)
		assert.Empty(t, parameters)
	})

	t.Run("Present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Authorization", "Basic ZGVtbzpwQDU1dzByZA==")
		scheme, parameters := Authorization(req)
		assert.Equal(t, "Basic", scheme)
		assert.Equal(t, "ZGVtbzpwQDU1dzByZA==", parameters)
	})
}

func TestFirst(t *testing.T) {
	t.Run("NotANumberError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests?first=zero", nil)
		_, err := First(req)
		assert.Equal(t, http.StatusBadRequest, errors.StatusCode(err))
	})

	t.Run("TooManyResultsError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests?first=500", nil)
		_, err := First(req)
		assert.Equal(t, http.StatusBadRequest, errors.StatusCode(err))
	})

	t.Run("NoError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests?first=5", nil)
		first, err := First(req)
		assert.Nil(t, err)
		assert.Equal(t, uint(5), first)
	})

	t.Run("Default", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		_, err := First(req)
		assert.Nil(t, err)
	})
}

func TestAfter(t *testing.T) {
	t.Run("DecodeError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests?after=museum", nil)
		_, err := After(req)
		assert.Equal(t, http.StatusBadRequest, errors.StatusCode(err))
	})

	t.Run("NoError", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests?after=bXVzZXVt", nil)
		after, err := After(req)
		assert.Nil(t, err)
		assert.Equal(t, "museum", after)
	})

	t.Run("Default", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		_, err := After(req)
		assert.Nil(t, err)
	})
}

func TestAccepts(t *testing.T) {
	t.Run("NotPresent", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		accepts := Accepts(req, "application/json")
		assert.True(t, accepts)
	})

	t.Run("ExactMatch", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "image/webp,image/png,image/svg+xml,image/*;application/json")
		accepts := Accepts(req, "application/json")
		assert.True(t, accepts)
	})

	t.Run("AnyMatch", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "image/webp,image/png,image/svg+xml,image/*;q=0.8,video/*;q=0.8,*/*;q=0.5")
		accepts := Accepts(req, "application/json")
		assert.True(t, accepts)
	})

	t.Run("NoMatch", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "image/webp,image/png,image/svg+xml,image/*;q=0.8,video/*;q=0.8;q=0.5")
		accepts := Accepts(req, "application/json")
		assert.False(t, accepts)
	})
}
