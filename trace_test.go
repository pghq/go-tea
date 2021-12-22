package tea

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestSpan_Recover(t *testing.T) {
	t.Parallel()

	t.Run("panic", func(t *testing.T) {
		span := Start(context.TODO(), "panic")
		defer span.End()
		panic("panic")
	})
}

func TestSpan_Capture(t *testing.T) {
	t.Parallel()

	t.Run("with request", func(t *testing.T) {
		span := Start(context.TODO(), "request")
		span.SetRequest(httptest.NewRequest("", "/test", nil))
		span.Capture(Err())
	})
}
