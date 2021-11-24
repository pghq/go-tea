package tea

import (
	"bytes"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestDebug(t *testing.T) {
	t.Run("logs message", func(t *testing.T) {
		SetGlobalLogLevel("debug")
		defer ResetGlobalLogger()
		var buf bytes.Buffer
		SetGlobalLogWriter(&buf)
		logger := Logf("debug", "%+v", errors.WithStack(errors.New("a log message")))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "debug")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")

		buf.Reset()
		SetGlobalLogWriter(&buf)
		Log("debug", errors.WithStack(errors.New("a log message")))
		assert.Contains(t, buf.String(), "debug")
		assert.Contains(t, buf.String(), "time")

		buf.Reset()
		Debugf(&buf, "%+v", errors.WithStack(errors.New("a log message")))
		assert.Contains(t, buf.String(), "debug")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")

		buf.Reset()
		Debug(&buf, errors.WithStack(errors.New("a log message")))
		assert.Contains(t, buf.String(), "debug")
		assert.Contains(t, buf.String(), "time")
	})
}

func TestInfo(t *testing.T) {
	t.Run("logs message", func(t *testing.T) {
		SetGlobalLogLevel("info")
		defer ResetGlobalLogger()
		var buf bytes.Buffer
		SetGlobalLogWriter(&buf)
		logger := Logf("info", "%+v", errors.WithStack(errors.New("a log message")))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "info")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")

		buf.Reset()
		SetGlobalLogWriter(&buf)
		Log("info", errors.WithStack(errors.New("a log message")))
		assert.Contains(t, buf.String(), "info")
		assert.Contains(t, buf.String(), "time")
	})
}

func TestWarn(t *testing.T) {
	t.Run("logs message", func(t *testing.T) {
		SetGlobalLogLevel("warn")
		defer ResetGlobalLogger()
		var buf bytes.Buffer
		SetGlobalLogWriter(&buf)
		logger := Logf("warn", "%+v", errors.WithStack(errors.New("a log message")))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "warn")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")

		buf.Reset()
		SetGlobalLogWriter(&buf)
		Log("warn", errors.WithStack(errors.New("a log message")))
		assert.Contains(t, buf.String(), "warn")
		assert.Contains(t, buf.String(), "time")
	})
}
