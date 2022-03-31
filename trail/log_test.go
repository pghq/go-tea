package trail

import (
	"testing"
)

func init() {
	Testing()
}

func TestLog(t *testing.T) {
	t.Run("can set verbosity", func(t *testing.T) {
		defer SetVerbosity("trail")
		SetVerbosity("debug")
		SetVerbosity("info")
		SetVerbosity("warn")
		SetVerbosity("error")
		SetVerbosity("fatal")
		SetVerbosity("unknown")
	})

	t.Parallel()

	t.Run("debug", func(t *testing.T) {
		Debugf("%s", "message")
	})

	t.Run("info", func(t *testing.T) {
		Infof("%s", "message")
	})

	t.Run("warn", func(t *testing.T) {
		Warnf("%s", "message")
	})

	t.Run("test", func(t *testing.T) {
		OneOff("message")
	})

	t.Run("error:cast", func(t *testing.T) {
		Errorf("%s", "message")
	})

	t.Run("fatal", func(t *testing.T) {
		Fatalf("%s", "message")
	})

	t.Run("error:value param", func(t *testing.T) {
		Error("error")
	})

	t.Run("error:non fatal", func(t *testing.T) {
		Error(NewErrorBadRequest("a message"))
	})

	t.Run("error:trail", func(t *testing.T) {
		Error(NewErrorBadRequest("a message"))
	})

	t.Run("error:fatal", func(t *testing.T) {
		Error(NewError("a message"))
	})

	t.Run("nil", func(t *testing.T) {
		Error(nil)
	})
}
