package tea

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/pghq/go-tea/trail"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestPart(t *testing.T) {
	t.Parallel()
	t.Run("raises bad request body errors", func(t *testing.T) {
		body := iotest.ErrReader(trail.NewError("an error has occurred"))
		req := httptest.NewRequest("POST", "/tests", body)
		_, err := part(httptest.NewRecorder(), req, "test")
		assert.True(t, trail.IsBadRequest(err))
	})

	t.Run("raises content type errors", func(t *testing.T) {
		body := strings.NewReader(`{
			"data": "test"
		}`)
		req := httptest.NewRequest("POST", "/tests", body)
		_, err := part(httptest.NewRecorder(), req, "test")
		assert.True(t, trail.IsBadRequest(err))
	})

	t.Run("raises not found errors", func(t *testing.T) {
		body := new(bytes.Buffer)
		mw := multipart.NewWriter(body)
		_ = mw.WriteField("foo", "test")
		_ = mw.WriteField("bar", "test")
		_ = mw.Close()

		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		_, err := part(httptest.NewRecorder(), req, "file")
		assert.NotNil(t, err)
	})

	// https://stackoverflow.com/questions/4238809/example-of-multipart-form-data
	t.Run("can get part", func(t *testing.T) {
		body := new(bytes.Buffer)
		mw := multipart.NewWriter(body)
		_ = mw.WriteField("foo", "test")
		_ = mw.WriteField("bar", "test")
		mp, _ := mw.CreateFormFile("file", "file.csv")
		_, _ = mp.Write([]byte(`example`))
		_ = mw.Close()

		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		p, err := part(httptest.NewRecorder(), req, "file")
		defer p.Close()
		assert.Nil(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, "file.csv", p.FileName())
		assert.Equal(t, "application/octet-stream", p.Header.Get("Content-Type"))
		data, _ := io.ReadAll(p)
		assert.Equal(t, "example", string(data))

		p, err = part(httptest.NewRecorder(), req, "foo")
		defer p.Close()
		assert.Nil(t, err)
		assert.NotNil(t, p)
		data, _ = io.ReadAll(p)
		assert.Equal(t, "test", string(data))
	})
}

func TestAuth(t *testing.T) {
	t.Parallel()
	t.Run("ignores no auth header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		auth := auth(req, "bearer")
		assert.Empty(t, auth)
	})

	t.Run("can detect auth header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Authorization", "Basic ZGVtbzpwQDU1dzByZA==")
		auth := auth(req, "basic")
		assert.Equal(t, "ZGVtbzpwQDU1dzByZA==", auth)
	})
}

func TestAccepts(t *testing.T) {
	t.Parallel()
	t.Run("accepts JSON if not present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		accepts := accepts(req, "application/json")
		assert.True(t, accepts)
	})

	t.Run("recognizes an exact match", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "image/webp,image/png,image/svg+xml,image/*;application/json")
		accepts := accepts(req, "application/json")
		assert.True(t, accepts)
	})

	t.Run("recognizes the wildcard", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "image/webp,image/png,image/svg+xml,image/*;q=0.8,video/*;q=0.8,*/*;q=0.5")
		accepts := accepts(req, "application/json")
		assert.True(t, accepts)
	})

	t.Run("does not match", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		req.Header.Set("Accept", "image/webp,image/png,image/svg+xml,image/*;q=0.8,video/*;q=0.8;q=0.5")
		accepts := accepts(req, "application/json")
		assert.False(t, accepts)
	})
}

func TestParse(t *testing.T) {
	t.Parallel()

	t.Run("raises nil value errors", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		err := Parse(httptest.NewRecorder(), req, nil)
		assert.True(t, trail.IsFatal(err))
	})

	t.Run("can send no content", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		err := Parse(httptest.NewRecorder(), req, &struct{}{})
		assert.Nil(t, err)
	})

	t.Run("raises bad request body errors", func(t *testing.T) {
		body := iotest.ErrReader(trail.NewError("an error has occurred"))
		req := httptest.NewRequest("POST", "/tests", body)
		err := Parse(httptest.NewRecorder(), req, &struct{}{})
		assert.True(t, trail.IsBadRequest(err))
	})

	t.Run("raises JSON errors", func(t *testing.T) {
		body := strings.NewReader(`{
			"data": "test",
		}`)
		value := struct{}{}
		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", "application/json")
		err := Parse(httptest.NewRecorder(), req, &value)
		assert.True(t, trail.IsBadRequest(err))
	})

	t.Run("raises content type error", func(t *testing.T) {
		body := strings.NewReader(`{
			"data": "test"
		}`)
		value := struct{}{}
		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", "application/bson")
		err := Parse(httptest.NewRecorder(), req, &value)
		assert.True(t, trail.IsBadRequest(err))
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

	t.Run("multipart", func(t *testing.T) {
		body := new(bytes.Buffer)
		mw := multipart.NewWriter(body)
		_ = mw.WriteField("foo", "test")
		_ = mw.WriteField("bar", "test")
		mp, _ := mw.CreateFormFile("file", "file.csv")
		_, _ = mp.Write([]byte(`example`))
		_ = mw.Close()
		req := httptest.NewRequest("POST", "/tests", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())

		t.Run("not a struct", func(t *testing.T) {
			var query int
			err := Parse(httptest.NewRecorder(), req, &query)
			assert.NotNil(t, err)
		})

		t.Run("missing part", func(t *testing.T) {
			type avatarQuery struct {
				Avatar io.Reader `form:"avatar"`
			}

			var query avatarQuery
			err := Parse(httptest.NewRecorder(), req, &query)
			assert.NotNil(t, err)
		})

		t.Run("unexported", func(t *testing.T) {
			type partQuery struct {
				file io.Reader `form:"file"`
			}

			var query partQuery
			err := Parse(httptest.NewRecorder(), req, &query)
			assert.Nil(t, err)
			assert.Nil(t, query.file)
		})

		t.Run("ok", func(t *testing.T) {
			type partQuery struct {
				File io.Reader `form:"file"`
			}

			var query partQuery
			err := Parse(httptest.NewRecorder(), req, &query)
			assert.Nil(t, err)
			assert.NotNil(t, query.File)
		})
	})

	t.Run("headers", func(t *testing.T) {
		t.Run("not a struct", func(t *testing.T) {
			var value int
			req := httptest.NewRequest("GET", "/tests", nil)
			req.Header.Set("Authorization", "Bearer foo")
			err := Parse(httptest.NewRecorder(), req, &value)
			assert.NotNil(t, err)
		})

		t.Run("auth", func(t *testing.T) {
			var value struct {
				AccessToken string `auth:"bearer"`
			}
			req := httptest.NewRequest("GET", "/tests", nil)
			req.Header.Set("Authorization", "Bearer foo")
			err := Parse(httptest.NewRecorder(), req, &value)
			assert.Nil(t, err)
			assert.Equal(t, "foo", value.AccessToken)
		})

		t.Run("embedded header value", func(t *testing.T) {
			type Embedded struct {
				Id string `header:"X-Network-Id"`
			}
			var value struct {
				Embedded
			}

			req := httptest.NewRequest("GET", "/tests", nil)
			req.Header.Set("X-Network-Id", "foo")
			err := Parse(httptest.NewRecorder(), req, &value)
			assert.Nil(t, err)
			assert.Equal(t, "foo", value.Id)
		})

		t.Run("header value string", func(t *testing.T) {
			var value struct {
				Id string `header:"X-Network-Id"`
			}
			req := httptest.NewRequest("GET", "/tests", nil)
			req.Header.Set("X-Network-Id", "foo")
			err := Parse(httptest.NewRecorder(), req, &value)
			assert.Nil(t, err)
			assert.Equal(t, "foo", value.Id)
		})

		t.Run("header value string default", func(t *testing.T) {
			var value struct {
				Id string `header:"X-Network-Id" default:"bar"`
			}
			req := httptest.NewRequest("GET", "/tests", nil)
			err := Parse(httptest.NewRecorder(), req, &value)
			assert.Nil(t, err)
			assert.Equal(t, "bar", value.Id)
		})

		t.Run("header value string slice", func(t *testing.T) {
			var value struct {
				Ids []string `header:"X-Network"`
			}
			req := httptest.NewRequest("GET", "/tests", nil)
			req.Header.Add("X-Network", "foo")
			req.Header.Add("X-Network", "bar")
			err := Parse(httptest.NewRecorder(), req, &value)
			assert.Nil(t, err)
			assert.Equal(t, []string{"foo", "bar"}, value.Ids)
		})
	})
}

func TestParseURL(t *testing.T) {
	t.Parallel()

	t.Run("can decode query", func(t *testing.T) {
		var query struct {
			QueryData string `query:"data"`
		}
		req := httptest.NewRequest("GET", "/tests?data=test", nil)
		req.Header.Set("Content-Type", "application/json")
		err := parseURL(req, &query)
		assert.Nil(t, err)
		assert.Equal(t, "test", query.QueryData)
	})

	t.Run("raises nil query errors", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/tests", nil)
		err := parseURL(req, nil)
		assert.True(t, trail.IsFatal(err))
	})

	t.Run("raises bad query errors", func(t *testing.T) {
		var query struct {
			First int `query:"first"`
		}

		req := httptest.NewRequest("GET", "/tests?first=three", nil)
		err := parseURL(req, &query)
		assert.True(t, trail.IsBadRequest(err))
	})

	t.Run("raises bad path errors", func(t *testing.T) {
		var query struct {
			Test int `path:"test"`
		}

		req := httptest.NewRequest("GET", "/tests/:test", nil)
		req = mux.SetURLVars(req, map[string]string{"test": "one"})
		err := parseURL(req, &query)
		assert.True(t, trail.IsBadRequest(err))
	})
}
