package mocking

import (
	"context"
	"testing"

	"github.com/pghq/go-museum/museum/internal"
	"github.com/pghq/go-museum/museum/store"
)

var (
	_ store.Update = NewUpdate(nil)
)

func (c *Store) Update() store.Update{
	c.t.Helper()
	res := c.Call(c.t)
	if len(res) != 1{
		c.Fatalf(c.t, "length of return values for Update is not equal to 1")
	}

	update, ok := res[0].(store.Update)
	if !ok{
		c.Fatalf(c.t,"return value #1 of Update is not a store.Update")
	}

	return update
}

type Update struct {
	internal.Mock
	t *testing.T
}

func (u *Update) Statement() (string, []interface{}, error) {
	u.t.Helper()
	res := u.Call(u.t)
	if len(res) != 3{
		u.Fatalf(u.t, "length of return values for Statement is not equal to 3")
	}

	if res[2] != nil{
		err, ok := res[2].(error)
		if !ok{
			u.Fatalf(u.t,"return value #3 of Statement is not an error")
		}
		return "", nil, err
	}

	statement, ok := res[0].(string)
	if !ok{
		u.Fatalf(u.t,"return value #1 of Statement is not an string")
	}

	if res[1] != nil{
		args, ok := res[1].([]interface{})
		if !ok{
			u.Fatalf(u.t,"return value #2 of Statement is not an []interface{}")
		}
		return statement, args, nil
	}

	return statement, nil, nil
}

func (u *Update) In(collection string) store.Update {
	u.t.Helper()
	res := u.Call(u.t, collection)
	if len(res) != 1{
		u.Fatalf(u.t, "length of return values for In is not equal to 1")
	}

	update, ok := res[0].(store.Update)
	if !ok{
		u.Fatalf(u.t,"return value #1 of In is not a store.Update")
	}

	return update
}

func (u *Update) Item(snapshot map[string]interface{}) store.Update {
	u.t.Helper()
	res := u.Call(u.t, snapshot)
	if len(res) != 1{
		u.Fatalf(u.t, "length of return values for Item is not equal to 1")
	}

	update, ok := res[0].(store.Update)
	if !ok{
		u.Fatalf(u.t,"return value #1 of Item is not a store.Update")
	}

	return update
}

func (u *Update) Filter(filter store.Filter) store.Update {
	u.t.Helper()
	res := u.Call(u.t, filter)
	if len(res) != 1{
		u.Fatalf(u.t, "length of return values for Filter is not equal to 1")
	}

	update, ok := res[0].(store.Update)
	if !ok{
		u.Fatalf(u.t,"return value #1 of Filter is not a store.Update")
	}

	return update
}

func (u *Update) Execute(ctx context.Context) (int, error) {
	u.t.Helper()
	res := u.Call(u.t, ctx)
	if len(res) != 2{
		u.Fatalf(u.t, "length of return values for Execute is not equal to 2")
	}

	if res[1] != nil{
		err, ok := res[1].(error)
		if !ok{
			u.Fatalf(u.t,"return value #2 of Execute is not an error")
		}
		return 0, err
	}

	count, ok := res[0].(int)
	if !ok{
		u.Fatalf(u.t,"return value #1 of Execute is not a int")
	}

	return count, nil
}

func NewUpdate(t *testing.T) *Update {
	u := Update{t: t}
	return &u
}
