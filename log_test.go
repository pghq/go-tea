package tea

import (
	"context"
	"testing"
)

func init() {
	defer SetVerbosity("trace")
	SetVerbosity("debug")
	SetVerbosity("info")
	SetVerbosity("warn")
	SetVerbosity("error")
	SetVerbosity("fatal")
	SetVerbosity("unknown")
}

func TestLogf(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		Logf(context.TODO(), "info", "ok")
	})
}

func TestLog(t *testing.T) {
	t.Parallel()

	t.Run("debug", func(t *testing.T) {
		Log(context.TODO(), "debug")
	})

	t.Run("info", func(t *testing.T) {
		Log(context.TODO(), "info")
	})

	t.Run("warn", func(t *testing.T) {
		Log(context.TODO(), "warn")
	})

	t.Run("test", func(t *testing.T) {
		Log(context.TODO(), "test")
	})

	t.Run("error:cast", func(t *testing.T) {
		Log(context.TODO(), "error", Err())
	})

	t.Run("error:value param", func(t *testing.T) {
		Log(context.TODO(), "error", "error")
	})

	t.Run("error:non fatal", func(t *testing.T) {
		Log(context.TODO(), "error", ErrBadRequest())
	})

	t.Run("error:trace", func(t *testing.T) {
		Log(context.TODO(), "error", ErrBadRequest())
	})

	t.Run("error:fatal", func(t *testing.T) {
		Log(context.TODO(), "fatal", Err())
	})
}
