package mocking

import (
	"testing"

	"github.com/pghq/go-museum/museum/internal"
	"github.com/pghq/go-museum/museum/store"
)

var (
	_ store.Filter = NewFilter(nil)
)

func (c *Store) Filter() store.Filter{
	c.t.Helper()
	res := c.Call(c.t)
	if len(res) != 1{
		c.Fatalf(c.t, "length of return values for Filter is not equal to 1")
	}

	filter, ok := res[0].(store.Filter)
	if !ok{
		c.Fatalf(c.t,"return value #1 of Filter is not a store.Filter")
	}

	return filter
}

type Filter struct {
	internal.Mock
	t *testing.T
}

func (f *Filter) BeginsWith(key string, prefix string) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, prefix)
	if len(res) != 1{
		f.Fatalf(f.t, "length of return values for BeginsWith is not equal to 1")
	}

	filter, ok := res[0].(store.Filter)
	if !ok{
		f.Fatalf(f.t,"return value #1 of BeginsWith is not a store.Filter")
	}

	return filter
}

func (f *Filter) EndsWith(key string, suffix string) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, suffix)
	if len(res) != 1{
		f.Fatalf(f.t, "length of return values for EndsWith is not equal to 1")
	}

	filter, ok := res[0].(store.Filter)
	if !ok{
		f.Fatalf(f.t,"return value #1 of EndsWith is not a store.Filter")
	}

	return filter
}

func (f *Filter) Contains(key string, value interface{}) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, value)
	if len(res) != 1{
		f.Fatalf(f.t, "length of return values for Contains is not equal to 1")
	}

	filter, ok := res[0].(store.Filter)
	if !ok{
		f.Fatalf(f.t,"return value #1 of Contains is not a store.Filter")
	}

	return filter
}

func (f *Filter) NotContains(key string, value interface{}) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, value)
	if len(res) != 1{
		f.Fatalf(f.t, "length of return values for NotContains is not equal to 1")
	}

	filter, ok := res[0].(store.Filter)
	if !ok{
		f.Fatalf(f.t,"return value #1 of NotContains is not a store.Filter")
	}

	return filter
}

func (f *Filter) Eq(key string, value interface{}) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, value)
	if len(res) != 1{
		f.Fatalf(f.t, "length of return values for Eq is not equal to 1")
	}

	filter, ok := res[0].(store.Filter)
	if !ok{
		f.Fatalf(f.t,"return value #1 of Eq is not a store.Filter")
	}

	return filter
}

func (f *Filter) Lt(key string, value interface{}) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, value)
	if len(res) != 1{
		f.Fatalf(f.t, "length of return values for Lt is not equal to 1")
	}

	filter, ok := res[0].(store.Filter)
	if !ok{
		f.Fatalf(f.t,"return value #1 of Lt is not a store.Filter")
	}

	return filter
}

func (f *Filter) Gt(key string, value interface{}) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, value)
	if len(res) != 1{
		f.Fatalf(f.t, "length of return values for Gt is not equal to 1")
	}

	filter, ok := res[0].(store.Filter)
	if !ok{
		f.Fatalf(f.t,"return value #1 of Gt is not a store.Filter")
	}

	return filter
}

func (f *Filter) NotEq(key string, value interface{}) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, key, value)
	if len(res) != 1{
		f.Fatalf(f.t, "length of return values for NotEq is not equal to 1")
	}

	filter, ok := res[0].(store.Filter)
	if !ok{
		f.Fatalf(f.t,"return value #1 of NotEq is not a store.Filter")
	}

	return filter
}

func (f *Filter) Or(another store.Filter) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, another)
	if len(res) != 1{
		f.Fatalf(f.t, "length of return values for Or is not equal to 1")
	}

	filter, ok := res[0].(store.Filter)
	if !ok{
		f.Fatalf(f.t,"return value #1 of Or is not a store.Filter")
	}

	return filter
}

func (f *Filter) And(another store.Filter) store.Filter {
	f.t.Helper()
	res := f.Call(f.t, another)
	if len(res) != 1{
		f.Fatalf(f.t, "length of return values for And is not equal to 1")
	}

	filter, ok := res[0].(store.Filter)
	if !ok{
		f.Fatalf(f.t,"return value #1 of And is not a store.Filter")
	}

	return filter
}

func NewFilter(t *testing.T) *Filter{
	f := Filter{t: t}
	return &f
}
