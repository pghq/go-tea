package worker

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/log"
)

func TestNew(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		w := New()
		assert.NotNil(t, w)
	})
}

func TestWorker_At(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		w := New().At(time.Second)
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
		w := New().At(time.Nanosecond)
		timings := make(chan time.Time, 1)
		go w.Start(func(ctx context.Context) {
			defer w.Stop()
			timings <- time.Now()
		})

		ticker := time.NewTicker(time.Millisecond)
		select {
		case run := <-timings:
			assert.True(t, time.Now().Sub(run) < time.Millisecond)
		case <-ticker.C:
			t.Fatal("job did not execute in time")
		}
	})

	t.Run("Panic", func(t *testing.T) {
		defer func() {
			if err := recover(); err != nil{
				t.Fatalf("panic not expected: %+v", err)
			}
		}()

		w := New().At(time.Nanosecond)
		defer w.Stop()
		go w.Start(func(ctx context.Context) {
			panic("an error has occurred")
		})

		ticker := time.NewTicker(time.Millisecond)
		<-ticker.C
	})
}
