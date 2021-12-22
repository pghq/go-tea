package tea

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestPart(t *testing.T) {
	t.Parallel()
	t.Run("raises bad request body errors", func(t *testing.T) {
		body := iotest.ErrReader(Err("an error has occurred"))
		req := httptest.NewRequest("POST", "/tests", body)
		_, err := Part(httptest.NewRecorder(), req, "test")
		assert.Equal(t, http.StatusBadRequest, ErrStatus(err))
	})

	t.Run("raises content type errors", func(t *testing.T) {
		body := strings.NewReader(`{
			"data": "test"
		}`)
		req := httptest.NewRequest("POST", "/tests", body)
		_, err := Part(httptest.NewRecorder(), req, "test")
		assert.Equal(t, http.StatusBadRequest, ErrStatus(err))
	})

	t.Run("raises not found errors", func(t *testing.T) {
		body := new(bytes.Buffer)
		mw := multipart.NewWriter(body)
		mw.WriteField("foo", "test")
		mw.WriteField("bar", "test")
		mw.Close()

		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		_, err := Part(httptest.NewRecorder(), req, "file")
		assert.NotNil(t, err)
	})

	// https://stackoverflow.com/questions/4238809/example-of-multipart-form-data
	t.Run("can get part", func(t *testing.T) {
		body := new(bytes.Buffer)
		mw := multipart.NewWriter(body)
		mw.WriteField("foo", "test")
		mw.WriteField("bar", "test")
		mp, _ := mw.CreateFormFile("file", "file.csv")
		mp.Write([]byte(`example`))
		mw.Close()

		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		part, err := Part(httptest.NewRecorder(), req, "file")
		defer part.Close()
		assert.Nil(t, err)
		assert.NotNil(t, part)
		assert.Equal(t, "file.csv", part.FileName())
		assert.Equal(t, "application/octet-stream", part.Header.Get("Content-Type"))
		data, _ := io.ReadAll(part)
		assert.Equal(t, "example", string(data))

		part, err = Part(httptest.NewRecorder(), req, "foo")
		defer part.Close()
		assert.Nil(t, err)
		assert.NotNil(t, part)
		data, _ = io.ReadAll(part)
		assert.Equal(t, "test", string(data))
	})
}

func TestAuth(t *testing.T) {
	t.Parallel()
	t.Run("ignores no auth header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		auth := Auth(req)
		assert.Empty(t, auth)
	})

	t.Run("can detect auth header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Authorization", "Basic ZGVtbzpwQDU1dzByZA==")
		auth := Auth(req)
		assert.Equal(t, "ZGVtbzpwQDU1dzByZA==", auth)
	})
}

func TestAccepts(t *testing.T) {
	t.Parallel()
	t.Run("accepts JSON if not present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		accepts := Accepts(req, "application/json")
		assert.True(t, accepts)
	})

	t.Run("recognizes an exact match", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "image/webp,image/png,image/svg+xml,image/*;application/json")
		accepts := Accepts(req, "application/json")
		assert.True(t, accepts)
	})

	t.Run("recognizes the wildcard", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "image/webp,image/png,image/svg+xml,image/*;q=0.8,video/*;q=0.8,*/*;q=0.5")
		accepts := Accepts(req, "application/json")
		assert.True(t, accepts)
	})

	t.Run("does not match", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "image/webp,image/png,image/svg+xml,image/*;q=0.8,video/*;q=0.8;q=0.5")
		accepts := Accepts(req, "application/json")
		assert.False(t, accepts)
	})
}

func TestAttach(t *testing.T) {
	t.Parallel()

	data := map[string][]string{"permissions": {"read:foo"}}
	t.Run("encodes", func(t *testing.T) {
		r := httptest.NewRequest("", "/test", nil)
		assert.Nil(t, Attach(r, "data", &data))
	})

	t.Run("bad data", func(t *testing.T) {
		r := httptest.NewRequest("", "/test", nil)
		assert.NotNil(t, Attach(r, "data", func() {}))
	})
}

func TestDetach(t *testing.T) {
	t.Parallel()

	data := map[string][]string{"permissions": {"read:foo"}}
	t.Run("encodes", func(t *testing.T) {
		r := httptest.NewRequest("", "/test", nil)
		assert.Nil(t, Attach(r, "data", &data))

		var value map[string][]string
		err := Detach(r, "data", &value)
		assert.Nil(t, err)
		assert.NotNil(t, value)
		assert.Equal(t, &value, &data)
	})

	t.Run("missing header", func(t *testing.T) {
		r := httptest.NewRequest("", "/test", nil)
		assert.NotNil(t, Detach(r, "data", nil))
	})

	t.Run("bad key", func(t *testing.T) {
		r := httptest.NewRequest("", "/test", nil)
		r.Header.Set("data", fmt.Sprintf("%d", ^uint(0)))
		assert.NotNil(t, Detach(r, "data", nil))
	})
}

func TestParse(t *testing.T) {
	t.Parallel()

	t.Run("raises nil value errors", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		err := Parse(httptest.NewRecorder(), req, nil)
		assert.Equal(t, http.StatusInternalServerError, ErrStatus(err))
	})

	t.Run("can send no content", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		err := Parse(httptest.NewRecorder(), req, &struct{}{})
		assert.Nil(t, err)
	})

	t.Run("raises bad request body errors", func(t *testing.T) {
		body := iotest.ErrReader(Err("an error has occurred"))
		req := httptest.NewRequest("POST", "/tests", body)
		err := Parse(httptest.NewRecorder(), req, &struct{}{})
		assert.Equal(t, http.StatusBadRequest, ErrStatus(err))
	})

	t.Run("raises JSON errors", func(t *testing.T) {
		body := strings.NewReader(`{
			"data": "test",
		}`)
		value := struct{}{}
		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", "application/json")
		err := Parse(httptest.NewRecorder(), req, &value)
		assert.Equal(t, http.StatusBadRequest, ErrStatus(err))
	})

	t.Run("raises content type error", func(t *testing.T) {
		body := strings.NewReader(`{
			"data": "test"
		}`)
		value := struct{}{}
		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", "application/bson")
		err := Parse(httptest.NewRecorder(), req, &value)
		assert.Equal(t, http.StatusBadRequest, ErrStatus(err))
	})

	t.Run("can decode body", func(t *testing.T) {
		body := strings.NewReader(`{
			"data": "test"
		}`)
		var value struct {
			Data string `json:"data"`
		}
		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", "application/json")
		err := Parse(httptest.NewRecorder(), req, &value)
		assert.Nil(t, err)
		assert.Equal(t, "test", value.Data)
	})
}

func TestParseURL(t *testing.T) {
	t.Parallel()

	t.Run("can decode query", func(t *testing.T) {
		var query struct {
			QueryData string `json:"data"`
		}
		req := httptest.NewRequest("GET", "/tests?data=test", nil)
		req.Header.Set("Content-Type", "application/json")
		err := ParseURL(req, &query)
		assert.Nil(t, err)
		assert.Equal(t, "test", query.QueryData)
	})

	t.Run("raises nil query errors", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		err := ParseURL(req, nil)
		assert.Equal(t, http.StatusInternalServerError, ErrStatus(err))
	})

	t.Run("raises bad query errors", func(t *testing.T) {
		var query struct {
			First int `json:"first"`
		}

		req := httptest.NewRequest("GET", "/tests?first=three", nil)
		err := ParseURL(req, &query)
		assert.Equal(t, http.StatusBadRequest, ErrStatus(err))
	})

	t.Run("raises bad path errors", func(t *testing.T) {
		var query struct {
			Test int `json:"test"`
		}

		req := httptest.NewRequest("GET", "/tests/:test", nil)
		req = mux.SetURLVars(req, map[string]string{"test": "one"})
		err := ParseURL(req, &query)
		assert.Equal(t, http.StatusBadRequest, ErrStatus(err))
	})
}
