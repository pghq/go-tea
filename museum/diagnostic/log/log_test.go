package log

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestDebug(t *testing.T) {
	t.Run("logs message", func(t *testing.T) {
		Level("debug")
		defer Reset()
		var buf bytes.Buffer
		Writer(&buf)
		logger := Debugf("%+v", errors.WithStack(errors.New("a log message")))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "debug")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")
	})
}

func TestInfo(t *testing.T) {
	t.Run("logs message", func(t *testing.T) {
		Level("info")
		defer Reset()
		var buf bytes.Buffer
		Writer(&buf)
		logger := Infof("%+v", errors.WithStack(errors.New("a log message")))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "info")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")
	})
}

func TestWarn(t *testing.T) {
	t.Run("logs message", func(t *testing.T) {
		Level("warn")
		defer Reset()
		var buf bytes.Buffer
		Writer(&buf)
		logger := Warnf("%+v", errors.WithStack(errors.New("a log message")))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "warn")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")
	})
}

func TestError(t *testing.T) {
	t.Run("logs message", func(t *testing.T) {
		Level("error")
		defer Reset()
		var buf bytes.Buffer
		Writer(&buf)
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
		Level("error")
		defer Reset()
		var buf bytes.Buffer
		Writer(&buf)
		logger := CurrentLogger().HTTPError(req, http.StatusBadRequest, errors.New("a log message"))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "error")
		assert.Contains(t, buf.String(), "/tests")
		assert.Contains(t, buf.String(), "400")
		assert.Contains(t, buf.String(), "time")
	})
}
