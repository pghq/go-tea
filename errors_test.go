package tea

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	t.Run("adds stacktrace to application errors", func(t *testing.T) {
		err := Error(NewError("an error has occurred"))
		assert.NotNil(t, err)
		assert.Less(t, 1, strings.Count(fmt.Sprintf("%+v", err), "\n"))
		assert.Contains(t, err.Error(), "an error has occurred")
	})

	t.Run("adds stacktrace to internal errors", func(t *testing.T) {
		err := Error(errors.New("an error has occurred"))
		assert.NotNil(t, err)
		assert.Less(t, 1, strings.Count(fmt.Sprintf("%+v", err), "\n"))
		assert.Contains(t, err.Error(), "an error has occurred")
	})

	t.Run("logs message", func(t *testing.T) {
		SetGlobalLogLevel("error")
		defer ResetGlobalLogger()
		var buf bytes.Buffer
		SetGlobalLogWriter(&buf)
		logger := CurrentLogger().Error(errors.New("a log message"))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "error")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")
	})
}

func TestHTTPError(t *testing.T) {
	req := httptest.NewRequest("GET", "/tests", nil)

	t.Run("logs message", func(t *testing.T) {
		SetGlobalLogLevel("error")
		defer ResetGlobalLogger()
		var buf bytes.Buffer
		SetGlobalLogWriter(&buf)
		logger := CurrentLogger().HTTPError(req, http.StatusBadRequest, errors.New("a log message"))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "error")
		assert.Contains(t, buf.String(), "/tests")
		assert.Contains(t, buf.String(), "400")
		assert.Contains(t, buf.String(), "time")
	})
}

func TestNew(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		err := NewError("an error has occurred")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "an error has occurred")
	})
}

func TestNewf(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		err := NewErrorf("an %s has occurred", "error")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "an error has occurred")
	})
}

func TestHTTP(t *testing.T) {
	t.Run("can cast", func(t *testing.T) {
		err := HTTPError(http.StatusBadRequest, errors.New("an error has occurred"))
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "an error has occurred")
	})

	t.Run("can create new", func(t *testing.T) {
		err := NewHTTPError(http.StatusConflict, "an error has occurred")
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusConflict, ae.code)
		assert.Contains(t, err.Error(), "an error has occurred")
	})
}

func TestNotFound(t *testing.T) {
	t.Run("from error", func(t *testing.T) {
		err := NotFound(NewError("error"))
		assert.True(t, IsNotFound(err))
	})

	t.Run("from values", func(t *testing.T) {
		err := NewNotFound("error")
		assert.True(t, IsNotFound(err))
	})
}

func TestBadRequest(t *testing.T) {
	t.Run("can cast", func(t *testing.T) {
		err := BadRequest(errors.New("an error has occurred"))
		assert.True(t, IsBadRequest(err))
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, ae.code)
		assert.Contains(t, err.Error(), "an error has occurred")
	})
}

func TestNewBadRequest(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		err := NewBadRequest("an error has occurred")
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, ae.code)
		assert.Contains(t, err.Error(), "an error has occurred")
	})
}

func TestNoContent(t *testing.T) {
	t.Run("can cast", func(t *testing.T) {
		err := NoContent(errors.New("an error has occurred"))
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNoContent, ae.code)
		assert.Contains(t, err.Error(), "an error has occurred")
	})
}

func TestNewNoContent(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		err := NewNoContent("an error has occurred")
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNoContent, ae.code)
		assert.Contains(t, err.Error(), "an error has occurred")
	})
}

func TestAsError(t *testing.T) {
	t.Run("can detect error type", func(t *testing.T) {
		err := NewError("an error has occurred")
		as := AsError(&applicationError{}, &err)
		assert.True(t, as)
	})
}

func TestIsFatal(t *testing.T) {
	t.Run("can detect fatal errors", func(t *testing.T) {
		err := NewError("an error has occurred")
		assert.True(t, IsFatal(err))
	})

	t.Run("can detect non fatal errors", func(t *testing.T) {
		err := NewHTTPError(http.StatusNoContent, "an error has occurred")
		assert.False(t, IsFatal(err))
		assert.False(t, IsFatal(nil))
	})
}

func TestStatusCode(t *testing.T) {
	t.Run("detects status code for no content errors", func(t *testing.T) {
		err := NewHTTPError(http.StatusNoContent, "an error has occurred")
		assert.Equal(t, http.StatusNoContent, StatusCode(err))
	})

	t.Run("detects status code for cancelled context errors", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		assert.Equal(t, http.StatusBadRequest, StatusCode(ctx.Err()))
		assert.Equal(t, http.StatusBadRequest, StatusCode(Error(ctx.Err())))
	})

	t.Run("detects status code for deadline exceeded errors", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()

		time.Sleep(time.Microsecond)
		assert.Equal(t, http.StatusRequestTimeout, StatusCode(ctx.Err()))
		assert.Equal(t, http.StatusRequestTimeout, StatusCode(Error(ctx.Err())))
	})

	t.Run("detects status code for internal errors", func(t *testing.T) {
		assert.Equal(t, http.StatusInternalServerError, StatusCode(errors.New("an error has occurred")))
	})
}

func TestSendError(t *testing.T) {
	t.Run("emits fatal errors", func(t *testing.T) {
		var buf bytes.Buffer
		SetGlobalLogWriter(&buf)
		defer ResetGlobalLogger()
		err := NewError("an error has occurred")
		SendError(err)
		assert.Contains(t, buf.String(), "an error has occurred")
	})

	t.Run("emits non fatal errors", func(t *testing.T) {
		var buf bytes.Buffer
		SetGlobalLogWriter(&buf)
		defer ResetGlobalLogger()
		err := NewHTTPError(http.StatusNoContent, "an error has occurred")
		SendError(err)
		assert.Empty(t, buf.String())
	})
}

func TestSendHTTP(t *testing.T) {
	req := httptest.NewRequest("GET", "/tests", nil)
	t.Run("does not emit fatal errors to client", func(t *testing.T) {
		var buf bytes.Buffer
		SetGlobalLogWriter(&buf)
		defer ResetGlobalLogger()
		w := httptest.NewRecorder()
		err := NewError("an error has occurred")
		SendHTTP(w, req, err)
		assert.Equal(t, 500, w.Code)
		assert.Contains(t, buf.String(), "an error has occurred")
	})

	t.Run("emits non fatal errors to client", func(t *testing.T) {
		var buf bytes.Buffer
		SetGlobalLogWriter(&buf)
		defer ResetGlobalLogger()
		w := httptest.NewRecorder()
		err := NewHTTPError(http.StatusNoContent, "an error has occurred")
		SendHTTP(w, req, err)
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, buf.String())
	})
}

func TestSendNotAuthorized(t *testing.T) {
	req := httptest.NewRequest("GET", "/tests", nil)
	t.Run("does not emit fatal errors to client", func(t *testing.T) {
		var buf bytes.Buffer
		SetGlobalLogWriter(&buf)
		defer ResetGlobalLogger()
		w := httptest.NewRecorder()
		err := NewError("an error has occurred")
		SendNotAuthorized(w, req, err)
		assert.Equal(t, 500, w.Code)
		assert.Contains(t, buf.String(), "an error has occurred")
	})

	t.Run("emits non fatal errors to client", func(t *testing.T) {
		var buf bytes.Buffer
		SetGlobalLogWriter(&buf)
		defer ResetGlobalLogger()
		w := httptest.NewRecorder()
		SendNewNotAuthorized(w, req, "an error has occurred")
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Empty(t, buf.String())
	})
}

func TestRecover(t *testing.T) {
	t.Run("monitors panics", func(t *testing.T) {
		SetGlobalLogWriter(io.Discard)
		defer ResetGlobalLogger()
		defer func() { Recover(recover()) }()
		panic("an error has occurred")
	})
}

func TestInit(t *testing.T) {
	t.Run("raises config errors", func(t *testing.T) {
		conf := MonitorConfig{
			Dsn: "https://localhost",
		}

		err := Init(conf)
		assert.NotNil(t, err)
	})

	t.Run("can initialize", func(t *testing.T) {
		conf := MonitorConfig{
			Version:      "0.0.1",
			Environment:  "test",
			Dsn:          "https://12345@678910.ingest.sentry.io/1",
			FlushTimeout: time.Second,
		}

		err := Init(conf)
		assert.Nil(t, err)
	})
}

func TestFatal(t *testing.T) {
	t.Run("exits", func(t *testing.T) {
		SetGlobalExitFunc(func(code int) {})
		defer ResetGlobalExitFunc()

		Fatal(NewError("error"))
	})
}
