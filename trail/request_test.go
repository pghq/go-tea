package trail

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

func TestNewRequest(t *testing.T) {
	t.Run("not base64 encoded", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Request-Trail", "{}")
		w := httptest.NewRecorder()

		_, err := NewRequest(w, r, "1.0.0")
		assert.NotNil(t, err)
	})

	t.Run("not zstd encoded", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Request-Trail", fmt.Sprintf("%d", ^uint(0)))
		w := httptest.NewRecorder()

		_, err := NewRequest(w, r, "1.0.0")
		assert.NotNil(t, err)
	})

	t.Run("not json encoded", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Request-Trail", "KLUv/QQAOQAAe3tiYWR9fRi0smE=")
		w := httptest.NewRecorder()
		_, err := NewRequest(w, r, "1.0.0")
		assert.NotNil(t, err)
	})

	t.Run("can continue from header", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Request-Trail", "KLUv/UQAsAC1BgCijisfYEfcABUOJnFtuxCOKKEymXXHBCOIpJLlRk03FgQBiwkJz1yWqMZKFLE0p3Rz7DTNsGvO0TGZ5rwn9t007ugDm6lgyswCw0fPXJaY1MiEFEv30DWSkiRUiaqp37PBorRThVGosxy9HzfA9wj3HD2LLNL3TLM8xwhHmbp/T7d+R2/c2By+b4YPAY7fc3RGeCUHN8UGhWolgosThexICnAxJWCGxr4vCPTj54DdEwEOAH4DAkjGqD3tNkOpKnPgBZR0eDJc0wCQrLv6UOXy1bzcrxkFkVtadA==")
		w := httptest.NewRecorder()

		req, err := NewRequest(w, r, "1.0.0")
		assert.Nil(t, err)
		assert.NotNil(t, req)

		req.Finish()
		assert.NotNil(t, req.Context())
		assert.NotNil(t, req.Origin())
		assert.NotNil(t, req.Response(true))
		assert.NotNil(t, req.Referrer())
		assert.NotNil(t, req.URL())
		assert.NotNil(t, req.Status())
		assert.NotEmpty(t, req.UserAgent())
		assert.NotEmpty(t, req.Version())
		assert.NotEqual(t, 0, req.Duration())
		assert.NotNil(t, req.IP())
		assert.NotEmpty(t, req.Operations())
		assert.NotEqual(t, uuid.Nil, req.RequestId())

		req.Recover(nil)
	})

	t.Run("can create new", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/test", nil)
		r.Header.Set("Referrer", "https://www.example.com")
		r.Header.Set("X-Forwarded-For", "1.2.3.4")
		r.Header.Set("User-Agent", "go-tea")

		w := httptest.NewRecorder()
		req, err := NewRequest(w, r, "1.0.0")
		assert.Nil(t, err)
		assert.NotNil(t, req)

		req.Finish()
		assert.NotNil(t, req.Context())
		assert.NotNil(t, req.Origin())
		assert.NotNil(t, req.Response(true))
		assert.NotNil(t, req.Referrer())
		assert.NotNil(t, req.URL())
		assert.NotNil(t, req.Status())
		assert.NotEmpty(t, req.UserAgent())
		assert.NotEmpty(t, req.Method())
		assert.NotEmpty(t, req.Version())
		assert.Less(t, time.Duration(0), req.Duration())
		assert.NotNil(t, req.IP())
		assert.NotEmpty(t, req.Operations())
		assert.NotEqual(t, uuid.Nil, req.RequestId())

		req.Recover(nil)

		t.Run("can set request properties", func(t *testing.T) {
			req.AddGroups("group")
			req.SetLocation(&Location{})
			req.SetStatus(200)
			req.SetUserId("user")

			assert.NotNil(t, req.UserId())
			assert.NotEmpty(t, req.Groups())
			assert.NotNil(t, req.Location())
			assert.Equal(t, 200, req.Status())

			req.SetProfile("foo")

			var profile string
			assert.Nil(t, req.Profile(&profile))
			assert.Equal(t, "foo", profile)
			req.SetProfile(func() {})
		})

		t.Run("can add response", func(t *testing.T) {
			req, _ := NewRequest(w, r, "1.0.0")
			w.Header().Set("Request-Trail", "KLUv/UQAyAFtCQDilD8kEGusAIj7D0lI+z3b+EhLuHpxjWCuqtJeBdjAXbzivqqqqpo02+Hj8Tgc2GL3Ns6vXWtPXe/4l9J6l7ApTJTT0XePkZ34luZ5JAKjwkxkGSEmeU98SRKT8/uqoFujay/KgG1hV8TQAJdkw0iJyKmZlNeOJN+U0fImSSLvSWYjynwcEWavIfBtbYzebZvG1/gWy2il7jbtRgxq3T0A0SAtHFyI89uqJlqxuy9QgQlW3f3avofzrZrdHRogOH/Sxu3r7hFGiJzvoJWtvXAXY8K32Y2uvXttm0V9H6XOv46mBGrb7O7PUqBYOLMxYMjHgXouFEqCSpw/Edd2ptD1Wg8Afa4xWhhLxQzDbCDLAvXClr1UGQYVwp7JHEKqeS38ukL7oikxra5NzCYUEnRnOA==")
			req.AddResponseHeaders(w.Header())
		})
	})

}
