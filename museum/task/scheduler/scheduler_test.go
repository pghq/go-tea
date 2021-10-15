package scheduler

import (
	"context"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/pghq/go-eque/eque"
	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/diagnostic/log"
	"github.com/pghq/go-museum/museum/internal"
)

func TestNew(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		queue := NewEqueQueue(t)
		s := New(queue)
		assert.NotNil(t, s)
		assert.Equal(t, s.queue, queue)
		assert.Equal(t, DefaultInterval, s.interval)
		assert.Equal(t, DefaultEnqueueTimeout, s.enqueueTimeout)
		assert.Equal(t, DefaultDequeueTimeout, s.dequeueTimeout)
		assert.Empty(t, s.tasks)
	})
}

func TestScheduler_Every(t *testing.T) {
	t.Run("can set new value", func(t *testing.T) {
		queue := NewEqueQueue(t)
		s := New(queue).Every(time.Second)
		assert.NotNil(t, s)
		assert.Equal(t, time.Second, s.interval)
	})
}

func TestScheduler_EnqueueTimeout(t *testing.T) {
	t.Run("can set new value", func(t *testing.T) {
		queue := NewEqueQueue(t)
		s := New(queue).EnqueueTimeout(time.Second)
		assert.NotNil(t, s)
		assert.Equal(t, time.Second, s.enqueueTimeout)
	})
}

func TestScheduler_DequeueTimeout(t *testing.T) {
	t.Run("can set new value", func(t *testing.T) {
		queue := NewEqueQueue(t)
		s := New(queue).DequeueTimeout(time.Second)
		assert.NotNil(t, s)
		assert.Equal(t, time.Second, s.dequeueTimeout)
	})
}

func TestScheduler_Add(t *testing.T) {
	log.Writer(io.Discard)

	t.Run("raises missing id errors", func(t *testing.T) {
		queue := NewEqueQueue(t)
		task := NewTask("")
		s := New(queue).Add(task)
		assert.NotNil(t, s)
		assert.Empty(t, s.tasks)
	})

	t.Run("can enqueue", func(t *testing.T) {
		queue := NewEqueQueue(t)
		task := NewTask("test")
		s := New(queue).Add(task)
		assert.NotNil(t, s)
		assert.NotEmpty(t, s.tasks)
		assert.Len(t, s.tasks, 1)
		assert.Equal(t, s.tasks[task.Id], task)
	})

	t.Run("does not add duplicates", func(t *testing.T) {
		queue := NewEqueQueue(t)
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
	t.Run("raises enqueue errors", func(t *testing.T) {
		task := NewTask("test")
		queue := NewEqueQueue(t)
		queue.Expect("Enqueue", context.TODO(), "test", task).
			Return(errors.NewBadRequest("an error has occurred"))
		defer queue.Assert(t)

		done := make(chan struct{}, 1)
		s := New(queue).Add(task).Notify(func(t *Task) {
			done <- struct{}{}
		})
		go s.Start()
		<-done
		s.Stop()
		assert.False(t, task.IsComplete())
	})

	t.Run("ignores tasks not ready yet", func(t *testing.T) {
		queue := NewEqueQueue(t)
		task := NewTask("test")
		_ = task.SetRecurrence("DTSTART=99990101T000000Z;FREQ=DAILY")

		done := make(chan struct{}, 1)
		s := New(queue).Add(task).Notify(func(t *Task) {
			done <- struct{}{}
		})

		go s.Start()
		<-done
		s.Stop()
		assert.False(t, task.IsComplete())
		assert.Equal(t, 0, task.Occurrences())
	})

	t.Run("schedules tasks that are ready", func(t *testing.T) {
		task := NewTask("test")

		queue := NewEqueQueue(t)
		queue.Expect("Enqueue", context.TODO(), "test", task).
			Return(nil)
		defer queue.Assert(t)

		done := make(chan struct{}, 1)
		s := New(queue).Add(task).Notify(func(t *Task) {
			done <- struct{}{}
		})
		go s.Start()
		<-done
		s.Stop()
		assert.True(t, task.IsComplete())
		assert.Equal(t, 1, task.Occurrences())
	})
}

func TestScheduler_Worker(t *testing.T) {
	t.Run("raises dequeue errors", func(t *testing.T) {
		task := NewTask("test")

		queue := NewEqueQueue(t)
		queue.Expect("Enqueue", context.TODO(), "test", task).
			Return(nil)
		queue.Expect("Dequeue", context.TODO()).
			Return(nil, errors.New("an error has occurred"))
		defer queue.Assert(t)

		done := make(chan struct{}, 1)
		scheduled := make(chan struct{}, 1)
		s := New(queue).Add(task).
			Notify(func(t *Task) {
				scheduled <- struct{}{}
			}).
			NotifyWorker(func(msg eque.Message) {
				done <- struct{}{}
			})
		defer s.Stop()
		go s.Start()

		<-scheduled

		w := s.Worker(func(_ *Task) {})
		go w.Start()
		defer w.Stop()
		<-done
	})

	t.Run("raises no messages errors", func(t *testing.T) {
		task := NewTask("test")

		queue := NewEqueQueue(t)
		queue.Expect("Enqueue", context.TODO(), "test", task).
			Return(nil)
		queue.Expect("Dequeue", context.TODO()).
			Return(nil, eque.ErrNoMessages)
		defer queue.Assert(t)

		scheduled := make(chan struct{}, 1)
		done := make(chan struct{}, 1)
		s := New(queue).Add(task).
			Notify(func(t *Task) {
				scheduled <- struct{}{}
			}).
			NotifyWorker(func(msg eque.Message) {
				done <- struct{}{}
			})
		defer s.Stop()
		go s.Start()

		w := s.Worker(func(_ *Task) {})
		<-scheduled
		go w.Start()
		defer w.Stop()
		<-done
	})

	t.Run("raises message errors", func(t *testing.T) {
		msg := NewEqueMessage(t)
		msg.Expect("Id").
			Return("test")
		msg.Expect("Decode", &Task{}).
			Return(errors.NewBadRequest("an error has occurred"))
		msg.Expect("Ack", context.TODO()).
			Return(errors.NewBadRequest("an error has occurred"))
		defer msg.Assert(t)

		task := NewTask("test")

		queue := NewEqueQueue(t)
		queue.Expect("Enqueue", context.TODO(), "test", task).
			Return(nil)
		queue.Expect("Dequeue", context.TODO()).
			Return(msg, nil)
		defer queue.Assert(t)

		scheduled := make(chan struct{}, 1)
		done := make(chan struct{}, 1)
		s := New(queue).Add(task).Add(task).
			Notify(func(t *Task) {
				scheduled <- struct{}{}
			}).
			NotifyWorker(func(msg eque.Message) {
				if msg != nil{
					done <- struct{}{}
				}
			})
		defer s.Stop()
		go s.Start()

		w := s.Worker(func(_ *Task) {})
		<-scheduled
		go w.Start()
		defer w.Stop()
		<-done
	})

	t.Run("can process tasks", func(t *testing.T) {
		msg := NewEqueMessage(t)
		msg.Expect("Id").
			Return("test")
		msg.Expect("Decode", &Task{}).
			Return(nil)
		msg.Expect("Ack", context.TODO()).
			Return(nil)
		defer msg.Assert(t)

		task := NewTask("test")

		queue := NewEqueQueue(t)
		queue.Expect("Enqueue", context.TODO(), "test", task).
			Return(nil)
		queue.Expect("Dequeue", context.TODO()).
			Return(msg, nil)
		defer queue.Assert(t)

		scheduled := make(chan struct{}, 1)
		done := make(chan struct{}, 1)
		s := New(queue).Add(task).
			Notify(func(t *Task) {
				scheduled <- struct{}{}
			}).
			NotifyWorker(func(msg eque.Message) {
				if msg != nil{
					done <- struct{}{}
				}
			})
		defer s.Stop()
		go s.Start()

		w := s.Worker(func(_ *Task) {})
		<-scheduled
		go w.Start()
		defer w.Stop()
		<-done
	})
}

func TestTask_CanSchedule(t *testing.T) {
	t.Run("tasks that have do not schedule", func(t *testing.T) {
		task := NewTask("test")
		_ = task.SetRecurrence("UNTIL=19700101T000000Z;FREQ=DAILY")
		canSchedule := task.CanSchedule(time.Now())
		assert.False(t, canSchedule)
	})

	t.Run("tasks with bad recurrence do not schedule", func(t *testing.T) {
		task := NewTask("test")
		task.Schedule.Recurrence = "DAILY"
		canSchedule := task.CanSchedule(time.Now())
		assert.False(t, canSchedule)
	})

	t.Run("tasks that have already reached limit do not schedule", func(t *testing.T) {
		task := NewTask("test")
		_ = task.SetRecurrence("DTSTART=99990101T000000Z;FREQ=DAILY;COUNT=1")
		task.Schedule.Count = 1
		canSchedule := task.CanSchedule(time.Now())
		assert.False(t, canSchedule)
	})

	t.Run("schedules tasks", func(t *testing.T) {
		task := NewTask("test")
		_ = task.SetRecurrence("FREQ=DAILY;COUNT=1")
		canSchedule := task.CanSchedule(time.Now())
		assert.True(t, canSchedule)
	})
}

func TestTask_IsComplete(t *testing.T) {
	t.Run("tasks that have ended are complete", func(t *testing.T) {
		task := NewTask("test")
		_ = task.SetRecurrence("UNTIL=19700101T000000Z;FREQ=DAILY")
		isComplete := task.IsComplete()
		assert.True(t, isComplete)
	})

	t.Run("tasks that have reached limit are complete", func(t *testing.T) {
		task := NewTask("test")
		_ = task.SetRecurrence("DTSTART=99990101T000000Z;FREQ=DAILY;COUNT=1")
		task.Schedule.Count = 1
		isComplete := task.IsComplete()
		assert.True(t, isComplete)
	})

	t.Run("tasks with bad recurrence are complete", func(t *testing.T) {
		task := NewTask("test")
		task.Schedule.Recurrence = "DAILY"
		isComplete := task.IsComplete()
		assert.True(t, isComplete)
	})
}

func TestTask_SetRecurrence(t *testing.T) {
	t.Run("does not set task with bad recurrence", func(t *testing.T) {
		task := NewTask("test")
		err := task.SetRecurrence("DAILY")
		assert.NotNil(t, err)
		assert.Empty(t, task.Schedule.Recurrence)
	})
}

type EqueQueue struct {
	internal.Mock
	t       *testing.T
	values chan interface{}
}

func (e *EqueQueue) Dequeue(ctx context.Context) (eque.Message, error) {
	e.t.Helper()

	select{
		case <-e.values:
			res := e.Call(e.t, ctx)
			if len(res) != 2{
				e.Fatalf(e.t, "length of return values for Dequeue is not equal to 2")
			}

			if res[1] != nil{
				err, ok := res[1].(error)
				if !ok{
					e.Fatalf(e.t,"return value #2 of Dequeue is not an error")
				}

				return nil, err
			}

			if res[0] != nil{
				message, ok := res[0].(eque.Message)
				if !ok{
					e.Fatalf(e.t,"return value #1 of Dequeue is not a eque.Message")
				}
				return message, nil
			}

			return nil, nil
	default:
		return nil, eque.ErrNoMessages
	}
}

func (e *EqueQueue) Enqueue(ctx context.Context, id string, value interface{}) error {
	e.t.Helper()
	select {
		case e.values <- value:
			res := e.Call(e.t, ctx, id, value)
			if len(res) != 1{
				e.Fatalf(e.t, "length of return values for Enqueue is not equal to 1")
			}

			if res[0] != nil{
				err, ok := res[0].(error)
				if !ok{
					e.Fatalf(e.t,"return value #1 of Enqueue is not an error")
				}
				return err
			}

			return nil
	default:
		return eque.ErrAcquireLockFailed
	}
}

func NewEqueQueue(t *testing.T) *EqueQueue {
	q := EqueQueue{
		t: t,
		values: make(chan interface{}, 1),
	}

	return &q
}

type EqueMessage struct {
	internal.Mock
	t   *testing.T
	lock sync.Mutex
}

func (m *EqueMessage) Id() string {
	m.t.Helper()
	res := m.Call(m.t)
	if len(res) != 1{
		m.Fatalf(m.t, "length of return values for Id is not equal to 1")
	}

	id, ok := res[0].(string)
	if !ok{
		m.Fatalf(m.t,"return value #1 of Id is not a string")
	}

	return id
}

func (m *EqueMessage) Ack(ctx context.Context) error {
	m.t.Helper()
	res := m.Call(m.t, ctx)
	if len(res) != 1{
		m.Fatalf(m.t, "length of return values for Ack is not equal to 1")
	}

	if res[0] != nil{
		err, ok := res[0].(error)
		if !ok{
			m.Fatalf(m.t,"return value #1 of Ack is not an error")
		}

		return err
	}

	return nil
}

func (m *EqueMessage) Reject(ctx context.Context) error {
	m.t.Helper()
	res := m.Call(m.t, ctx)
	if len(res) != 1{
		m.Fatalf(m.t, "length of return values for Reject is not equal to 1")
	}

	if res[0] != nil{
		err, ok := res[0].(error)
		if !ok{
			m.Fatalf(m.t,"return value #1 of Reject is not an error")
		}

		return err
	}

	return nil
}

func (m *EqueMessage) Decode(v interface{}) error {
	m.t.Helper()
	res := m.Call(m.t, v)
	if len(res) != 1{
		m.Fatalf(m.t, "length of return values for Decode is not equal to 1")
	}

	if res[0] != nil{
		err, ok := res[0].(error)
		if !ok{
			m.Fatalf(m.t,"return value #1 of Decode is not an error")
		}

		return err
	}

	return nil
}

func NewEqueMessage(t *testing.T) *EqueMessage {
	m := EqueMessage{
		t: t,
	}

	return &m
}
