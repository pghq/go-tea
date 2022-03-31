package trail

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
)

func TestStartSpan(t *testing.T) {
	t.Parallel()

	t.Run("without span parent", func(t *testing.T) {
		child := StartSpan(context.TODO(), "child")
		defer child.Finish()
	})

	t.Run("with span parent", func(t *testing.T) {
		parent := StartSpan(context.TODO(), "parent")
		defer parent.Finish()
		child := StartSpan(parent.Context(), "child")
		defer child.Finish()
	})
}

func TestSpan_Capture(t *testing.T) {
	t.Parallel()

	t.Run("with request", func(t *testing.T) {
		span := StartSpan(context.TODO(), "request", WithSpanRequest(httptest.NewRequest("", "/test", nil)))
		span.Capture(NewError("a message"))
	})
}

func TestSpan_SetField(t *testing.T) {
	t.Parallel()

	t.Run("can set and get", func(t *testing.T) {
		span := StartSpan(context.TODO(), "test")
		defer span.Finish()
		span.Fields.Set("key", "value")
		assert.Equal(t, "value", span.Fields.Get("key"))
	})
}

func TestSpan_Bundle(t *testing.T) {
	t.Run("too many spans", func(t *testing.T) {
		span := StartSpan(context.TODO(), "test")
		defer span.Finish()
		for i := 0; i < maxSpans+1; i++ {
			child := StartSpan(span.Context(), "test")
			child.Finish()
		}
	})
}

func TestWithSpanRequest(t *testing.T) {
	t.Run("detects request id", func(t *testing.T) {
		span := StartSpan(context.TODO(), "test", WithSpanRequest(&http.Request{
			Header: map[string][]string{
				"Request-Id": {uuid.NewString()},
			},
		}))
		defer span.Finish()
	})
}
