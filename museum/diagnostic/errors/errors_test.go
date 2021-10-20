package errors

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

	"github.com/pghq/go-eque/eque"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/diagnostic/log"
)

func TestWrap(t *testing.T) {
	t.Run("adds stacktrace to application errors", func(t *testing.T) {
		err := Wrap(New("an error has occurred"))
		assert.NotNil(t, err)
		assert.Less(t, 1, strings.Count(fmt.Sprintf("%+v", err), "\n"))
	})

	t.Run("adds stacktrace to internal errors", func(t *testing.T) {
		err := Wrap(errors.New("an error has occurred"))
		assert.NotNil(t, err)
		assert.Less(t, 1, strings.Count(fmt.Sprintf("%+v", err), "\n"))
	})
}

func TestNew(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		err := New("an error has occurred")
		assert.NotNil(t, err)
	})
}

func TestHTTP(t *testing.T) {
	t.Run("can cast", func(t *testing.T) {
		err := HTTP(errors.New("an error has occurred"), http.StatusBadRequest)
		assert.NotNil(t, err)
	})
}

func TestNewHTTP(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		err := NewHTTP("an error has occurred", http.StatusConflict)
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusConflict, ae.code)
	})
}

func TestBadRequest(t *testing.T) {
	t.Run("can cast", func(t *testing.T) {
		err := BadRequest(errors.New("an error has occurred"))
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, ae.code)
	})
}

func TestNewBadRequest(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		err := NewBadRequest("an error has occurred")
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, ae.code)
	})
}

func TestNoContent(t *testing.T) {
	t.Run("can cast", func(t *testing.T) {
		err := NoContent(errors.New("an error has occurred"))
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNoContent, ae.code)
	})
}

func TestNewNoContent(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		err := NewNoContent("an error has occurred")
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNoContent, ae.code)
	})
}

func TestAs(t *testing.T) {
	t.Run("can detect error type", func(t *testing.T) {
		err := New("an error has occurred")
		as := As(&applicationError{}, &err)
		assert.True(t, as)
	})
}

func TestIsFatal(t *testing.T) {
	t.Run("can detect fatal errors", func(t *testing.T) {
		err := New("an error has occurred")
		assert.True(t, IsFatal(err))
	})

	t.Run("can detect non fatal errors", func(t *testing.T) {
		err := NewHTTP("an error has occurred", http.StatusNoContent)
		assert.False(t, IsFatal(err))
	})
}

func TestStatusCode(t *testing.T) {
	t.Run("detects status code for no content errors", func(t *testing.T) {
		err := NewHTTP("an error has occurred", http.StatusNoContent)
		assert.Equal(t, http.StatusNoContent, StatusCode(err))
	})

	t.Run("detects status code for cancelled context errors", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		assert.Equal(t, http.StatusBadRequest, StatusCode(ctx.Err()))
	})

	t.Run("detects status code for deadline exceeded errors", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		time.Sleep(time.Microsecond)
		assert.Equal(t, http.StatusRequestTimeout, StatusCode(ctx.Err()))
	})

	t.Run("detects status code for no eque messages errors", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, StatusCode(eque.ErrNoMessages))
	})

	t.Run("detects status code for eque lock acquire errors", func(t *testing.T) {
		assert.Equal(t, http.StatusBadRequest, StatusCode(eque.ErrAcquireLockFailed))
	})

	t.Run("detects status code for internal errors", func(t *testing.T) {
		assert.Equal(t, http.StatusInternalServerError, StatusCode(errors.New("an error has occurred")))
	})
}

func TestEmit(t *testing.T) {
	t.Run("emits fatal errors", func(t *testing.T) {
		var buf bytes.Buffer
		log.Writer(&buf)
		defer log.Reset()
		err := New("an error has occurred")
		Emit(err)
		assert.Contains(t, buf.String(), "an error has occurred")
	})

	t.Run("emits non fatal errors", func(t *testing.T) {
		var buf bytes.Buffer
		log.Writer(&buf)
		defer log.Reset()
		err := NewHTTP("an error has occurred", http.StatusNoContent)
		Emit(err)
		assert.Empty(t, buf.String())
	})
}

func TestEmitHTTP(t *testing.T) {
	req := httptest.NewRequest("GET", "/tests", nil)

	t.Run("does not emit fatal errors to client", func(t *testing.T) {
		var buf bytes.Buffer
		log.Writer(&buf)
		defer log.Reset()
		w := httptest.NewRecorder()
		err := New("an error has occurred")
		EmitHTTP(w, req, err)
		assert.Equal(t, 500, w.Code)
		assert.Contains(t, buf.String(), "an error has occurred")
	})

	t.Run("emits non fatal errors to client", func(t *testing.T) {
		var buf bytes.Buffer
		log.Writer(&buf)
		defer log.Reset()
		w := httptest.NewRecorder()
		err := NewHTTP("an error has occurred", http.StatusNoContent)
		EmitHTTP(w, req, err)
		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, buf.String())
	})
}

func TestRecover(t *testing.T) {
	t.Run("monitors panics", func(t *testing.T) {
		log.Writer(io.Discard)
		defer log.Reset()
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
