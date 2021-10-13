package test

import (
	"context"
	"sync"
	"testing"

	"github.com/pghq/go-eque/eque"
	"github.com/stretchr/testify/assert"
)

type EqueQueue struct {
	t       *testing.T
	err     error
	lock    sync.Mutex
	dequeue struct {
		calls int
		ctx   context.Context
		msg   eque.Message
		err   error
	}
	enqueue struct {
		calls int
		ctx   context.Context
		id    string
		value interface{}
	}
	values chan interface{}
}

func (e *EqueQueue) Dequeue(ctx context.Context) (eque.Message, error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.t.Helper()
	if e.err != nil {
		return nil, e.err
	}

	if e.dequeue.err != nil {
		return nil, e.dequeue.err
	}

	e.dequeue.calls -= 1

	if e.dequeue.ctx != nil {
		assert.Equal(e.t, e.dequeue.ctx, ctx)
	}

	select {
	case <-e.values:
		return e.dequeue.msg, nil
	default:
		return nil, nil
	}
}

func (e *EqueQueue) ExpectDequeue(ctx context.Context) *EqueQueue {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.err = nil
	e.dequeue.calls += 1
	e.dequeue.ctx = ctx
	return e
}

func (e *EqueQueue) ReturnDequeue(msg eque.Message, err error) *EqueQueue {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.dequeue.err = err
	e.dequeue.msg = msg
	return e
}

func (e *EqueQueue) Enqueue(ctx context.Context, id string, value interface{}) error {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.t.Helper()
	if e.err != nil {
		return e.err
	}

	e.enqueue.calls -= 1

	if e.enqueue.ctx != nil {
		assert.Equal(e.t, e.enqueue.ctx, ctx)
	}

	if e.enqueue.id != "" {
		assert.Equal(e.t, e.enqueue.id, id)
	}

	if e.enqueue.value != nil {
		assert.Equal(e.t, e.enqueue.value, value)
	}

	select {
	case e.values <- value:
	default:
	}

	return nil
}

func (e *EqueQueue) ExpectEnqueue(ctx context.Context, id string, value interface{}) *EqueQueue {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.err = nil
	e.enqueue.ctx = ctx
	e.enqueue.id = id
	e.enqueue.value = value
	return e
}

func (e *EqueQueue) Error(err error) *EqueQueue {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.err = err
	return e
}

func (e *EqueQueue) Assert() {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.t.Helper()
	if e.dequeue.calls > 0 {
		e.t.Fatal("not enough calls to dequeue")
	}

	if e.enqueue.calls > 0 {
		e.t.Fatal("not enough calls to enqueue")
	}
}

func NewEqueQueue(t *testing.T) *EqueQueue {
	q := EqueQueue{
		t:      t,
		values: make(chan interface{}, 1),
	}

	return &q
}

type EqueMessage struct {
	t   *testing.T
	err error
	id  struct {
		calls int
	}
	ack struct {
		calls int
		ctx   context.Context
	}
	reject struct {
		calls int
		ctx   context.Context
	}
	decode struct {
		calls int
		v     interface{}
	}
	lock sync.Mutex
}

func (m *EqueMessage) Id() string {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.id.calls -= 1
	return ""
}

func (m *EqueMessage) ExpectId() *EqueMessage {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.id.calls += 1
	return m
}

func (m *EqueMessage) Ack(ctx context.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.t.Helper()
	if m.err != nil {
		return m.err
	}

	m.ack.calls -= 1
	if m.ack.ctx != nil {
		assert.Equal(m.t, m.ack.ctx, ctx)
	}

	return nil
}

func (m *EqueMessage) ExpectAck(ctx context.Context) *EqueMessage {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.ack.calls += 1
	m.ack.ctx = ctx
	return m
}

func (m *EqueMessage) Reject(ctx context.Context) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.t.Helper()
	if m.err != nil {
		return m.err
	}

	m.reject.calls += 1
	if m.reject.ctx != nil {
		assert.Equal(m.t, m.reject.ctx, ctx)
	}

	return nil
}

func (m *EqueMessage) ExpectReject(ctx context.Context) *EqueMessage {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.reject.calls += 1
	m.reject.ctx = ctx
	return m
}

func (m *EqueMessage) Decode(v interface{}) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.t.Helper()
	if m.err != nil {
		return m.err
	}

	m.decode.calls -= 1
	if m.decode.v != nil {
		assert.IsType(m.t, m.decode.v, v)
	}

	return nil
}

func (m *EqueMessage) ExpectDecode(v interface{}) *EqueMessage {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.decode.calls += 1
	m.decode.v = v
	return m
}

func (m *EqueMessage) Error(err error) *EqueMessage {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.err = err
	return m
}

func (m *EqueMessage) Assert() {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.t.Helper()
	if m.id.calls != 0 {
		m.t.Fatal("not enough calls to id")
	}

	if m.ack.calls != 0 {
		m.t.Fatal("not enough calls to ack")
	}

	if m.reject.calls != 0 {
		m.t.Fatal("not enough calls to reject")
	}

	if m.decode.calls != 0 {
		m.t.Fatal("not enough calls to decode")
	}
}

func NewEqueMessage(t *testing.T) *EqueMessage {
	m := EqueMessage{
		t: t,
	}

	return &m
}
