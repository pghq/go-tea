package test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/internal/database"
)

type DatabaseClient struct {
	t *testing.T
	err error
	database.Client
	connected struct{
		calls int
		value bool
	}
	transaction struct{
		calls int
		ctx context.Context
		value database.Transaction
	}
	query struct{
		calls int
		value database.Query
	}
	add struct{
		calls int
		value database.Add
	}
	remove struct{
		calls int
		value database.Remove
	}
}

func (c *DatabaseClient) IsConnected() bool{
	c.connected.calls -= 1
	return c.connected.value
}

func (c *DatabaseClient) ExpectIsConnected() *DatabaseClient{
	c.err = nil
	c.connected.calls += 1

	return c
}

func (c *DatabaseClient) ReturnIsConnected(value bool) *DatabaseClient{
	c.connected.value = value
	return c
}

func (c *DatabaseClient) Query() database.Query{
	c.query.calls -= 1

	return c.query.value
}

func (c *DatabaseClient) ExpectQuery() *DatabaseClient{
	c.err = nil
	c.query.calls += 1
	return c
}

func (c *DatabaseClient) ReturnQuery(value database.Query) *DatabaseClient{
	c.query.value = value
	return c
}

func (c *DatabaseClient) Add() database.Add{
	c.add.calls -= 1

	return c.add.value
}

func (c *DatabaseClient) ExpectAdd() *DatabaseClient{
	c.err = nil
	c.add.calls += 1
	return c
}

func (c *DatabaseClient) ReturnAdd(value database.Add) *DatabaseClient{
	c.add.value = value
	return c
}

func (c *DatabaseClient) Remove() database.Remove{
	c.remove.calls -= 1

	return c.remove.value
}

func (c *DatabaseClient) ExpectRemove() *DatabaseClient{
	c.err = nil
	c.remove.calls += 1
	return c
}

func (c *DatabaseClient) ReturnRemove(value database.Remove) *DatabaseClient{
	c.remove.value = value
	return c
}

func (c *DatabaseClient) Transaction(ctx context.Context) (database.Transaction, error){
	c.t.Helper()
	if c.err != nil{
		return nil, c.err
	}

	c.transaction.calls -= 1
	if c.transaction.ctx != nil{
		assert.Equal(c.t, c.transaction.ctx, ctx)
	}

	return c.transaction.value, nil
}

func (c *DatabaseClient) ExpectTransaction(ctx context.Context) *DatabaseClient{
	c.err = nil
	c.transaction.calls += 1
	c.transaction.ctx = ctx
	return c
}

func (c *DatabaseClient) ReturnTransaction(value database.Transaction) *DatabaseClient{
	c.transaction.value = value
	return c
}

func (c *DatabaseClient) Error(err error) *DatabaseClient{
	c.err = err
	return c
}

func (c *DatabaseClient) Assert(){
	c.t.Helper()
	if c.connected.calls != 0{
		c.t.Fatal("not enough calls to connected")
	}

	if c.query.calls != 0{
		c.t.Fatal("not enough calls to query")
	}

	if c.add.calls != 0{
		c.t.Fatal("not enough calls to add")
	}

	if c.remove.calls != 0{
		c.t.Fatal("not enough calls to remove")
	}

	if c.transaction.calls != 0{
		c.t.Fatal("not enough calls to transaction")
	}
}

func NewDatabaseClient(t *testing.T) *DatabaseClient{
	c := DatabaseClient{
		t: t,
	}

	return &c
}

type Query struct {
	t *testing.T
	err error
	decode struct{
		calls int
		to database.Query
	}
	execute struct{
		calls int
		ctx context.Context
	}
}

func (q *Query) Statement() (string, []interface{}, error) {
	panic("implement me")
}

func (q *Query) Secondary() database.Query {
	panic("implement me")
}

func (q *Query) From(_ string) database.Query {
	panic("implement me")
}

func (q *Query) And(_ string, _ ...interface{}) database.Query {
	panic("implement me")
}

func (q *Query) Filter(_ string, _ ...interface{}) database.Query {
	panic("implement me")
}

func (q *Query) Order(_ string) database.Query {
	panic("implement me")
}

func (q *Query) First(_ uint) database.Query {
	panic("implement me")
}

func (q *Query) After(_ string, _ interface{}) database.Query {
	panic("implement me")
}

func (q *Query) Return(_ string, _ ...interface{}) database.Query {
	panic("implement me")
}

func (q *Query) Decode(to database.Query) error {
	q.t.Helper()
	if q.err != nil{
		return q.err
	}

	q.decode.calls -= 1
	if q.decode.to != nil{
		assert.Equal(q.t, q.decode.to, to)
	}

	return nil
}

func (q *Query) ExpectDecode(to database.Query) *Query{
	q.err = nil
	q.decode.calls += 1
	q.decode.to = to
	return q
}

func (q *Query) Execute(ctx context.Context) (database.Cursor, error) {
	q.t.Helper()
	if q.err != nil{
		return nil, q.err
	}

	if q.execute.ctx != nil{
		assert.Equal(q.t, q.execute.ctx, ctx)
	}

	q.execute.calls -= 1
	return NewCursor(q.t), nil
}

func (q *Query) ExpectExecute(ctx context.Context) *Query{
	q.err = nil
	q.execute.calls += 1
	q.execute.ctx = ctx
	return q
}

func (q *Query) Error(err error) *Query{
	q.err = err
	return q
}

func (q *Query) Assert(){
	if q.decode.calls != 0{
		q.t.Fatal("not enough calls to decode")
	}

	if q.execute.calls != 0{
		q.t.Fatal("not enough calls to execute")
	}
}

func NewQuery(t *testing.T) *Query{
	q := Query{
		t: t,
	}

	return &q
}

type Add struct{
	t *testing.T
	err error
	to struct{
		calls int
		collection string
	}
	item struct{
		calls int
		value map[string]interface{}
	}
}

func (a *Add) Statement() (string, []interface{}, error) {
	panic("implement me")
}

func (a *Add) To(collection string) database.Add {
	a.t.Helper()
	a.to.calls -= 1
	if a.to.collection != ""{
		assert.Equal(a.t, a.to.collection, collection)
	}

	return a
}

func (a *Add) ExpectTo(collection string) *Add{
	a.err = nil
	a.to.calls += 1
	a.to.collection = collection
	return a
}

func (a *Add) Item(value map[string]interface{}) database.Add {
	a.t.Helper()
	a.item.calls -= 1
	if a.item.value != nil{
		assert.Equal(a.t, a.item.value, value)
	}

	return a
}

func (a *Add) ExpectItem(value map[string]interface{}) *Add{
	a.err = nil
	a.item.calls += 1
	a.item.value = value
	return a
}

func (a *Add) Query(_ database.Query) database.Add {
	panic("implement me")
}

func (a *Add) Decode(_ database.Add) error {
	panic("implement me")
}

func (a *Add) Execute(_ context.Context) (uint, error) {
	panic("implement me")
}

func (a *Add) Error(err error) *Add{
	a.err = err
	return a
}

func (a *Add) Assert(){
	a.t.Helper()
	if a.item.calls != 0{
		a.t.Fatal("not enough calls to item")
	}

	if a.to.calls != 0{
		a.t.Fatal("not enough calls to to")
	}
}

func NewAdd(t *testing.T) *Add{
	a := Add{
		t: t,
	}

	return &a
}

type Remove struct{
	t *testing.T
	err error
	decode struct{
		calls int
		to database.Remove
	}
	execute struct{
		calls int
		ctx context.Context
	}
}

func (r *Remove) Statement() (string, []interface{}, error) {
	panic("implement me")
}

func (r *Remove) From(_ string) database.Remove {
	panic("implement me")
}

func (r *Remove) Filter(_ string, _ ...interface{}) database.Remove {
	panic("implement me")
}

func (r *Remove) Order(_ string) database.Remove {
	panic("implement me")
}

func (r *Remove) First(_ uint) database.Remove {
	panic("implement me")
}

func (r *Remove) After(_ string, _ interface{}) database.Remove {
	panic("implement me")
}

func (r *Remove) Decode(to database.Remove) error {
	r.t.Helper()
	if r.err != nil{
		return r.err
	}

	r.decode.calls -= 1
	if r.decode.to != nil{
		assert.Equal(r.t, r.decode.to, to)
	}

	return nil
}

func (r *Remove) ExpectDecode(to database.Remove) *Remove{
	r.err = nil
	r.decode.calls += 1
	r.decode.to = to
	return r
}

func (r *Remove) Execute(ctx context.Context) (uint, error) {
	r.t.Helper()
	if r.err != nil{
		return 0, r.err
	}

	if r.execute.ctx != nil{
		assert.Equal(r.t, r.execute.ctx, ctx)
	}

	r.execute.calls -= 1
	return 0, nil
}

func (r *Remove) ExpectExecute(ctx context.Context) *Remove{
	r.err = nil
	r.execute.calls += 1
	r.execute.ctx = ctx
	return r
}

func (r *Remove) Error(err error) *Remove{
	r.err = err
	return r
}

func (r *Remove) Assert(){
	if r.decode.calls != 0{
		r.t.Fatal("not enough calls to decode")
	}

	if r.execute.calls != 0{
		r.t.Fatal("not enough calls to execute")
	}
}

func NewRemove(t *testing.T) *Remove{
	r := Remove{
		t: t,
	}

	return &r
}

type Transaction struct {
	t *testing.T
	err error
	execute struct{
		err error
		calls int
		statement database.Communicator
	}
	commit struct{
		err error
		calls int
	}
	rollback struct{
		calls int
	}
}

func (t *Transaction) Execute(statement database.Communicator) (uint, error) {
	t.t.Helper()
	if t.err != nil{
		return 0, t.err
	}

	if t.execute.err != nil{
		return 0, t.execute.err
	}

	if t.execute.statement != nil{
		assert.Equal(t.t, t.execute.statement, statement)
	}

	t.execute.calls -= 1
	return 0, nil
}

func (t *Transaction) ExpectExecute(statement database.Communicator) *Transaction{
	t.err = nil
	t.execute.calls += 1
	t.execute.statement = statement

	return t
}

func (t *Transaction) ReturnExecute(err error) *Transaction{
	t.execute.err = err

	return t
}

func (t *Transaction) Commit() error {
	if t.err != nil{
		return t.err
	}

	if t.commit.err != nil{
		return t.commit.err
	}

	t.commit.calls -= 1
	return nil
}

func (t *Transaction) ExpectCommit() *Transaction{
	t.err = nil
	t.commit.calls += 1

	return t
}

func (t *Transaction) ReturnCommit(err error) *Transaction{
	t.commit.err = err

	return t
}

func (t *Transaction) Rollback() error {
	if t.err != nil{
		return t.err
	}

	t.rollback.calls -= 1
	return nil
}

func (t *Transaction) ExpectRollback() *Transaction{
	t.err = nil
	t.rollback.calls += 1

	return t
}

func (t *Transaction) Error(err error) *Transaction{
	t.err = err
	return t
}

func (t *Transaction) Assert(){
	t.t.Helper()
	if t.commit.calls != 0{
		t.t.Fatal("not enough calls to commit")
	}

	if t.rollback.calls != 0{
		t.t.Fatal("not enough calls to rollback")
	}

	if t.execute.calls != 0{
		t.t.Fatal("not enough calls to execute")
	}
}

func NewTransaction(t *testing.T) *Transaction{
	tx := Transaction{
		t: t,
	}

	return &tx
}

type Item struct{
	t *testing.T
	err error
	validate struct{
		calls int
	}
	value struct{
		calls int
		value map[string]interface{}
	}
}

func (i *Item) Validate() error{
	i.validate.calls -= 1
	return i.err
}

func (i *Item) Map() map[string]interface{}{
	i.value.calls -= 1
	return i.value.value
}

func (i *Item) ExpectMap() *Item{
	i.err = nil
	i.value.calls += 1
	return i
}

func (i *Item) ReturnMap(value map[string]interface{}) *Item{
	i.value.value = value
	return i
}

func (i *Item) ExpectValidate() *Item{
	i.err = nil
	i.validate.calls += 1
	return i
}

func (i *Item) Error(err error) *Item{
	i.err = err
	return i
}

func (i *Item) Assert(){
	i.t.Helper()
	if i.validate.calls != 0{
		i.t.Fatal("not enough calls to validate")
	}

	if i.value.calls != 0{
		i.t.Fatal("not enough calls to map")
	}
}

func NewItem(t *testing.T) *Item{
	i := Item{
		t: t,
	}

	return &i
}

type Cursor struct {
	t * testing.T
	database.Cursor
}

func NewCursor(t *testing.T) *Cursor{
	c := Cursor{
		t: t,
	}

	return &c
}
