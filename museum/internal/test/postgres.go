package test

import (
	"context"
	"testing"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
)

type PostgresPool struct {
	t     *testing.T
	err   error
	query struct {
		ctx   context.Context
		sql   string
		args  []interface{}
		calls int
	}
	exec struct {
		ctx   context.Context
		sql   string
		args  []interface{}
		calls int
	}
	begin struct {
		ctx   context.Context
		calls int
		value pgx.Tx
	}
}

func (p *PostgresPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	p.t.Helper()
	if p.err != nil {
		return nil, p.err
	}

	p.query.calls -= 1
	if p.query.ctx != nil {
		assert.Equal(p.t, p.query.ctx, ctx)
	}

	if p.query.sql != "" {
		assert.Equal(p.t, p.query.sql, sql)
	}

	if p.query.args != nil {
		assert.Equal(p.t, p.query.args, args)
	}

	return nil, nil
}

func (p *PostgresPool) ExpectQuery(ctx context.Context, sql string, args ...interface{}) *PostgresPool {
	p.err = nil
	p.query.ctx = ctx
	p.query.sql = sql
	p.query.args = args
	p.query.calls += 1
	return p
}

func (p *PostgresPool) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	p.t.Helper()
	if p.err != nil {
		return nil, p.err
	}

	p.exec.calls -= 1
	if p.exec.ctx != nil {
		assert.Equal(p.t, p.exec.ctx, ctx)
	}

	if p.exec.sql != "" {
		assert.Equal(p.t, p.exec.sql, sql)
	}

	if p.exec.args != nil {
		assert.Equal(p.t, p.exec.args, args)
	}

	return nil, nil
}

func (p *PostgresPool) ExpectExec(ctx context.Context, sql string, args ...interface{}) *PostgresPool {
	p.err = nil
	p.exec.ctx = ctx
	p.exec.sql = sql
	p.exec.args = args
	p.exec.calls += 1
	return p
}

func (p *PostgresPool) Begin(ctx context.Context) (pgx.Tx, error) {
	p.t.Helper()
	if p.err != nil {
		return nil, p.err
	}

	p.begin.calls -= 1
	if p.begin.ctx != nil {
		assert.Equal(p.t, p.begin.ctx, ctx)
	}

	return p.begin.value, nil
}

func (p *PostgresPool) ExpectBegin(ctx context.Context) *PostgresPool {
	p.err = nil
	p.begin.ctx = ctx
	p.begin.calls += 1
	return p
}

func (p *PostgresPool) ReturnBegin(value pgx.Tx) *PostgresPool {
	p.begin.value = value

	return p
}

func (p *PostgresPool) Error(err error) *PostgresPool {
	p.err = err
	return p
}

func (p *PostgresPool) Assert() {
	p.t.Helper()
	if p.exec.calls != 0 {
		p.t.Fatal("not enough exec calls")
	}

	if p.query.calls != 0 {
		p.t.Fatal("not enough query calls")
	}

	if p.begin.calls != 0 {
		p.t.Fatal("not enough begin calls")
	}
}

func NewPostgresPool(t *testing.T) *PostgresPool {
	p := PostgresPool{
		t: t,
	}
	return &p
}

type PostgresRows struct {
	t *testing.T
	pgx.Rows
	err  error
	scan struct {
		dest  []interface{}
		calls int
	}
	next struct {
		calls int
	}
	close struct {
		calls int
	}
}

func (r *PostgresRows) Close() {
	r.close.calls -= 1
}

func (r *PostgresRows) ExpectClose() *PostgresRows {
	r.err = nil
	r.close.calls += 1
	return r
}

func (r *PostgresRows) Err() error {
	if r.err != nil {
		return r.err
	}

	return nil
}

func (r *PostgresRows) Next() bool {
	r.next.calls -= 1
	return false
}

func (r *PostgresRows) ExpectNext() *PostgresRows {
	r.err = nil
	r.next.calls += 1
	return r
}

func (r *PostgresRows) Scan(dest ...interface{}) error {
	r.t.Helper()
	if r.err != nil {
		return r.err
	}

	r.scan.calls -= 1
	if r.scan.dest != nil {
		assert.Equal(r.t, r.scan.dest, dest)
	}

	return nil
}

func (r *PostgresRows) ExpectScan(dest ...interface{}) *PostgresRows {
	r.err = nil
	r.scan.dest = dest
	r.scan.calls += 1
	return r
}

func (r *PostgresRows) Error(err error) *PostgresRows {
	r.err = err
	return r
}

func (r *PostgresRows) Assert() {
	r.t.Helper()
	if r.next.calls != 0 {
		r.t.Fatal("not enough next calls")
	}

	if r.scan.calls != 0 {
		r.t.Fatal("not enough scan calls")
	}

	if r.close.calls != 0 {
		r.t.Fatal("not enough close calls")
	}
}

func NewPostgresRows(t *testing.T) *PostgresRows {
	rows := PostgresRows{
		t: t,
	}

	return &rows
}

type PostgresTx struct {
	pgx.Tx
	t        *testing.T
	err      error
	rollback struct {
		ctx   context.Context
		calls int
	}
	commit struct {
		ctx   context.Context
		calls int
	}
	exec struct {
		ctx   context.Context
		sql   string
		args  []interface{}
		calls int
	}
}

func (t *PostgresTx) Commit(ctx context.Context) error {
	t.t.Helper()
	if t.err != nil {
		return t.err
	}

	t.commit.calls -= 1
	if t.commit.ctx != nil {
		assert.Equal(t.t, t.commit.ctx, ctx)
	}

	return nil
}

func (t *PostgresTx) ExpectCommit(ctx context.Context) *PostgresTx {
	t.err = nil
	t.commit.ctx = ctx
	t.commit.calls += 1
	return t
}

func (t *PostgresTx) Rollback(ctx context.Context) error {
	t.t.Helper()
	if t.err != nil {
		return t.err
	}

	t.rollback.calls -= 1

	if t.rollback.ctx != nil {
		assert.Equal(t.t, t.rollback.ctx, ctx)
	}

	return nil
}

func (t *PostgresTx) ExpectRollback(ctx context.Context) *PostgresTx {
	t.err = nil
	t.rollback.ctx = ctx
	t.rollback.calls += 1
	return t
}

func (t *PostgresTx) Exec(ctx context.Context, sql string, args ...interface{}) (commandTag pgconn.CommandTag, err error) {
	t.t.Helper()
	if t.err != nil {
		return nil, t.err
	}

	t.exec.calls -= 1

	if t.exec.ctx != nil {
		assert.Equal(t.t, t.exec.ctx, ctx)
	}

	if t.exec.sql != "" {
		assert.Equal(t.t, t.exec.sql, sql)
	}

	if t.exec.args != nil {
		assert.Equal(t.t, t.exec.args, args)
	}

	return nil, nil
}

func (t *PostgresTx) ExpectExec(ctx context.Context, sql string, args ...interface{}) *PostgresTx {
	t.err = nil
	t.exec.ctx = ctx
	t.exec.sql = sql
	t.exec.args = args
	t.exec.calls += 1
	return t
}

func (t *PostgresTx) Error(err error) *PostgresTx {
	t.err = err
	return t
}

func (t *PostgresTx) Assert() {
	t.t.Helper()
	if t.exec.calls != 0 {
		t.t.Fatal("not enough exec calls")
	}

	if t.commit.calls != 0 {
		t.t.Fatal("not enough commit calls")
	}

	if t.rollback.calls != 0 {
		t.t.Fatal("not enough rollback calls")
	}
}

func NewPostgresTx(t *testing.T) *PostgresTx {
	tx := PostgresTx{t: t}

	return &tx
}
