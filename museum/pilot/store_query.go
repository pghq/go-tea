package pilot

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/internal"
	"github.com/pghq/go-museum/museum/store"
)

var (
	_ store.Query = NewQuery(nil)
)

func (s *Store) Query() store.Query {
	s.t.Helper()
	res := s.Call(s.t)
	if len(res) != 1 {
		s.fail(s.t, "unexpected length of return values")
		return nil
	}

	query, ok := res[0].(store.Query)
	if !ok {
		s.fail(s.t, "unexpected type of return value")
		return nil
	}

	return query
}

// Query is a mock store.Query
type Query struct {
	internal.Mock
	t *testing.T
	fail func(v ...interface{})
}

func (q *Query) Statement() (string, []interface{}, error) {
	q.t.Helper()
	res := q.Call(q.t)
	if len(res) != 3 {
		q.fail(q.t, "unexpected length of return values")
		return "", nil, nil
	}

	if res[2] != nil {
		err, ok := res[2].(error)
		if !ok {
			q.fail(q.t, "unexpected type of return value")
			return "", nil, nil
		}
		return "", nil, err
	}

	statement, ok := res[0].(string)
	if !ok {
		q.fail(q.t, "unexpected type of return value")
		return "", nil, nil
	}

	if res[1] != nil {
		args, ok := res[1].([]interface{})
		if !ok {
			q.fail(q.t, "unexpected type of return value")
			return "", nil, nil
		}
		return statement, args, nil
	}

	return statement, nil, nil
}

func (q *Query) Secondary() store.Query {
	q.t.Helper()
	res := q.Call(q.t)
	if len(res) != 1 {
		q.fail(q.t, "unexpected length of return values")
		return nil
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.fail(q.t, "unexpected type of return value")
		return nil
	}

	return query
}

func (q *Query) From(collection string) store.Query {
	q.t.Helper()
	res := q.Call(q.t, collection)
	if len(res) != 1 {
		q.fail(q.t, "unexpected length of return values")
		return nil
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.fail(q.t, "unexpected type of return value")
		return nil
	}

	return query
}

func (q *Query) And(collection string, args ...interface{}) store.Query {
	q.t.Helper()
	res := q.Call(q.t, append([]interface{}{collection}, args...)...)
	if len(res) != 1 {
		q.fail(q.t, "unexpected length of return values")
		return nil
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.fail(q.t, "unexpected type of return value")
		return nil
	}

	return query
}

func (q *Query) Filter(filter store.Filter) store.Query {
	q.t.Helper()
	res := q.Call(q.t, filter)
	if len(res) != 1 {
		q.fail(q.t, "unexpected length of return values")
		return nil
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.fail(q.t, "unexpected type of return value")
		return nil
	}

	return query
}

func (q *Query) Order(by string) store.Query {
	q.t.Helper()
	res := q.Call(q.t, by)
	if len(res) != 1 {
		q.fail(q.t, "unexpected length of return values")
		return nil
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.fail(q.t, "unexpected type of return value")
		return nil
	}

	return query
}

func (q *Query) First(first int) store.Query {
	q.t.Helper()
	res := q.Call(q.t, first)
	if len(res) != 1 {
		q.fail(q.t, "unexpected length of return values")
		return nil
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.fail(q.t, "unexpected type of return value")
		return nil
	}

	return query
}

func (q *Query) After(key string, value interface{}) store.Query {
	q.t.Helper()
	res := q.Call(q.t, key, value)
	if len(res) != 1 {
		q.fail(q.t, "unexpected length of return values")
		return nil
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.fail(q.t, "unexpected type of return value")
		return nil
	}

	return query
}

func (q *Query) Return(key string, args ...interface{}) store.Query {
	q.t.Helper()
	res := q.Call(q.t, append([]interface{}{key}, args...)...)
	if len(res) != 1 {
		q.fail(q.t, "unexpected length of return values")
		return nil
	}

	query, ok := res[0].(store.Query)
	if !ok {
		q.fail(q.t, "unexpected type of return value")
		return nil
	}

	return query
}

func (q *Query) Execute(ctx context.Context) (store.Cursor, error) {
	q.t.Helper()
	res := q.Call(q.t, ctx)
	if len(res) != 2 {
		q.fail(q.t, "unexpected length of return values")
		return nil, nil
	}

	if res[1] != nil {
		err, ok := res[1].(error)
		if !ok {
			q.fail(q.t, "unexpected type of return value")
			return nil, nil
		}
		return nil, err
	}

	cursor, ok := res[0].(store.Cursor)
	if !ok {
		q.fail(q.t, "unexpected type of return value")
		return nil, nil
	}

	return cursor, nil
}

// NewQuery creates a mock store.Query
func NewQuery(t *testing.T) *Query {
	q := Query{
		t: t,
	}

	if t != nil{
		q.fail = t.Fatal
	}

	return &q
}

// NewQueryWithFail creates a mock store.Query with an expected failure
func NewQueryWithFail(t *testing.T, expect ...interface{}) *Query {
	q := NewQuery(t)
	q.fail = func(v ...interface{}) {
		t.Helper()
		assert.Equal(t, append([]interface{}{t}, expect...), v)
	}

	return q
}
