package trail

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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
		span := StartSpan(context.TODO(), "request")
		span.Capture(NewError("a message"))
	})
}

func TestSpan_SetTag(t *testing.T) {
	t.Parallel()

	t.Run("can set and get", func(t *testing.T) {
		span := StartSpan(context.TODO(), "test")
		defer span.Finish()
		span.Tags.Set("key", "value")
		assert.NotNil(t, span.sentryHub())
		assert.Equal(t, "value", span.Tags.Get("key"))
	})

	t.Run("can set and get json", func(t *testing.T) {
		span := StartSpan(context.TODO(), "test")
		defer span.Finish()
		span.Tags.SetJSON("json", map[string]interface{}{"key": "value"})
		assert.JSONEq(t, `{"key": "value"}`, span.Tags.Get("json"))
	})

	t.Run("ignore empty value", func(t *testing.T) {
		span := StartSpan(context.TODO(), "test")
		defer span.Finish()
		span.Tags.Set("key", "")
		assert.Equal(t, "", span.Tags.Get("key"))
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

func TestSpan_SetResponse(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		span := StartSpan(context.TODO(), "test")
		defer span.Finish()

		span.SetResponse(httptest.NewRecorder())
	})
}
