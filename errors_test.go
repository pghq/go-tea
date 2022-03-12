package tea

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	t.Parallel()

	t.Run("adds stacktrace to application errors", func(t *testing.T) {
		err := Stacktrace(Err("an error has occurred"))
		assert.NotNil(t, err)
		assert.Less(t, 1, strings.Count(fmt.Sprintf("%+v", err), "\n"))
		assert.Contains(t, err.Error(), "an error has occurred")
	})

	t.Run("adds stacktrace to internal errors", func(t *testing.T) {
		err := Stacktrace(errors.New("an error has occurred"))
		assert.NotNil(t, err)
		assert.Less(t, 1, strings.Count(fmt.Sprintf("%+v", err), "\n"))
		assert.Contains(t, err.Error(), "an error has occurred")
	})

	t.Run("context cancelled", func(t *testing.T) {
		assert.False(t, IsFatal(Stacktrace(context.Canceled)))
	})

	t.Run("context timeout", func(t *testing.T) {
		assert.False(t, IsFatal(Stacktrace(context.DeadlineExceeded)))
	})
}

func TestErrf(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		assert.NotNil(t, Errf("%s", "err"))
	})
}

func TestErrNoContent(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		assert.NotNil(t, AsErrNoContent(ErrNoContent()))
	})
}

func TestErrNotFound(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		assert.True(t, IsNotFound(AsErrNotFound(ErrNotFound())))
	})
}

func TestErrBadRequest(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		assert.True(t, IsBadRequest(AsErrBadRequest(ErrBadRequest())))
	})
}

func TestAsError(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		assert.True(t, AsError(Err(), Err()))
	})
}

func TestErrStatus(t *testing.T) {
	t.Parallel()

	t.Run("unknown", func(t *testing.T) {
		assert.Equal(t, 500, ErrStatus(errors.New("")))
	})
}
