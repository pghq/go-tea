package worker

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/diagnostic/log"
)

func TestNew(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		w := New()
		assert.NotNil(t, w)
		assert.True(t, w.IsStopped())
		assert.False(t, w.IsStopping())
	})
}

func TestWorker_Every(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		w := New().Every(time.Second)
		assert.NotNil(t, w)
		assert.Equal(t, w.interval, time.Second)
	})
}

func TestWorker_Concurrent(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		w := New().Concurrent(5)
		assert.NotNil(t, w)
		assert.Equal(t, w.instances, 5)
	})
}

func TestWorker_Start(t *testing.T) {
	log.Writer(io.Discard)

	t.Run("NoError", func(t *testing.T) {
		done := make(chan struct{}, 1)
		job := func(ctx context.Context, stop func()) {
			stop()
			done <- struct{}{}
		}
		w := New(job).Every(time.Nanosecond)
		go w.Start(func(){
			assert.False(t, w.IsStopped())
		})
		<-done
	})

	t.Run("Panic", func(t *testing.T) {
		defer func() {
			if err := recover(); err != nil{
				t.Fatalf("panic not expected: %+v", err)
			}
		}()

		job := func(ctx context.Context, _ func()) {
			panic("an error has occurred")
		}

		w := New(job).Every(time.Nanosecond)
		defer w.Stop()
		go w.Start()
		
		<-time.After(time.Millisecond)
	})
}
