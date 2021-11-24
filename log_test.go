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
		LogLevel("debug")
		defer ResetLog()
		var buf bytes.Buffer
		LogWriter(&buf)
		logger := Debugf("%+v", errors.WithStack(errors.New("a log message")))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "debug")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")

		buf.Reset()
		LogWriter(&buf)
		Debug(errors.WithStack(errors.New("a log message")))
		assert.Contains(t, buf.String(), "debug")
		assert.Contains(t, buf.String(), "time")

		buf.Reset()
		Writef(&buf, "%+v", errors.WithStack(errors.New("a log message")))
		assert.Contains(t, buf.String(), "debug")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")

		buf.Reset()
		Write(&buf, errors.WithStack(errors.New("a log message")))
		assert.Contains(t, buf.String(), "debug")
		assert.Contains(t, buf.String(), "time")
	})
}

func TestInfo(t *testing.T) {
	t.Run("logs message", func(t *testing.T) {
		LogLevel("info")
		defer ResetLog()
		var buf bytes.Buffer
		LogWriter(&buf)
		logger := Infof("%+v", errors.WithStack(errors.New("a log message")))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "info")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")

		buf.Reset()
		LogWriter(&buf)
		Info(errors.WithStack(errors.New("a log message")))
		assert.Contains(t, buf.String(), "info")
		assert.Contains(t, buf.String(), "time")
	})
}

func TestWarn(t *testing.T) {
	t.Run("logs message", func(t *testing.T) {
		LogLevel("warn")
		defer ResetLog()
		var buf bytes.Buffer
		LogWriter(&buf)
		logger := Warnf("%+v", errors.WithStack(errors.New("a log message")))
		assert.NotNil(t, logger)
		assert.Contains(t, buf.String(), "warn")
		assert.Less(t, 1, strings.Count(buf.String(), "\\n"))
		assert.Contains(t, buf.String(), "time")

		buf.Reset()
		LogWriter(&buf)
		Warn(errors.WithStack(errors.New("a log message")))
		assert.Contains(t, buf.String(), "warn")
		assert.Contains(t, buf.String(), "time")
	})
}
