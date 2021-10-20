package pilot

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/internal"
	"github.com/pghq/go-museum/museum/store"
)

var (
	_ store.Filter = NewFilter(nil)
)

func (s *Store) Filter() store.Filter {
	s.t.Helper()
	res := s.Call(s.t)
	if len(res) != 1 {
		s.fail(s.t, "unexpected length of return values")
		return nil
	}

	filter, ok := res[0].(store.Filter)
	if !ok {
		s.fail(s.t, "unexpected type of return value")
		return nil
	}

	return filter
}

// Filter is a mock store.Filter
type Filter struct {
	internal.Mock
	t *testing.T
	fail func(v ...interface{})
}

func (f *Filter) BeginsWith(key string, prefix string) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, prefix)
	if len(res) != 1 {
		f.fail(f.t, "unexpected length of return values")
		return nil
	}

	filter, ok := res[0].(store.Filter)
	if !ok {
		f.fail(f.t, "unexpected type of return value")
		return nil
	}

	return filter
}

func (f *Filter) EndsWith(key string, suffix string) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, suffix)
	if len(res) != 1 {
		f.fail(f.t, "unexpected length of return values")
		return nil
	}

	filter, ok := res[0].(store.Filter)
	if !ok {
		f.fail(f.t, "unexpected type of return value")
		return nil
	}

	return filter
}

func (f *Filter) Contains(key string, value interface{}) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, value)
	if len(res) != 1 {
		f.fail(f.t, "unexpected length of return values")
		return nil
	}

	filter, ok := res[0].(store.Filter)
	if !ok {
		f.fail(f.t, "unexpected type of return value")
		return nil
	}

	return filter
}

func (f *Filter) NotContains(key string, value interface{}) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, value)
	if len(res) != 1 {
		f.fail(f.t, "unexpected length of return values")
		return nil
	}

	filter, ok := res[0].(store.Filter)
	if !ok {
		f.fail(f.t, "unexpected type of return value")
		return nil
	}

	return filter
}

func (f *Filter) Eq(key string, value interface{}) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, value)
	if len(res) != 1 {
		f.fail(f.t, "unexpected length of return values")
		return nil
	}

	filter, ok := res[0].(store.Filter)
	if !ok {
		f.fail(f.t, "unexpected type of return value")
		return nil
	}

	return filter
}

func (f *Filter) Lt(key string, value interface{}) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, value)
	if len(res) != 1 {
		f.fail(f.t, "unexpected length of return values")
		return nil
	}

	filter, ok := res[0].(store.Filter)
	if !ok {
		f.fail(f.t, "unexpected type of return value")
		return nil
	}

	return filter
}

func (f *Filter) Gt(key string, value interface{}) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, value)
	if len(res) != 1 {
		f.fail(f.t, "unexpected length of return values")
		return nil
	}

	filter, ok := res[0].(store.Filter)
	if !ok {
		f.fail(f.t, "unexpected type of return value")
		return nil
	}

	return filter
}

func (f *Filter) NotEq(key string, value interface{}) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, value)
	if len(res) != 1 {
		f.fail(f.t, "unexpected length of return values")
		return nil
	}

	filter, ok := res[0].(store.Filter)
	if !ok {
		f.fail(f.t, "unexpected type of return value")
		return nil
	}

	return filter
}

func (f *Filter) Or(another store.Filter) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, another)
	if len(res) != 1 {
		f.fail(f.t, "unexpected length of return values")
		return nil
	}

	filter, ok := res[0].(store.Filter)
	if !ok {
		f.fail(f.t, "unexpected type of return value")
		return nil
	}

	return filter
}

func (f *Filter) And(another store.Filter) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, another)
	if len(res) != 1 {
		f.fail(f.t, "unexpected length of return values")
		return nil
	}

	filter, ok := res[0].(store.Filter)
	if !ok {
		f.fail(f.t, "unexpected type of return value")
		return nil
	}

	return filter
}

// NewFilter creates a mock store.Filter
func NewFilter(t *testing.T) *Filter {
	f := Filter{
		t: t,
	}

	if t != nil{
		f.fail = t.Fatal
	}

	return &f
}

// NewFilterWithFail creates a mock store.Filter with an expected failure
func NewFilterWithFail(t *testing.T, expect ...interface{}) *Filter {
	f := NewFilter(t)
	f.fail = func(v ...interface{}) {
		t.Helper()
		assert.Equal(t, append([]interface{}{t}, expect...), v)
	}

	return f
}
