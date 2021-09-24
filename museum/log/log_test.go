package log

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	t.Run("NoError", func(t *testing.T){
		Init()
		assert.Equal(t, zerolog.GlobalLevel(), DefaultLogLevel)
	})
}

func TestDebug(t *testing.T) {
	t.Run("Level", func(t *testing.T){
		Level("debug")
		var buf bytes.Buffer
		Writer(&buf)
		logger := Debug(fmt.Sprintf("%+v", errors.WithStack(errors.New("a log message"))))
		assert.NotNil(t, logger)
		assert.True(t, strings.Contains(buf.String(), "debug"))
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
	})
}

func TestInfo(t *testing.T) {
	t.Run("Level", func(t *testing.T){
		Level("info")
		var buf bytes.Buffer
		Writer(&buf)
		logger := Info(fmt.Sprintf("%+v", errors.WithStack(errors.New("a log message"))))
		assert.NotNil(t, logger)
		assert.True(t, strings.Contains(buf.String(), "info"))
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
	})
}

func TestWarn(t *testing.T) {
	t.Run("Level", func(t *testing.T){
		Level("warn")
		var buf bytes.Buffer
		Writer(&buf)
		logger := Warn(fmt.Sprintf("%+v", errors.WithStack(errors.New("a log message"))))
		assert.NotNil(t, logger)
		assert.True(t, strings.Contains(buf.String(), "warn"))
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
	})
}

func TestError(t *testing.T) {
	t.Run("Level", func(t *testing.T){
		Level("error")
		var buf bytes.Buffer
		Writer(&buf)
		logger := Error(errors.New("a log message"))
		assert.NotNil(t, logger)
		assert.True(t, strings.Contains(buf.String(), "error"))
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
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
		assert.True(t, strings.Contains(buf.String(), "error"))
		assert.True(t, strings.Contains(buf.String(), "/tests"))
		assert.True(t, strings.Contains(buf.String(), "400"))
	})
}
