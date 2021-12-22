package tea

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.NotNil(t, Out())
	})
}

func TestLog(t *testing.T) {
	t.Parallel()

	t.Run("debug", func(t *testing.T) {
		Log(context.TODO(), "debug")
		assert.NotNil(t, Out())
	})

	t.Run("info", func(t *testing.T) {
		Log(context.TODO(), "info")
		assert.NotNil(t, Out())
	})

	t.Run("warn", func(t *testing.T) {
		Log(context.TODO(), "warn")
		assert.NotNil(t, Out())
	})

	t.Run("test", func(t *testing.T) {
		Log(context.TODO(), "test")
		assert.NotNil(t, Out())
	})

	t.Run("error:cast", func(t *testing.T) {
		Log(context.TODO(), "error", Err())
		assert.NotNil(t, Out())
	})

	t.Run("error:value param", func(t *testing.T) {
		Log(context.TODO(), "error", "error")
		assert.NotNil(t, Out())
	})

	t.Run("error:non fatal", func(t *testing.T) {
		Log(context.TODO(), "error", ErrBadRequest())
		assert.Empty(t, Out())
	})

	t.Run("error:trace", func(t *testing.T) {
		SetVerbosity("trace")
		Log(context.TODO(), "error", ErrBadRequest())
		assert.NotNil(t, Out())
	})

	t.Run("error:fatal", func(t *testing.T) {
		Log(context.TODO(), "fatal", Err())
		assert.NotNil(t, Out())
	})
}
