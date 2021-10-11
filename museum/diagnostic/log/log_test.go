package log

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestDebug(t *testing.T) {
	t.Run("Level", func(t *testing.T){
		Level("debug")
		var buf bytes.Buffer
		Writer(&buf)
		logger := Debugf("%+v", errors.WithStack(errors.New("a log message")))
		log.Print(buf.String())
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "debug")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")
	})
}

func TestInfo(t *testing.T) {
	t.Run("Level", func(t *testing.T){
		Level("info")
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
	t.Run("Level", func(t *testing.T){
		Level("warn")
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
	t.Run("Level", func(t *testing.T){
		Level("error")
		var buf bytes.Buffer
		Writer(&buf)
		logger := Error(errors.New("a log message"))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "error")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")
	})
}

func TestHTTPError(t *testing.T) {
	req := httptest.NewRequest("GET", "/tests", nil)

	t.Run("Transfer", func(t *testing.T){
		Level("error")
		var buf bytes.Buffer
		Writer(&buf)
		logger := HTTPError(req, http.StatusBadRequest, errors.New("a log message"))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "error")
		assert.Contains(t, buf.String(), "/tests")
		assert.Contains(t, buf.String(), "400")
		assert.Contains(t, buf.String(), "time")
	})
}
