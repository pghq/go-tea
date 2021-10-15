package mocking

import (
	"context"
	"testing"

	"github.com/pghq/go-museum/museum/internal"
	"github.com/pghq/go-museum/museum/store"
)

var (
	_ store.Add = NewAdd(nil)
)

func (c *Store) Add() store.Add {
	c.t.Helper()
	res := c.Call(c.t)
	if len(res) != 1{
		c.Fatalf(c.t, "length of return values for Add is not equal to 1")
	}

	add, ok := res[0].(store.Add)
	if !ok{
		c.Fatalf(c.t,"return value #1 of Add is not a store.Add")
	}

	return add
}

type Add struct {
	internal.Mock
	t   *testing.T
}

func (a *Add) Query(query store.Query) store.Add {
	a.t.Helper()
	res := a.Call(a.t, query)
	if len(res) != 1{
		a.Fatalf(a.t, "length of return values for Query is not equal to 1")
	}

	add, ok := res[0].(store.Add)
	if !ok{
		a.Fatalf(a.t,"return value #1 of Query is not a store.Add")
	}

	return add
}

func (a *Add) Execute(ctx context.Context) (int, error) {
	a.t.Helper()
	res := a.Call(a.t, ctx)
	if len(res) != 2{
		a.Fatalf(a.t, "length of return values for Execute is not equal to 2")
	}

	if res[1] != nil{
		err, ok := res[1].(error)
		if !ok{
			a.Fatalf(a.t,"return value #2 of Execute is not an error")
		}
		return 0, err
	}

	count, ok := res[0].(int)
	if !ok{
		a.Fatalf(a.t,"return value #1 of Execute is not an int")
	}

	return count, nil
}

func (a *Add) Statement() (string, []interface{}, error) {
	a.t.Helper()
	res := a.Call(a.t)
	if len(res) != 3{
		a.Fatalf(a.t, "length of return values for Statement is not equal to 3")
	}

	if res[2] != nil{
		err, ok := res[2].(error)
		if !ok{
			a.Fatalf(a.t,"return value #3 of Statement is not an error")
		}
		return "", nil, err
	}

	statement, ok := res[0].(string)
	if !ok{
		a.Fatalf(a.t,"return value #1 of Statement is not an string")
	}

	if res[1] != nil{
		args, ok := res[1].([]interface{})
		if !ok{
			a.Fatalf(a.t,"return value #2 of Statement is not an []interface{}")
		}
		return statement, args, nil
	}

	return statement, nil, nil
}

func (a *Add) To(collection string) store.Add {
	a.t.Helper()
	res := a.Call(a.t, collection)
	if len(res) != 1{
		a.Fatalf(a.t, "length of return values for To is not equal to 1")
	}

	add, ok := res[0].(store.Add)
	if !ok{
		a.Fatalf(a.t,"return value #1 of To is not a store.Add")
	}

	return add
}

func (a *Add) Item(value map[string]interface{}) store.Add {
	a.t.Helper()
	res := a.Call(a.t, value)
	if len(res) != 1{
		a.Fatalf(a.t, "length of return values for Item is not equal to 1")
	}

	add, ok := res[0].(store.Add)
	if !ok{
		a.Fatalf(a.t,"return value #1 of Item is not a store.Add")
	}

	return add
}

func NewAdd(t *testing.T) *Add {
	a := Add{t: t}

	return &a
}
