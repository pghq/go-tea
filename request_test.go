package tea

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	t.Run("raises nil query errors", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		err := DecodeURL(req, nil)
		assert.Equal(t, http.StatusInternalServerError, StatusCode(err))
	})

	t.Run("raises nil value errors", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		err := DecodeBody(httptest.NewRecorder(), req, nil)
		assert.Equal(t, http.StatusInternalServerError, StatusCode(err))
	})

	t.Run("raises bad query errors", func(t *testing.T) {
		var query struct {
			First int `json:"first"`
		}

		req := httptest.NewRequest("GET", "/tests?first=three", nil)
		err := DecodeURL(req, &query)
		assert.Equal(t, http.StatusBadRequest, StatusCode(err))

		err = NewRequest().Query(&query).Decode(nil, req)
		assert.NotNil(t, err)
	})

	t.Run("raises bad path errors", func(t *testing.T) {
		var query struct {
			Test int `json:"test"`
		}

		req := httptest.NewRequest("GET", "/tests/:test", nil)
		req = mux.SetURLVars(req, map[string]string{"test": "one"})
		err := DecodeURL(req, &query)
		assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	})

	t.Run("can send no content", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		err := DecodeBody(httptest.NewRecorder(), req, &struct{}{})
		assert.Nil(t, err)
	})

	t.Run("raises bad request body errors", func(t *testing.T) {
		body := iotest.ErrReader(NewError("an error has occurred"))
		req := httptest.NewRequest("POST", "/tests", body)
		err := DecodeBody(httptest.NewRecorder(), req, &struct{}{})
		assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	})

	t.Run("raises JSON errors", func(t *testing.T) {
		body := strings.NewReader(`{
			"data": "test",
		}`)
		value := struct{}{}
		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", "application/json")
		err := DecodeBody(httptest.NewRecorder(), req, &value)
		assert.Equal(t, http.StatusBadRequest, StatusCode(err))

		err = NewRequest().Body(&value).Decode(httptest.NewRecorder(), req)
		assert.NotNil(t, err)
	})

	t.Run("raises content type error", func(t *testing.T) {
		body := strings.NewReader(`{
			"data": "test"
		}`)
		value := struct{}{}
		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", "application/bson")
		err := DecodeBody(httptest.NewRecorder(), req, &value)
		assert.Equal(t, http.StatusBadRequest, StatusCode(err))
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
		err := DecodeBody(httptest.NewRecorder(), req, &value)
		assert.Nil(t, err)
		assert.Equal(t, "test", value.Data)

		err = NewRequest().Body(&value).Decode(httptest.NewRecorder(), req)
		assert.Nil(t, err)
	})

	t.Run("can decode query", func(t *testing.T) {
		var query struct {
			QueryData string `json:"data"`
		}
		req := httptest.NewRequest("GET", "/tests?data=test", nil)
		req.Header.Set("Content-Type", "application/json")
		err := DecodeURL(req, &query)
		assert.Nil(t, err)
		assert.Equal(t, "test", query.QueryData)

		err = NewRequest().Query(&query).Decode(httptest.NewRecorder(), req)
		assert.Nil(t, err)
	})
}

func TestMultipartPart(t *testing.T) {
	t.Run("raises bad request body errors", func(t *testing.T) {
		body := iotest.ErrReader(NewError("an error has occurred"))
		req := httptest.NewRequest("POST", "/tests", body)
		_, err := MultipartPart(httptest.NewRecorder(), req, "test")
		assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	})

	t.Run("raises content type errors", func(t *testing.T) {
		body := strings.NewReader(`{
			"data": "test"
		}`)
		req := httptest.NewRequest("POST", "/tests", body)
		_, err := MultipartPart(httptest.NewRecorder(), req, "test")
		assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	})

	t.Run("raises not found errors", func(t *testing.T) {
		body := new(bytes.Buffer)
		mw := multipart.NewWriter(body)
		mw.WriteField("foo", "test")
		mw.WriteField("bar", "test")
		mw.Close()

		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		_, err := MultipartPart(httptest.NewRecorder(), req, "file")
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
		part, err := MultipartPart(httptest.NewRecorder(), req, "file")
		defer part.Close()
		assert.Nil(t, err)
		assert.NotNil(t, part)
		assert.Equal(t, "file.csv", part.FileName())
		assert.Equal(t, "application/octet-stream", part.Header.Get("Content-Type"))
		data, _ := io.ReadAll(part)
		assert.Equal(t, "example", string(data))

		part, err = MultipartPart(httptest.NewRecorder(), req, "foo")
		defer part.Close()
		assert.Nil(t, err)
		assert.NotNil(t, part)
		data, _ = io.ReadAll(part)
		assert.Equal(t, "test", string(data))
	})
}

func TestAuthorization(t *testing.T) {
	t.Run("ignores no auth header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		scheme, parameters := Authorization(req)
		assert.Empty(t, scheme)
		assert.Empty(t, parameters)
	})

	t.Run("can detect auth header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Authorization", "Basic ZGVtbzpwQDU1dzByZA==")
		scheme, parameters := Authorization(req)
		assert.Equal(t, "Basic", scheme)
		assert.Equal(t, "ZGVtbzpwQDU1dzByZA==", parameters)
	})
}

func TestPage(t *testing.T) {
	t.Run("raises not a number errors", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests?first=zero", nil)
		_, _, err := Page(req)
		assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	})

	t.Run("raises too many results errors", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests?first=500", nil)
		_, _, err := Page(req)
		assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	})

	t.Run("can detect first query", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests?first=5", nil)
		first, _, err := Page(req)
		assert.Nil(t, err)
		assert.Equal(t, 5, first)
	})

	t.Run("raises base64 errors", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests?after=museum", nil)
		_, _, err := Page(req)
		assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	})

	t.Run("raises time errors", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests?after=MjAwNi0wMS0wMg==", nil)
		_, _, err := Page(req)
		assert.Equal(t, http.StatusBadRequest, StatusCode(err))
	})

	t.Run("can decode paginated queries", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests?after=MjAwNi0wMS0wMlQxNTowNDowNS45OTk5OS0wNTowMA==", nil)
		_, after, err := Page(req)
		assert.Nil(t, err)
		assert.Equal(t, "2006-01-02T15:04:05.99999-05:00", after.Format(time.RFC3339Nano))
	})

	t.Run("uses default if not present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		_, _, err := Page(req)
		assert.Nil(t, err)
	})
}

func TestAccepts(t *testing.T) {
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
