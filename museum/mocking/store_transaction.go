package mocking

import (
	"context"
	"testing"

	"github.com/pghq/go-museum/museum/internal"
	"github.com/pghq/go-museum/museum/store"
)

var (
	_ store.Transaction = NewTransaction(nil)
)

func (c *Store) Transaction(ctx context.Context) (store.Transaction, error) {
	c.t.Helper()
	res := c.Call(c.t, ctx)
	if len(res) != 2{
		c.Fatalf(c.t, "length of return values for Remove is not equal to 2")
	}

	if res[1] != nil{
		err, ok := res[1].(error)
		if !ok{
			c.Fatalf(c.t,"return value #2 of Transaction is not an error")
		}
		return nil, err
	}

	transaction, ok := res[0].(store.Transaction)
	if !ok{
		c.Fatalf(c.t,"return value #1 of Transaction is not a store.Transaction")
	}

	return transaction, nil
}

type Transaction struct {
	internal.Mock
	t *testing.T
}

func (tx *Transaction) Commit() error {
	tx.t.Helper()
	res := tx.Call(tx.t)
	if len(res) != 1{
		tx.Fatalf(tx.t, "length of return values for Commit is not equal to 1")
	}

	if res[0] != nil{
		err, ok := res[0].(error)
		if !ok{
			tx.Fatalf(tx.t,"return value #1 of Commit is not an error")
		}
		return err
	}

	return nil
}

func (tx *Transaction) Rollback() error {
	tx.t.Helper()
	res := tx.Call(tx.t)
	if len(res) != 1{
		tx.Fatalf(tx.t, "length of return values for Rollback is not equal to 1")
	}
	
	if res[0] != nil{
		err, ok := res[0].(error)
		if !ok{
			tx.Fatalf(tx.t,"return value #1 of Rollback is not an error")
		}
		return err
	}

	return nil
}

func (tx *Transaction) Execute(statement store.Encoder) (int, error) {
	tx.t.Helper()
	res := tx.Call(tx.t, statement)
	if len(res) != 2{
		tx.Fatalf(tx.t, "length of return values for Execute is not equal to 2")
	}

	if res[1] != nil{
		err, ok := res[1].(error)
		if !ok{
			tx.Fatalf(tx.t,"return value #2 of Execute is not an error")
		}
		return 0, err
	}

	count, ok := res[0].(int)
	if !ok{
		tx.Fatalf(tx.t,"return value #1 of Execute is not a int")
	}

	return count, nil
}

func NewTransaction(t *testing.T) *Transaction {
	tx := Transaction{t: t}

	return &tx
}
