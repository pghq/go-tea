package internal

import (
	"context"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// NoopHandler is a http handler that does nothing.
var NoopHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

type Mock struct {
	expected map[string][]*Expect
	lock     sync.Mutex
}

func (m *Mock) Expect(funcName string, args ...interface{}) *Expect {
	m.lock.Lock()
	defer m.lock.Unlock()

	if m.expected == nil {
		m.expected = make(map[string][]*Expect)
	}

	expect := &Expect{funcName: funcName, input: args}
	m.expected[funcName] = append(m.expected[funcName], expect)

	return expect
}

func (m *Mock) Assert(t *testing.T) {
	m.lock.Lock()
	defer m.lock.Unlock()

	t.Helper()
	for k, _ := range m.expected {
		expectations, ok := m.expected[k]
		if ok && len(expectations) > 0 {
			m.Fatalf(t, "not enough calls to %s, %d missing", k, len(expectations))
			return
		}
	}
}

func (m *Mock) Fatal(t *testing.T, args ...interface{}) {
	t.Helper()
	t.Fatal(args...)
}

func (m *Mock) Fatalf(t *testing.T, format string, args ...interface{}) {
	t.Helper()
	t.Fatalf(format, args...)
}

func (m *Mock) Call(t *testing.T, args ...interface{}) []interface{} {
	m.lock.Lock()
	defer m.lock.Unlock()

	t.Helper()
	pc, _, _, ok := runtime.Caller(1)
	caller := runtime.FuncForPC(pc)
	if ok && caller != nil {
		path := strings.Split(caller.Name()[5:], ".")
		funcName := path[len(path)-1]
		expectations, ok := m.expected[funcName]
		if !ok || len(expectations) == 0 {
			m.Fatalf(t, "unexpected call to %s", funcName)
		}

		expect, expects := expectations[0], expectations[1:]
		if len(expect.input) != len(args) {
			m.Fatalf(t, "arguments %+v to call %s != %+v", args, funcName, expect.input)
		}

		for i, arg := range expect.input {
			if _, ok := arg.(context.Context); ok {
				assert.Implements(t, (*context.Context)(nil), args[i])
			} else {
				assert.Equal(t, arg, args[i])
			}
		}

		m.expected[funcName] = expects
		return expect.output
	}

	m.Fatal(t, "could not obtain caller information")
	return nil
}

type Expect struct {
	funcName string
	input    []interface{}
	output   []interface{}
}

func (e *Expect) Return(output ...interface{}) {
	e.output = output
}
