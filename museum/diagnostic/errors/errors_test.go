package errors

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestWrap(t *testing.T) {
	t.Run("ApplicationError", func(t *testing.T){
		err := Wrap(New("an error has occurred"))
		assert.NotNil(t, err)
		assert.Less(t, 1, strings.Count(fmt.Sprintf("%+v", err), "\n"))
	})

	t.Run("InternalError", func(t *testing.T){
		err := Wrap(errors.New("an error has occurred"))
		assert.NotNil(t, err)
		assert.Less(t, 1, strings.Count(fmt.Sprintf("%+v", err), "\n"))
	})
}

func TestNew(t *testing.T) {
	t.Run("Error", func(t *testing.T){
		err := New("an error has occurred")
		assert.NotNil(t, err)
	})
}

func TestHTTP(t *testing.T) {
	t.Run("Error", func(t *testing.T){
		err := HTTP(errors.New("an error has occurred"), http.StatusBadRequest)
		assert.NotNil(t, err)
	})
}

func TestNewHTTP(t *testing.T) {
	t.Run("Error", func(t *testing.T){
		err := NewHTTP("an error has occurred", http.StatusConflict)
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusConflict, ae.code)
	})
}

func TestBadRequest(t *testing.T) {
	t.Run("Error", func(t *testing.T){
		err := BadRequest(errors.New("an error has occurred"))
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, ae.code)
	})
}

func TestNewBadRequest(t *testing.T) {
	t.Run("Error", func(t *testing.T){
		err := NewBadRequest("an error has occurred")
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, ae.code)
	})
}

func TestNoContent(t *testing.T) {
	t.Run("Error", func(t *testing.T){
		err := NoContent(errors.New("an error has occurred"))
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNoContent, ae.code)
	})
}

func TestNewNoContent(t *testing.T) {
	t.Run("Error", func(t *testing.T){
		err := NewNoContent("an error has occurred")
		assert.NotNil(t, err)
		ae, ok := err.(*applicationError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusNoContent, ae.code)
	})
}

func TestAs(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		err := New("an error has occurred")
		as := As(&applicationError{}, &err)
		assert.True(t, as)
	})
}

func TestIsFatal(t *testing.T) {
	t.Run("FatalError", func(t *testing.T){
		err := New("an error has occurred")
		assert.True(t, IsFatal(err))
	})

	t.Run("NonFatalError", func(t *testing.T){
		err := NewHTTP("an error has occurred", http.StatusNoContent)
		assert.False(t, IsFatal(err))
	})
}

func TestStatusCode(t *testing.T) {
	t.Run("NoContent", func(t *testing.T){
		err := NewHTTP("an error has occurred", http.StatusNoContent)
		assert.Equal(t, http.StatusNoContent, StatusCode(err))
	})

	t.Run("ContextCancelled", func(t *testing.T){
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		assert.Equal(t, http.StatusBadRequest, StatusCode(ctx.Err()))
	})

	t.Run("DeadlineExceeded", func(t *testing.T){
		ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
		defer cancel()
		time.Sleep(time.Microsecond)
		assert.Equal(t, http.StatusRequestTimeout, StatusCode(ctx.Err()))
	})

	t.Run("Internal", func(t *testing.T){
		assert.Equal(t, http.StatusInternalServerError, StatusCode(errors.New("an error has occurred")))
	})
}

func TestEmit(t *testing.T) {
	t.Run("FatalError", func(t *testing.T){
		err := New("an error has occurred")
		Emit(err)
	})

	t.Run("NonFatalError", func(t *testing.T){
		err := NewHTTP("an error has occurred", http.StatusNoContent)
		Emit(err)
	})
}

func TestEmitHTTP(t *testing.T) {
	req := httptest.NewRequest("GET", "/tests", nil)

	t.Run("FatalError", func(t *testing.T){
		err := New("an error has occurred")
		EmitHTTP(httptest.NewRecorder(), req, err)
	})

	t.Run("NonFatalError", func(t *testing.T){
		err := NewHTTP("an error has occurred", http.StatusNoContent)
		EmitHTTP(httptest.NewRecorder(), req, err)
	})
}

func TestRecover(t *testing.T) {
	t.Run("Panic", func(t *testing.T){
		defer func(){
			Recover(recover())
		}()

		panic("an error has occurred")
	})
}

func TestInit(t *testing.T) {
	t.Run("ConfigError", func(t *testing.T) {
		conf := MonitorConfig{
			Dsn: "https://localhost",
		}

		err := Init(conf)
		assert.NotNil(t, err)
	})

	t.Run("NoError", func(t *testing.T) {
		conf := MonitorConfig{
			Version: "0.0.1",
			Environment: "test",
			Dsn: "https://12345@678910.ingest.sentry.io/1",
			FlushTimeout: time.Second,
		}

		err := Init(conf)
		assert.Nil(t, err)
	})
}