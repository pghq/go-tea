package pilot

import (
	"testing"

	"github.com/pghq/go-museum/museum/internal"
	"github.com/pghq/go-museum/museum/store"
)

var (
	_ store.Client  = NewStore(nil)
	_ store.Snapper = NewSnapper(nil)
	_ store.Cursor  = NewCursor(nil)
)

type Store struct {
	internal.Mock
	t *testing.T
}

func (c *Store) IsConnected() bool {
	c.t.Helper()
	res := c.Call(c.t)
	if len(res) != 1 {
		c.Fatalf(c.t, "length of return values for IsConnected is not equal to 1")
	}

	connected, ok := res[0].(bool)
	if !ok {
		c.Fatalf(c.t, "return value #1 of IsConnected is not a bool")
	}

	return connected
}

func NewDisconnectedStore(t *testing.T) *Store {
	c := Store{t: t}

	return &c
}

func NewStore(t *testing.T) *Store {
	c := Store{t: t}
	c.Expect("IsConnected").Return(true)

	return &c
}

type Snapper struct {
	internal.Mock
	t *testing.T
}

func (s *Snapper) Snapshot() map[string]interface{} {
	s.t.Helper()
	res := s.Call(s.t)
	if len(res) != 1 {
		s.Fatalf(s.t, "length of return values for Snapshot is not equal to 1")
	}

	snapshot, ok := res[0].(map[string]interface{})
	if !ok {
		s.Fatalf(s.t, "return value #1 of Snapshot is not a map[string]interface{}")
	}

	return snapshot
}

func NewSnapper(t *testing.T) *Snapper {
	s := Snapper{t: t}

	return &s
}

type Cursor struct {
	internal.Mock
	t *testing.T
}

func (c *Cursor) Next() bool {
	c.t.Helper()
	res := c.Call(c.t)
	if len(res) != 1 {
		c.Fatalf(c.t, "length of return values for Next is not equal to 1")
	}

	next, ok := res[0].(bool)
	if !ok {
		c.Fatalf(c.t, "return value #1 of Next is not a bool")
	}

	return next
}

func (c *Cursor) Decode(values ...interface{}) error {
	c.t.Helper()
	res := c.Call(c.t, values...)
	if len(res) != 1 {
		c.Fatalf(c.t, "length of return values for Decode is not equal to 1")
	}

	if res[0] != nil {
		err, ok := res[0].(error)
		if !ok {
			c.Fatalf(c.t, "return value #1 of Decode is not a error")
		}
		return err
	}

	return nil
}

func (c *Cursor) Close() {
	c.t.Helper()
	res := c.Call(c.t)
	if len(res) != 0 {
		c.Fatalf(c.t, "length of return values for Close is not equal to 0")
	}
}

func (c *Cursor) Error() error {
	c.t.Helper()
	res := c.Call(c.t)
	if len(res) != 1 {
		c.Fatalf(c.t, "length of return values for Error is not equal to 1")
	}

	if res[0] != nil {
		err, ok := res[0].(error)
		if !ok {
			c.Fatalf(c.t, "return value #1 of Error is not a error")
		}
		return err
	}

	return nil
}

func NewCursor(t *testing.T) *Cursor {
	c := Cursor{t: t}

	return &c
}
