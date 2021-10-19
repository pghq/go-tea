package pilot

import (
	"context"
	"testing"

	"github.com/pghq/go-museum/museum/internal"
	"github.com/pghq/go-museum/museum/store"
)

var (
	_ store.Query = NewQuery(nil)
)

func (c *Store) Query() store.Query {
	c.t.Helper()
	res := c.Call(c.t)
	if len(res) != 1 {
		c.Fatalf(c.t, "length of return values for Query is not equal to 1")
	}

	query, ok := res[0].(store.Query)
	if !ok {
		c.Fatalf(c.t, "return value #1 of Query is not a store.Query")
	}

	return query
}

type Query struct {
	internal.Mock
	t *testing.T
}

func (q *Query) Statement() (string, []interface{}, error) {
	q.t.Helper()
	res := q.Call(q.t)
	if len(res) != 3 {
		q.Fatalf(q.t, "length of return values for Statement is not equal to 3")
	}

	if res[2] != nil {
		err, ok := res[2].(error)
		if !ok {
			q.Fatalf(q.t, "return value #3 of Statement is not an error")
		}
		return "", nil, err
	}

	statement, ok := res[0].(string)
	if !ok {
		q.Fatalf(q.t, "return value #1 of Statement is not an string")
	}

	if res[1] != nil {
		args, ok := res[1].([]interface{})
		if !ok {
			q.Fatalf(q.t, "return value #2 of Statement is not an []interface{}")
		}
		return statement, args, nil
	}

	return statement, nil, nil
}

func (q *Query) Secondary() store.Query {
	q.t.Helper()
	res := q.Call(q.t)
	if len(res) != 1 {
		q.Fatalf(q.t, "length of return values for Secondary is not equal to 1")
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.Fatalf(q.t, "return value #1 of Secondary is not a store.Query")
	}

	return query
}

func (q *Query) From(collection string) store.Query {
	q.t.Helper()
	res := q.Call(q.t, collection)
	if len(res) != 1 {
		q.Fatalf(q.t, "length of return values for From is not equal to 1")
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.Fatalf(q.t, "return value #1 of From is not a store.Query")
	}

	return query
}

func (q *Query) And(collection string, args ...interface{}) store.Query {
	q.t.Helper()
	res := q.Call(q.t, append([]interface{}{collection}, args...)...)
	if len(res) != 1 {
		q.Fatalf(q.t, "length of return values for And is not equal to 1")
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.Fatalf(q.t, "return value #1 of And is not a store.Query")
	}

	return query
}

func (q *Query) Filter(filter store.Filter) store.Query {
	q.t.Helper()
	res := q.Call(q.t, filter)
	if len(res) != 1 {
		q.Fatalf(q.t, "length of return values for Filter is not equal to 1")
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.Fatalf(q.t, "return value #1 of Filter is not a store.Query")
	}

	return query
}

func (q *Query) Order(by string) store.Query {
	q.t.Helper()
	res := q.Call(q.t, by)
	if len(res) != 1 {
		q.Fatalf(q.t, "length of return values for Order is not equal to 1")
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.Fatalf(q.t, "return value #1 of Order is not a store.Query")
	}

	return query
}

func (q *Query) First(first int) store.Query {
	q.t.Helper()
	res := q.Call(q.t, first)
	if len(res) != 1 {
		q.Fatalf(q.t, "length of return values for First is not equal to 1")
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.Fatalf(q.t, "return value #1 of First is not a store.Query")
	}

	return query
}

func (q *Query) After(key string, value interface{}) store.Query {
	q.t.Helper()
	res := q.Call(q.t, key, value)
	if len(res) != 1 {
		q.Fatalf(q.t, "length of return values for After is not equal to 1")
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.Fatalf(q.t, "return value #1 of After is not a store.Query")
	}

	return query
}

func (q *Query) Return(key string, args ...interface{}) store.Query {
	q.t.Helper()
	res := q.Call(q.t, append([]interface{}{key}, args...)...)
	if len(res) != 1 {
		q.Fatalf(q.t, "length of return values for Return is not equal to 1")
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.Fatalf(q.t, "return value #1 of Return is not a store.Query")
	}

	return query
}

func (q *Query) Execute(ctx context.Context) (store.Cursor, error) {
	q.t.Helper()
	res := q.Call(q.t, ctx)
	if len(res) != 2 {
		q.Fatalf(q.t, "length of return values for Execute is not equal to 2")
	}

	if res[1] != nil {
		err, ok := res[1].(error)
		if !ok {
			q.Fatalf(q.t, "return value #2 of Execute is not an error")
		}
		return nil, err
	}

	cursor, ok := res[0].(store.Cursor)
	if !ok {
		q.Fatalf(q.t, "return value #1 of Execute is not a store.Cursor")
	}

	return cursor, nil
}

func NewQuery(t *testing.T) *Query {
	q := Query{t: t}

	return &q
}
