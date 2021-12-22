package tea

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestNest(t *testing.T) {
	t.Parallel()

	t.Run("without span parent", func(t *testing.T) {
		child := Nest(context.TODO(), "child")
		defer child.End()
	})

	t.Run("with span parent", func(t *testing.T) {
		parent := Start(context.TODO(), "parent")
		defer parent.End()
		child := Nest(parent, "child")
		defer child.End()
	})
}

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
