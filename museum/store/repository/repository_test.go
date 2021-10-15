package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/mocking"
)

func TestRepository(t *testing.T) {
	t.Run("raises error on undefined client", func(t *testing.T) {
		_, err := New(nil)
		assert.NotNil(t, err)
	})

	t.Run("raises error on disconnected client", func(t *testing.T) {
		client := mocking.NewDisconnectedStore(t)
		client.Expect("IsConnected").Return(false)
		defer client.Assert(t)

		_, err := New(client)
		assert.NotNil(t, err)
	})

	t.Run("can create instance", func(t *testing.T) {
		client := mocking.NewStore(t)
		defer client.Assert(t)

		r, err := New(client)
		assert.Nil(t, err)
		assert.NotNil(t, r)
	})
}

func TestRepository_Add(t *testing.T) {
	t.Run("ignores undefined items", func(t *testing.T) {
		client := mocking.NewStore(t)
		defer client.Assert(t)

		r, _ := New(client)
		err := r.Add(context.TODO(), "tests")
		assert.Nil(t, err)
	})

	t.Run("raises transaction errors", func(t *testing.T) {
		client := mocking.NewStore(t)
		client.Expect("Transaction", context.TODO()).
			Return(nil, errors.New("an error has occurred"))
		defer client.Assert(t)

		r, _ := New(client)

		item := mocking.NewSnapper(t)
		defer item.Assert(t)

		err := r.Add(context.TODO(), "tests", item)
		assert.NotNil(t, err)
	})

	t.Run("raises execution errors", func(t *testing.T) {
		expect := mocking.NewAdd(t)
		expect.Expect("To", "tests").
			Return(expect)
		expect.Expect("Item", map[string]interface{}{"key": 1337}).
			Return(expect)

		transaction := mocking.NewTransaction(t)
		transaction.Expect("Execute", expect.To("tests").Item(map[string]interface{}{"key": 1337})).
			Return(nil, errors.New("an error has occurred"))
		transaction.Expect("Rollback").
			Return(nil)
		defer transaction.Assert(t)

		item := mocking.NewSnapper(t)
		item.Expect("Snapshot").
			Return(map[string]interface{}{"key": 1337})
		defer item.Assert(t)

		add := mocking.NewAdd(t)
		add.Expect("To", "tests").
			Return(add)
		add.Expect("Item", map[string]interface{}{"key": 1337}).
			Return(add)
		defer add.Assert(t)

		client := mocking.NewStore(t)
		client.Expect("Transaction", context.TODO()).
			Return(transaction, nil)
		client.Expect("Add").
			Return(add)
		defer client.Assert(t)

		r, _ := New(client)

		err := r.Add(context.TODO(), "tests", item)
		assert.NotNil(t, err)
	})

	t.Run("raises commit errors", func(t *testing.T) {
		expect := mocking.NewAdd(t)
		expect.Expect("To", "tests").
			Return(expect)
		expect.Expect("Item", map[string]interface{}{"key": 1337}).
			Return(expect)

		transaction := mocking.NewTransaction(t)
		transaction.Expect("Execute", expect.To("tests").Item(map[string]interface{}{"key": 1337})).
			Return(0, nil)
		transaction.Expect("Commit").
			Return(errors.New("an error has occurred"))
		transaction.Expect("Rollback").
			Return(nil)
		defer transaction.Assert(t)

		add := mocking.NewAdd(t)
		add.Expect("To", "tests").
			Return(add)
		add.Expect("Item", map[string]interface{}{"key": 1337}).
			Return(add)
		defer add.Assert(t)

		item := mocking.NewSnapper(t)
		item.Expect("Snapshot").
			Return(map[string]interface{}{"key": 1337})
		defer item.Assert(t)

		client := mocking.NewStore(t)
		client.Expect("Transaction", context.TODO()).
			Return(transaction, nil)
		client.Expect("Add").
			Return(add)
		defer client.Assert(t)

		r, _ := New(client)

		err := r.Add(context.TODO(), "tests", item)
		assert.NotNil(t, err)
	})

	t.Run("can add", func(t *testing.T) {
		expect := mocking.NewAdd(t)
		expect.Expect("To", "tests").
			Return(expect)
		expect.Expect("Item", map[string]interface{}{"key": 1337}).
			Return(expect)

		transaction := mocking.NewTransaction(t)
		transaction.Expect("Execute", expect.To("tests").Item(map[string]interface{}{"key": 1337})).
			Return(0, nil)
		transaction.Expect("Commit").
			Return(nil)
		transaction.Expect("Rollback").
			Return(nil)
		defer transaction.Assert(t)

		add := mocking.NewAdd(t)
		add.Expect("To", "tests").
			Return(add)
		add.Expect("Item", map[string]interface{}{"key": 1337}).
			Return(add)
		defer add.Assert(t)

		item := mocking.NewSnapper(t)
		item.Expect("Snapshot").
			Return(map[string]interface{}{"key": 1337})
		defer item.Assert(t)

		client := mocking.NewStore(t)
		client.Expect("Transaction", context.TODO()).
			Return(transaction, nil)
		client.Expect("Add").
			Return(add)
		defer client.Assert(t)

		r, _ := New(client)

		err := r.Add(context.TODO(), "tests", item)
		assert.Nil(t, err)
	})
}

func TestRepository_Search(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		client := mocking.NewStore(t)
		client.Expect("Query").
			Return(mocking.NewQuery(t))
		defer client.Assert(t)

		r, _ := New(client)
		assert.NotNil(t, r.Query())
	})

	t.Run("can execute", func(t *testing.T) {
		query := mocking.NewQuery(t)
		query.Expect("Execute", context.TODO()).
			Return(mocking.NewCursor(t), nil)
		defer query.Assert(t)

		client := mocking.NewStore(t)
		defer client.Assert(t)

		r, _ := New(client)
		c, _ := r.Search(context.TODO(), query)
		assert.NotNil(t, c)
	})
}

func TestRepository_Remove(t *testing.T) {
	t.Run("can execute", func(t *testing.T) {
		remove := mocking.NewRemove(t)
		remove.Expect("From", "tests").
			Return(remove)
		remove.Expect("Filter", nil).
			Return(remove)
		remove.Expect("First", 1).
			Return(remove)
		remove.Expect("Execute", context.TODO()).
			Return(0, nil)
		defer remove.Assert(t)

		client := mocking.NewStore(t)
		client.Expect("Remove").
			Return(remove)
		defer client.Assert(t)

		r, _ := New(client)
		_, err := r.Remove(context.TODO(), "tests", nil, 1)
		assert.Nil(t, err)
	})
}

func TestRepository_Filter(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		client := mocking.NewStore(t)
		client.Expect("Filter").
			Return(mocking.NewFilter(t))
		defer client.Assert(t)

		r, _ := New(client)
		assert.NotNil(t, r.Filter())
	})
}

func TestRepository_Update(t *testing.T) {
	t.Run("can execute", func(t *testing.T) {
		item := mocking.NewSnapper(t)
		item.Expect("Snapshot").
			Return(map[string]interface{}{"key": 1337})
		defer item.Assert(t)

		update := mocking.NewUpdate(t)
		update.Expect("In", "tests").
			Return(update)
		update.Expect("Filter", nil).
			Return(update)
		update.Expect("Item", map[string]interface {}{"key":1337}).
			Return(update)
		update.Expect("Execute", context.TODO()).
			Return(0, nil)
		defer update.Assert(t)

		client := mocking.NewStore(t)
		client.Expect("Update").
			Return(update)
		defer client.Assert(t)

		r, _ := New(client)
		_, err := r.Update(context.TODO(), "tests", nil, item)
		assert.Nil(t, err)
	})
}
