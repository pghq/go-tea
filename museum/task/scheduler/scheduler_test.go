package scheduler

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/diagnostic/log"
	"github.com/pghq/go-museum/museum/internal/test"
)

func TestNew(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		queue := test.NewEqueQueue(t)
		s := New(queue)
		assert.NotNil(t, s)
		assert.Equal(t, s.queue, queue)
		assert.Equal(t, DefaultInterval, s.interval)
		assert.Equal(t, DefaultEnqueueTimeout, s.enqueueTimeout)
		assert.Equal(t, DefaultDequeueTimeout, s.dequeueTimeout)
		assert.Empty(t, s.tasks)
		assert.True(t, s.IsStopped())
		assert.False(t, s.IsStopping())
	})
}

func TestScheduler_Every(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		queue := test.NewEqueQueue(t)
		s := New(queue).Every(time.Second)
		assert.NotNil(t, s)
		assert.Equal(t, time.Second, s.interval)
	})
}

func TestScheduler_EnqueueTimeout(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		queue := test.NewEqueQueue(t)
		s := New(queue).EnqueueTimeout(time.Second)
		assert.NotNil(t, s)
		assert.Equal(t, time.Second, s.enqueueTimeout)
	})
}

func TestScheduler_DequeueTimeout(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		queue := test.NewEqueQueue(t)
		s := New(queue).DequeueTimeout(time.Second)
		assert.NotNil(t, s)
		assert.Equal(t, time.Second, s.dequeueTimeout)
	})
}

func TestScheduler_Add(t *testing.T) {
	log.Writer(io.Discard)

	t.Run("NoId", func(t *testing.T) {
		queue := test.NewEqueQueue(t)
		task := NewTask("")
		s := New(queue).Add(task)
		assert.NotNil(t, s)
		assert.Empty(t, s.tasks)
	})

	t.Run("NoError", func(t *testing.T) {
		queue := test.NewEqueQueue(t)
		task := NewTask("test")
		s := New(queue).Add(task)
		assert.NotNil(t, s)
		assert.NotEmpty(t, s.tasks)
		assert.Len(t, s.tasks, 1)
		assert.Equal(t, s.tasks[task.Id], task)
	})

	t.Run("Duplicate", func(t *testing.T) {
		queue := test.NewEqueQueue(t)
		task := NewTask("test")
		s := New(queue).Add(task).Add(task)
		assert.NotNil(t, s)
		assert.NotEmpty(t, s.tasks)
		assert.Len(t, s.tasks, 1)
		assert.Equal(t, s.tasks[task.Id], task)
	})
}

func TestScheduler_Start(t *testing.T) {
	log.Writer(io.Discard)
	log.Writer(os.Stderr)
	t.Run("EnqueueError", func(t *testing.T) {
		queue := test.NewEqueQueue(t).Error(errors.NewBadRequest("an error has occurred"))
		task := NewTask("test")
		s := New(queue).Every(time.Nanosecond).Add(task)
		defer s.Stop()
		go s.Start(func() {
			assert.False(t, s.IsStopped())
		})
		<-time.After(time.Millisecond)
		assert.False(t, task.IsComplete())
	})

	t.Run("NotYet", func(t *testing.T) {
		queue := test.NewEqueQueue(t)
		task := NewTask("test")
		_ = task.SetRecurrence("DTSTART=99990101T000000Z;FREQ=DAILY")
		s := New(queue).Every(time.Nanosecond).Add(task)
		go s.Start()
		defer s.Stop()
		<-time.After(time.Millisecond)
		assert.False(t, task.IsComplete())
		assert.Equal(t, 0, task.Occurrences())
	})

	t.Run("NoError", func(t *testing.T) {
		task := NewTask("test")
		queue := test.NewEqueQueue(t).ExpectEnqueue(nil, "test", task)
		defer queue.Assert()
		s := New(queue).Every(time.Nanosecond).Add(task)
		go s.Start(func() { assert.False(t, s.IsStopped()) })
		<-time.After(time.Millisecond)
		s.Stop()
		assert.True(t, task.IsComplete())
		assert.Equal(t, 1, task.Occurrences())
		assert.True(t, s.IsStopped())
	})
}

func TestScheduler_Worker(t *testing.T) {
	t.Run("DequeueError", func(t *testing.T) {
		task := NewTask("test")
		queue := test.NewEqueQueue(t).ReturnDequeue(nil, errors.NewBadRequest("an error has occurred"))
		defer queue.Assert()
		s := New(queue).Every(time.Nanosecond).Add(task)
		go s.Start()
		defer s.Stop()
		done := make(chan struct{}, 1)
		job := func(got *Task) {
			assert.Equal(t, task, got)
			done <- struct{}{}
		}
		w := s.Worker(job).Every(time.Nanosecond)
		go w.Start()
		defer w.Stop()
		<-time.After(100 * time.Millisecond)
		assert.Empty(t, done)
	})

	t.Run("NoMessages", func(t *testing.T) {
		task := NewTask("test")
		queue := test.NewEqueQueue(t)
		defer queue.Assert()
		s := New(queue).Every(time.Nanosecond).Add(task)
		go s.Start()
		defer s.Stop()
		done := make(chan struct{}, 1)
		job := func(got *Task) {
			assert.Equal(t, task, got)
			done <- struct{}{}
		}
		w := s.Worker(job).Every(time.Nanosecond)
		go w.Start()
		defer w.Stop()
		<-time.After(100 * time.Millisecond)
		assert.Empty(t, done)
	})

	t.Run("MessageError", func(t *testing.T) {
		task := NewTask("test")
		msg := test.NewEqueMessage(t).Error(errors.NewBadRequest("an error has occurred"))
		queue := test.NewEqueQueue(t).ReturnDequeue(msg, nil)
		defer queue.Assert()
		s := New(queue).Every(time.Nanosecond).Add(task)
		go s.Start()
		defer s.Stop()
		done := make(chan struct{}, 1)
		job := func(got *Task) {
			assert.Equal(t, task, got)
			done <- struct{}{}
		}
		w := s.Worker(job).Every(time.Nanosecond)
		go w.Start()
		defer w.Stop()
		<-time.After(100 * time.Millisecond)
		assert.Empty(t, done)
	})

	t.Run("NoError", func(t *testing.T) {
		task := NewTask("test")
		msg := test.NewEqueMessage(t).ExpectDecode(&Task{}).ExpectId().ExpectAck(nil)
		defer msg.Assert()
		queue := test.NewEqueQueue(t).ExpectDequeue(nil).ReturnDequeue(msg, nil)
		defer queue.Assert()
		s := New(queue).Every(time.Nanosecond).Add(task)
		go s.Start()
		defer s.Stop()
		done := make(chan struct{}, 1)
		job := func(got *Task) { done <- struct{}{}}
		w := s.Worker(job).Every(time.Nanosecond)
		go w.Start()
		defer w.Stop()
		<-done
	})
}

func TestTask_CanSchedule(t *testing.T) {
	t.Run("TaskEnded", func(t *testing.T) {
		task := NewTask("test")
		_ = task.SetRecurrence("UNTIL=19700101T000000Z;FREQ=DAILY")
		canSchedule := task.CanSchedule(time.Now())
		assert.False(t, canSchedule)
	})

	t.Run("BadRecurrence", func(t *testing.T) {
		task := NewTask("test")
		task.Schedule.Recurrence = "DAILY"
		canSchedule := task.CanSchedule(time.Now())
		assert.False(t, canSchedule)
	})

	t.Run("TooManyOccurrences", func(t *testing.T) {
		task := NewTask("test")
		_ = task.SetRecurrence("DTSTART=99990101T000000Z;FREQ=DAILY;COUNT=1")
		task.Schedule.Count = 1
		canSchedule := task.CanSchedule(time.Now())
		assert.False(t, canSchedule)
	})

	t.Run("MatchesRecurrence", func(t *testing.T) {
		task := NewTask("test")
		_ = task.SetRecurrence("FREQ=DAILY;COUNT=1")
		canSchedule := task.CanSchedule(time.Now())
		assert.True(t, canSchedule)
	})
}

func TestTask_IsComplete(t *testing.T) {
	t.Run("TaskEnded", func(t *testing.T) {
		task := NewTask("test")
		_ = task.SetRecurrence("UNTIL=19700101T000000Z;FREQ=DAILY")
		isComplete := task.IsComplete()
		assert.True(t, isComplete)
	})

	t.Run("TooManyOccurrences", func(t *testing.T) {
		task := NewTask("test")
		_ = task.SetRecurrence("DTSTART=99990101T000000Z;FREQ=DAILY;COUNT=1")
		task.Schedule.Count = 1
		isComplete := task.IsComplete()
		assert.True(t, isComplete)
	})

	t.Run("BadRecurrence", func(t *testing.T) {
		task := NewTask("test")
		task.Schedule.Recurrence = "DAILY"
		isComplete := task.IsComplete()
		assert.True(t, isComplete)
	})
}

func TestTask_SetRecurrence(t *testing.T) {
	t.Run("BadRecurrence", func(t *testing.T) {
		task := NewTask("test")
		err := task.SetRecurrence("DAILY")
		assert.NotNil(t, err)
		assert.Empty(t, task.Schedule.Recurrence)
	})
}