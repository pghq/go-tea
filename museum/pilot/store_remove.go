package pilot

import (
	"context"
	"testing"

	"github.com/pghq/go-museum/museum/internal"
	"github.com/pghq/go-museum/museum/store"
)

var (
	_ store.Remove = NewRemove(nil)
)

func (c *Store) Remove() store.Remove {
	c.t.Helper()
	res := c.Call(c.t)
	if len(res) != 1 {
		c.Fatalf(c.t, "length of return values for Remove is not equal to 1")
	}

	remove, ok := res[0].(store.Remove)
	if !ok {
		c.Fatalf(c.t, "return value #1 of Remove is not a store.Remove")
	}

	return remove
}

type Remove struct {
	internal.Mock
	t *testing.T
}

func (r *Remove) Statement() (string, []interface{}, error) {
	r.t.Helper()
	res := r.Call(r.t)
	if len(res) != 3 {
		r.Fatalf(r.t, "length of return values for Remove is not equal to 3")
	}

	if res[2] != nil {
		err, ok := res[2].(error)
		if !ok {
			r.Fatalf(r.t, "return value #2 of Statement is not an error")
		}
		return "", nil, err
	}

	statement, ok := res[0].(string)
	if !ok {
		r.Fatalf(r.t, "return value #1 of Statement is not a store.Remove")
	}

	if res[1] != nil {
		args, ok := res[1].([]interface{})
		if !ok {
			r.Fatalf(r.t, "return value #2 of Statement is not an []interface{}")
		}
		return statement, args, nil
	}

	return statement, nil, nil
}

func (r *Remove) Filter(filter store.Filter) store.Remove {
	r.t.Helper()
	res := r.Call(r.t, filter)
	if len(res) != 1 {
		r.Fatalf(r.t, "length of return values for Filter is not equal to 1")
	}

	remove, ok := res[0].(store.Remove)
	if !ok {
		r.Fatalf(r.t, "return value #1 of Filter is not a store.Remove")
	}

	return remove
}

func (r *Remove) Order(by string) store.Remove {
	r.t.Helper()
	res := r.Call(r.t, by)
	if len(res) != 1 {
		r.Fatalf(r.t, "length of return values for Order is not equal to 1")
	}

	remove, ok := res[0].(store.Remove)
	if !ok {
		r.Fatalf(r.t, "return value #1 of Order is not a store.Remove")
	}

	return remove
}

func (r *Remove) First(first int) store.Remove {
	r.t.Helper()
	res := r.Call(r.t, first)
	if len(res) != 1 {
		r.Fatalf(r.t, "length of return values for First is not equal to 1")
	}

	remove, ok := res[0].(store.Remove)
	if !ok {
		r.Fatalf(r.t, "return value #1 of First is not a store.Remove")
	}

	return remove
}

func (r *Remove) After(key string, value interface{}) store.Remove {
	r.t.Helper()
	res := r.Call(r.t, key, value)
	if len(res) != 1 {
		r.Fatalf(r.t, "length of return values for After is not equal to 1")
	}

	remove, ok := res[0].(store.Remove)
	if !ok {
		r.Fatalf(r.t, "return value #1 of After is not a store.Remove")
	}

	return remove
}

func (r *Remove) Execute(ctx context.Context) (int, error) {
	r.t.Helper()
	res := r.Call(r.t, ctx)
	if len(res) != 2 {
		r.Fatalf(r.t, "length of return values for Execute is not equal to 2")
	}

	if res[1] != nil {
		err, ok := res[1].(error)
		if !ok {
			r.Fatalf(r.t, "return value #2 of Execute is not an error")
		}
		return 0, err
	}

	count, ok := res[0].(int)
	if !ok {
		r.Fatalf(r.t, "return value #1 of Execute is not a int")
	}

	return count, nil
}

func (r *Remove) From(collection string) store.Remove {
	r.t.Helper()
	res := r.Call(r.t, collection)
	if len(res) != 1 {
		r.Fatalf(r.t, "length of return values for From is not equal to 1")
	}

	remove, ok := res[0].(store.Remove)
	if !ok {
		r.Fatalf(r.t, "return value #1 of From is not a store.Remove")
	}

	return remove
}

func NewRemove(t *testing.T) *Remove {
	r := Remove{t: t}

	return &r
}
