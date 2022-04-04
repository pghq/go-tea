package trail

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
		err := Stacktrace(NewError("an error has occurred"))
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
		assert.NotNil(t, NewErrorf("%s", "err"))
	})
}

func TestErrorNoContent(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		assert.True(t, IsNoContent(ErrorNoContent(NewErrorNoContent("a message"))))
	})
}

func TestErrNotFound(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		assert.True(t, IsNotFound(ErrorNotFound(NewErrorNotFound("a message"))))
	})
}

func TestErrBadRequest(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		assert.True(t, IsBadRequest(ErrorBadRequest(NewErrorBadRequest("a message"))))
	})
}

func TestErrorConflict(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		assert.True(t, IsConflict(ErrorConflict(NewErrorConflict("a message"))))
	})
}

func TestErrorTooManyRequests(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		assert.True(t, IsTooManyRequests(ErrorTooManyRequests(NewErrorTooManyRequests("a message"))))
	})
}

func TestErrorNotAuthorized(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		assert.True(t, IsNotAuthorized(ErrorNotAuthorized(NewErrorNotAuthorized("a message"))))
	})
}

func TestAsError(t *testing.T) {
	t.Parallel()

	t.Run("ok", func(t *testing.T) {
		assert.True(t, AsError(NewError("a message"), NewError("a message")))
	})
}

func TestErrStatus(t *testing.T) {
	t.Parallel()

	t.Run("unknown", func(t *testing.T) {
		assert.Equal(t, 500, StatusCode(errors.New("a message")))
	})
}
