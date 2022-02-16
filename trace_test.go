package tea

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-version"
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
		m := Trace(&version.Version{})
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("panic")
		})).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/test", nil))
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
