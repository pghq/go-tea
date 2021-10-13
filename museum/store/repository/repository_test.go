package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/internal/test"
)

func TestNew(t *testing.T) {
	t.Run("NoClientError", func(t *testing.T) {
		_, err := New(nil)
		assert.NotNil(t, err)
	})

	t.Run("DisconnectedClientError", func(t *testing.T) {
		client := test.NewDatabaseClient(t).ExpectIsConnected()
		defer client.Assert()
		_, err := New(client)
		assert.NotNil(t, err)
	})

	t.Run("NoError", func(t *testing.T) {
		client := test.NewDatabaseClient(t).ExpectIsConnected().ReturnIsConnected(true)
		defer client.Assert()
		r, err := New(client)
		assert.Nil(t, err)
		assert.NotNil(t, r)
	})
}

func TestRepository_Add(t *testing.T) {
	t.Run("NoItems", func(t *testing.T) {
		client := test.NewDatabaseClient(t).ReturnIsConnected(true)
		r, _ := New(client)
		err := r.Add(context.TODO(), "tests")
		assert.Nil(t, err)
	})

	t.Run("TransactionError", func(t *testing.T) {
		client := test.NewDatabaseClient(t).ReturnIsConnected(true).Error(errors.New("an error has occurred"))
		r, _ := New(client)
		item := test.NewItem(t)
		err := r.Add(context.TODO(), "tests", item)
		assert.NotNil(t, err)
	})

	t.Run("ValidateError", func(t *testing.T) {
		tx := test.NewTransaction(t).ExpectRollback()
		defer tx.Assert()
		client := test.NewDatabaseClient(t).ReturnIsConnected(true).ReturnTransaction(tx)
		r, _ := New(client)
		item := test.NewItem(t).Error(errors.New("an error has occurred"))
		err := r.Add(context.TODO(), "tests", item)
		assert.NotNil(t, err)
	})

	t.Run("ExecuteError", func(t *testing.T) {
		tx := test.NewTransaction(t).ExpectRollback().ReturnExecute(errors.New("an error has occurred"))
		defer tx.Assert()
		value := map[string]interface{}{"coverage": 50}
		add := test.NewAdd(t).ExpectTo("tests").ExpectItem(value)
		defer add.Assert()
		client := test.NewDatabaseClient(t).ReturnIsConnected(true).ReturnTransaction(tx).ReturnAdd(add)
		r, _ := New(client)
		item := test.NewItem(t).ExpectValidate().ExpectMap().ReturnMap(value)
		defer item.Assert()
		err := r.Add(context.TODO(), "tests", item)
		assert.NotNil(t, err)
	})

	t.Run("CommitError", func(t *testing.T) {
		value := map[string]interface{}{"coverage": 50}
		add := test.NewAdd(t).ExpectTo("tests").ExpectItem(value)
		defer add.Assert()
		tx := test.NewTransaction(t).ExpectRollback().ExpectExecute(add.ExpectItem(value).ExpectTo("tests").To("tests").Item(value)).ReturnCommit(errors.New("an error has occurred"))
		defer tx.Assert()
		client := test.NewDatabaseClient(t).ReturnIsConnected(true).ReturnTransaction(tx).ReturnAdd(add)
		r, _ := New(client)
		item := test.NewItem(t).ExpectValidate().ExpectMap().ReturnMap(value)
		defer item.Assert()
		err := r.Add(context.TODO(), "tests", item)
		assert.NotNil(t, err)
	})

	t.Run("NoError", func(t *testing.T) {
		value := map[string]interface{}{"coverage": 50}
		add := test.NewAdd(t).ExpectTo("tests").ExpectItem(value)
		defer add.Assert()
		tx := test.NewTransaction(t).ExpectRollback().ExpectExecute(add.ExpectItem(value).ExpectTo("tests").To("tests").Item(value)).ExpectCommit()
		defer tx.Assert()
		client := test.NewDatabaseClient(t).ReturnIsConnected(true).ReturnTransaction(tx).ReturnAdd(add)
		r, _ := New(client)
		item := test.NewItem(t).ExpectValidate().ExpectMap().ReturnMap(value)
		defer item.Assert()
		err := r.Add(context.TODO(), "tests", item)
		assert.Nil(t, err)
	})
}

func TestRepository_Query(t *testing.T) {
	t.Run("DecodeError", func(t *testing.T) {
		q := test.NewQuery(t).Error(errors.New("an error has occurred"))
		client := test.NewDatabaseClient(t).ReturnIsConnected(true).ReturnQuery(q)
		r, _ := New(client)
		_, err := r.Query(context.TODO(), q)
		assert.NotNil(t, err)
	})

	t.Run("Execute", func(t *testing.T) {
		ctx := context.TODO()
		q := test.NewQuery(t)
		q.ExpectExecute(ctx).ExpectDecode(q)
		defer q.Assert()
		client := test.NewDatabaseClient(t).ReturnIsConnected(true).ReturnQuery(q)
		r, _ := New(client)
		c, _ := r.Query(context.TODO(), q)
		assert.NotNil(t, c)
	})
}

func TestRepository_Remove(t *testing.T) {
	t.Run("DecodeError", func(t *testing.T) {
		remove := test.NewRemove(t).Error(errors.New("an error has occurred"))
		client := test.NewDatabaseClient(t).ReturnIsConnected(true).ReturnRemove(remove)
		r, _ := New(client)
		_, err := r.Remove(context.TODO(), remove)
		assert.NotNil(t, err)
	})

	t.Run("Execute", func(t *testing.T) {
		ctx := context.TODO()
		remove := test.NewRemove(t)
		remove.ExpectExecute(ctx).ExpectDecode(remove)
		defer remove.Assert()
		client := test.NewDatabaseClient(t).ReturnIsConnected(true).ReturnRemove(remove)
		r, _ := New(client)
		c, _ := r.Remove(context.TODO(), remove)
		assert.NotNil(t, c)
	})
}
