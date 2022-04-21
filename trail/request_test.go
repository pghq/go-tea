package trail

import (
	"fmt"
	"net/http/httptest"
	"testing"

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
		assert.NotNil(t, req.Response())
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
		assert.NotNil(t, req.Response())
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

		t.Run("can set request properties", func(t *testing.T) {
			req.AddFactors(uuid.New())
			req.AddDemographics(uuid.New())
			req.SetLocation(&Location{})
			req.SetStatus(200)
			req.SetUserId(uuid.New())
			req.SetPrimaryId(uuid.New())
			req.SetSecondaryId(uuid.New())

			assert.NotNil(t, req.UserId())
			assert.NotEmpty(t, req.Factors())
			assert.NotEmpty(t, req.Demographics())
			assert.NotNil(t, req.Location())
			assert.Equal(t, 200, req.Status())
			assert.NotEqual(t, uuid.Nil, req.PrimaryId())
			assert.NotEqual(t, uuid.Nil, req.SecondaryId())

			req.SetProfile("foo")

			var profile string
			assert.Nil(t, req.Profile(&profile))
			assert.Equal(t, "foo", profile)
			req.SetProfile(func() {})
		})

		t.Run("can add response", func(t *testing.T) {
			req, _ := NewRequest(w, r, "1.0.0")
			w.Header().Set("Request-Trail", "KLUv/UQAVQLdCgAy2UchQGkrAG2gmO1YTbhx7HPCGS/J3AQIx3HZZD8ceXHppasID3xdFnp0Ash82IOZU3IGT1NuKn10Zth4X1tKGcxgU5Adh3fs4k2J0+EbITlUqkgiPRy+Kk0jc/RlJTLNnQL4hjazsHTzOcShXsvtDnDrWK1Nz7FVfUTm1WRywESaWBB6Hj2PA/9elI4FsEKY3NjLba+2NrEZN7mUHJvyDnxDlyojcvQ4NcpFcXQBwwJCgQDBcPhWJMhCmaMvQxmJSY5em64Ah1fQxdHZYOM4vLIkDW2FI+YE3yxqrM3Rj/la4hbdQtw267gSmWsRqntnau+YLMauuyQOX5VODNBuFkdvNZlcMT2ZkE3t4r631rKUb+N8K0lEO/w01KYXCU3vBBUAT5dFCDSiCzqMZADDw1Blw09FwVocdChWVRkB70JJgueV2IYGRhCC/MJrM0KCEMvnAsEtDnSnGWGuPEE=")
			req.AddResponse(w)
		})
	})

}
