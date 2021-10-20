package pilot

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
)

func TestStore(t *testing.T) {
	t.Run("can create disconnected store instance", func(t *testing.T) {
		store := NewDisconnectedStoreWithFail(t)
		assert.NotNil(t, store)
		defer store.Assert(t)

		store.Expect("IsConnected").
			Return(false)
		assert.False(t, store.IsConnected())
	})

	t.Run("raises bad call errors on store", func(t *testing.T) {
		store := NewDisconnectedStoreWithFail(t, "unexpected length of return values")
		defer store.Assert(t)

		store.Expect("IsConnected").
			Return(true, false)
		store.IsConnected()
	})

	t.Run("raises return value errors on store", func(t *testing.T) {
		store := NewDisconnectedStoreWithFail(t, "unexpected type of return value")
		defer store.Assert(t)

		store.Expect("IsConnected").
			Return("false")
		store.IsConnected()
	})

	t.Run("can create snapper instance", func(t *testing.T) {
		snapper := NewSnapperWithFail(t)
		assert.NotNil(t, snapper)
		defer snapper.Assert(t)

		snapper.Expect("Snapshot").
			Return(map[string]interface{}{})
		assert.Equal(t, map[string]interface{}{}, snapper.Snapshot())
	})

	t.Run("raises bad call errors on snapper", func(t *testing.T) {
		snapper := NewSnapperWithFail(t, "unexpected length of return values")
		defer snapper.Assert(t)

		snapper.Expect("Snapshot").
			Return(true, false)
		snapper.Snapshot()
	})

	t.Run("raises return value errors on snapper", func(t *testing.T) {
		snapper := NewSnapperWithFail(t, "unexpected type of return value")
		defer snapper.Assert(t)

		snapper.Expect("Snapshot").
			Return("false")
		snapper.Snapshot()
	})

	t.Run("can create cursor instance", func(t *testing.T) {
		cursor := NewCursorWithFail(t)
		assert.NotNil(t, cursor)
		defer cursor.Assert(t)

		cursor.Expect("Next").
			Return(true)
		assert.True(t, cursor.Next())

		cursor.Expect("Decode", "test").
			Return(nil)
		assert.Nil(t, cursor.Decode("test"))

		cursor.Expect("Decode").
			Return(errors.New("an error has occurred"))
		assert.NotNil(t, cursor.Decode())

		cursor.Expect("Close")
		cursor.Close()

		cursor.Expect("Error").
			Return(nil)
		assert.Nil(t, cursor.Error())

		cursor.Expect("Error").
			Return(errors.New("an error has occurred"))
		assert.NotNil(t, cursor.Error())
	})

	t.Run("raises bad calls errors on cursor", func(t *testing.T) {
		cursor := NewCursorWithFail(t, "unexpected length of return values")
		defer cursor.Assert(t)

		cursor.Expect("Next").
			Return(true, false)
		cursor.Next()

		cursor.Expect("Decode").
			Return(nil, nil)
		cursor.Decode()

		cursor.Expect("Close").
			Return(nil)
		cursor.Close()

		cursor.Expect("Error").
			Return(nil, nil)
		cursor.Error()
	})

	t.Run("raises bad return value errors on cursor", func(t *testing.T) {
		cursor := NewCursorWithFail(t, "unexpected type of return value")
		defer cursor.Assert(t)

		cursor.Expect("Next").
			Return("true")
		cursor.Next()

		cursor.Expect("Decode").
			Return("")
		cursor.Decode()

		cursor.Expect("Error").
			Return("")
		cursor.Error()
	})

	t.Run("can create add instance", func(t *testing.T) {
		add := NewAddWithFail(t)
		assert.NotNil(t, add)
		defer add.Assert(t)

		query := NewQuery(t)
		add.Expect("Query", query).
			Return(add)
		assert.Equal(t, add, add.Query(query))

		add.Expect("Execute", context.TODO()).
			Return(1, nil)
		count, err := add.Execute(context.TODO())
		assert.Equal(t, 1, count)
		assert.Nil(t, err)

		add.Expect("Execute", context.TODO()).
			Return(0, errors.New("an error has occurred"))
		count, err = add.Execute(context.TODO())
		assert.Equal(t, 0, count)
		assert.NotNil(t, err)

		add.Expect("Statement").
			Return("test", []interface{}{"test"}, nil)
		statement, args, err := add.Statement()
		assert.Equal(t, "test", statement)
		assert.Equal(t, []interface{}{"test"}, args)
		assert.Nil(t, err)

		add.Expect("Statement").
			Return("", nil, errors.New("an error has occurred"))
		statement, args, err = add.Statement()
		assert.NotNil(t, err)

		add.Expect("Statement").
			Return("test", nil, nil)
		statement, args, err = add.Statement()
		assert.Equal(t, "test", statement)

		add.Expect("To", "tests").
			Return(add)
		assert.Equal(t, add, add.To("tests"))

		add.Expect("Item", map[string]interface{}{"count": 1}).
			Return(add)
		assert.Equal(t, add, add.Item(map[string]interface{}{"count": 1}))

		store := NewDisconnectedStoreWithFail(t)
		defer store.Assert(t)

		store.Expect("Add").
			Return(add)
		assert.Equal(t, add, store.Add())
	})

	t.Run("raises bad calls errors on add", func(t *testing.T) {
		add := NewAddWithFail(t, "unexpected length of return values")
		defer add.Assert(t)

		query := NewQuery(t)
		add.Expect("Query", query).
			Return(add, add)
		add.Query(query)

		add.Expect("Execute", context.TODO()).
			Return(1, nil, nil)
		add.Execute(context.TODO())

		add.Expect("Statement").
			Return("test", []interface{}{"test"}, nil, nil)
		add.Statement()

		add.Expect("To", "tests").
			Return(add, add)
		add.To("tests")

		add.Expect("Item", map[string]interface{}{"count": 1}).
			Return(add, add)
		add.Item(map[string]interface{}{"count": 1})

		store := NewDisconnectedStoreWithFail(t, "unexpected length of return values")
		defer store.Assert(t)

		store.Expect("Add").
			Return(add, add)
		store.Add()
	})

	t.Run("raises return value errors on add", func(t *testing.T) {
		add := NewAddWithFail(t, "unexpected type of return value")
		defer add.Assert(t)

		query := NewQuery(t)
		add.Expect("Query", query).
			Return("")
		add.Query(query)

		add.Expect("Execute", context.TODO()).
			Return("", nil)
		add.Execute(context.TODO())

		add.Expect("Execute", context.TODO()).
			Return(0, "")
		add.Execute(context.TODO())

		add.Expect("Statement").
			Return(0, []interface{}{"test"}, nil)
		add.Statement()

		add.Expect("Statement").
			Return("", 0, nil)
		add.Statement()

		add.Expect("Statement").
			Return(0, []interface{}{"test"}, "")
		add.Statement()

		add.Expect("To", "tests").
			Return("")
		add.To("tests")

		add.Expect("Item", map[string]interface{}{"count": 1}).
			Return("")
		add.Item(map[string]interface{}{"count": 1})

		store := NewDisconnectedStoreWithFail(t, "unexpected type of return value")
		defer store.Assert(t)

		store.Expect("Add").
			Return("")
		store.Add()
	})

	t.Run("can create filter instance", func(t *testing.T) {
		filter := NewFilterWithFail(t)
		assert.NotNil(t, filter)
		defer filter.Assert(t)

		filter.Expect("BeginsWith", "test", "value").
			Return(filter)
		assert.Equal(t, filter, filter.BeginsWith("test", "value"))

		filter.Expect("EndsWith", "test", "value").
			Return(filter)
		assert.Equal(t, filter, filter.EndsWith("test", "value"))

		filter.Expect("Contains", "test", "value").
			Return(filter)
		assert.Equal(t, filter, filter.Contains("test", "value"))

		filter.Expect("NotContains", "test", "value").
			Return(filter)
		assert.Equal(t, filter, filter.NotContains("test", "value"))

		filter.Expect("Eq", "test", "value").
			Return(filter)
		assert.Equal(t, filter, filter.Eq("test", "value"))

		filter.Expect("Lt", "test", "value").
			Return(filter)
		assert.Equal(t, filter, filter.Lt("test", "value"))

		filter.Expect("Gt", "test", "value").
			Return(filter)
		assert.Equal(t, filter, filter.Gt("test", "value"))

		filter.Expect("NotEq", "test", "value").
			Return(filter)
		assert.Equal(t, filter, filter.NotEq("test", "value"))

		filter.Expect("Or", filter).
			Return(filter)
		assert.Equal(t, filter, filter.Or(filter))

		filter.Expect("And", filter).
			Return(filter)
		assert.Equal(t, filter, filter.And(filter))

		store := NewDisconnectedStoreWithFail(t)
		defer store.Assert(t)

		store.Expect("Filter").
			Return(filter)
		assert.Equal(t, filter, store.Filter())
	})

	t.Run("raises bad calls errors on filter", func(t *testing.T) {
		filter := NewFilterWithFail(t, "unexpected length of return values")
		defer filter.Assert(t)

		filter.Expect("BeginsWith", "test", "value").
			Return(filter, filter)
		filter.BeginsWith("test", "value")

		filter.Expect("EndsWith", "test", "value").
			Return(filter, filter)
		filter.EndsWith("test", "value")

		filter.Expect("Contains", "test", "value").
			Return(filter, filter)
		filter.Contains("test", "value")

		filter.Expect("NotContains", "test", "value").
			Return(filter, filter)
		filter.NotContains("test", "value")

		filter.Expect("Eq", "test", "value").
			Return(filter, filter)
		filter.Eq("test", "value")

		filter.Expect("Lt", "test", "value").
			Return(filter, filter)
		filter.Lt("test", "value")

		filter.Expect("Gt", "test", "value").
			Return(filter, filter)
		filter.Gt("test", "value")

		filter.Expect("NotEq", "test", "value").
			Return(filter, filter)
		filter.NotEq("test", "value")

		filter.Expect("Or", filter).
			Return(filter, filter)
		filter.Or(filter)

		filter.Expect("And", filter).
			Return(filter, filter)
		filter.And(filter)

		store := NewDisconnectedStoreWithFail(t, "unexpected length of return values")
		defer store.Assert(t)

		store.Expect("Filter").
			Return(filter, filter)
		store.Filter()
	})

	t.Run("raises return value errors on filter", func(t *testing.T) {
		filter := NewFilterWithFail(t, "unexpected type of return value")
		defer filter.Assert(t)

		filter.Expect("BeginsWith", "test", "value").
			Return("")
		filter.BeginsWith("test", "value")

		filter.Expect("EndsWith", "test", "value").
			Return("")
		filter.EndsWith("test", "value")

		filter.Expect("Contains", "test", "value").
			Return("")
		filter.Contains("test", "value")

		filter.Expect("NotContains", "test", "value").
			Return("")
		filter.NotContains("test", "value")

		filter.Expect("Eq", "test", "value").
			Return("")
		filter.Eq("test", "value")

		filter.Expect("Lt", "test", "value").
			Return("")
		filter.Lt("test", "value")

		filter.Expect("Gt", "test", "value").
			Return("")
		filter.Gt("test", "value")

		filter.Expect("NotEq", "test", "value").
			Return("")
		filter.NotEq("test", "value")

		filter.Expect("Or", filter).
			Return("")
		filter.Or(filter)

		filter.Expect("And", filter).
			Return("")
		filter.And(filter)

		store := NewDisconnectedStoreWithFail(t, "unexpected type of return value")
		defer store.Assert(t)

		store.Expect("Filter").
			Return("")
		store.Filter()
	})

	t.Run("can create query instance", func(t *testing.T) {
		query := NewQueryWithFail(t)
		assert.NotNil(t, query)
		defer query.Assert(t)

		filter := NewFilter(t)
		query.Expect("Filter", filter).
			Return(query)
		assert.Equal(t, query, query.Filter(filter))

		cursor := NewCursor(t)
		query.Expect("Execute", context.TODO()).
			Return(cursor, nil)
		got, err := query.Execute(context.TODO())
		assert.Equal(t, cursor, got)
		assert.Nil(t, err)

		query.Expect("Execute", context.TODO()).
			Return(nil, errors.New("an error has occurred"))
		_, err = query.Execute(context.TODO())
		assert.NotNil(t, err)

		query.Expect("Statement").
			Return("test", []interface{}{"test"}, nil)
		statement, args, err := query.Statement()
		assert.Equal(t, "test", statement)
		assert.Equal(t, []interface{}{"test"}, args)
		assert.Nil(t, err)

		query.Expect("Statement").
			Return("", nil, errors.New("an error has occurred"))
		statement, args, err = query.Statement()
		assert.NotNil(t, err)

		query.Expect("Statement").
			Return("test", nil, nil)
		statement, args, err = query.Statement()
		assert.Equal(t, "test", statement)

		query.Expect("From", "tests").
			Return(query)
		assert.Equal(t, query, query.From("tests"))

		query.Expect("Order", "tests").
			Return(query)
		assert.Equal(t, query, query.Order("tests"))

		query.Expect("First", 1).
			Return(query)
		assert.Equal(t, query, query.First(1))

		query.Expect("After", "count", 1).
			Return(query)
		assert.Equal(t, query, query.After("count", 1))

		query.Expect("Secondary").
			Return(query)
		assert.Equal(t, query, query.Secondary())

		query.Expect("And", "count", 1).
			Return(query)
		assert.Equal(t, query, query.And("count", 1))

		query.Expect("Return", "count", 1).
			Return(query)
		assert.Equal(t, query, query.Return("count", 1))

		store := NewDisconnectedStoreWithFail(t)
		defer store.Assert(t)

		store.Expect("Query").
			Return(query)
		assert.Equal(t, query, store.Query())
	})

	t.Run("raises bad calls errors on query", func(t *testing.T) {
		query := NewQueryWithFail(t, "unexpected length of return values")
		defer query.Assert(t)

		filter := NewFilter(t)
		query.Expect("Filter", filter).
			Return(query, query)
		query.Filter(filter)

		query.Expect("Execute", context.TODO()).
			Return(1, nil, nil)
		query.Execute(context.TODO())

		query.Expect("Statement").
			Return("test", []interface{}{"test"}, nil, nil)
		query.Statement()

		query.Expect("From", "tests").
			Return(query, query)
		query.From("tests")

		query.Expect("Order", "tests").
			Return(query, query)
		query.Order("tests")

		query.Expect("First", 1).
			Return(query, query)
		query.First(1)

		query.Expect("After", "count", 1).
			Return(query, query)
		query.After("count", 1)

		query.Expect("Secondary").
			Return(query, query)
		query.Secondary()

		query.Expect("And", "count", 1).
			Return(query, query)
		query.And("count", 1)

		query.Expect("Return", "count", 1).
			Return(query, query)
		query.Return("count", 1)

		store := NewDisconnectedStoreWithFail(t, "unexpected length of return values")
		defer store.Assert(t)

		store.Expect("Query").
			Return(query, query)
		store.Query()
	})

	t.Run("raises return value errors on query", func(t *testing.T) {
		query := NewQueryWithFail(t, "unexpected type of return value")
		defer query.Assert(t)

		filter := NewFilter(t)
		query.Expect("Filter", filter).
			Return("")
		query.Filter(filter)

		query.Expect("Execute", context.TODO()).
			Return("", nil)
		query.Execute(context.TODO())

		query.Expect("Execute", context.TODO()).
			Return(0, "")
		query.Execute(context.TODO())

		query.Expect("Statement").
			Return(0, []interface{}{"test"}, nil)
		query.Statement()

		query.Expect("Statement").
			Return("", 0, nil)
		query.Statement()

		query.Expect("Statement").
			Return(0, []interface{}{"test"}, "")
		query.Statement()

		query.Expect("From", "tests").
			Return("")
		query.From("tests")

		query.Expect("Order", "tests").
			Return("")
		query.Order("tests")

		query.Expect("First", 1).
			Return("")
		query.First(1)

		query.Expect("After", "count", 1).
			Return("")
		query.After("count", 1)

		query.Expect("Secondary").
			Return("")
		query.Secondary()

		query.Expect("And", "count", 1).
			Return("")
		query.And("count", 1)

		query.Expect("Return", "count", 1).
			Return("")
		query.Return("count", 1)

		store := NewDisconnectedStoreWithFail(t, "unexpected type of return value")
		defer store.Assert(t)

		store.Expect("Query").
			Return("")
		store.Query()
	})

	t.Run("can create remove instance", func(t *testing.T) {
		remove := NewRemoveWithFail(t)
		assert.NotNil(t, remove)
		defer remove.Assert(t)

		filter := NewFilter(t)
		remove.Expect("Filter", filter).
			Return(remove)
		assert.Equal(t, remove, remove.Filter(filter))

		remove.Expect("Execute", context.TODO()).
			Return(1, nil)
		count, err := remove.Execute(context.TODO())
		assert.Equal(t, 1, count)
		assert.Nil(t, err)

		remove.Expect("Execute", context.TODO()).
			Return(0, errors.New("an error has occurred"))
		count, err = remove.Execute(context.TODO())
		assert.Equal(t, 0, count)
		assert.NotNil(t, err)

		remove.Expect("Statement").
			Return("test", []interface{}{"test"}, nil)
		statement, args, err := remove.Statement()
		assert.Equal(t, "test", statement)
		assert.Equal(t, []interface{}{"test"}, args)
		assert.Nil(t, err)

		remove.Expect("Statement").
			Return("", nil, errors.New("an error has occurred"))
		statement, args, err = remove.Statement()
		assert.NotNil(t, err)

		remove.Expect("Statement").
			Return("test", nil, nil)
		statement, args, err = remove.Statement()
		assert.Equal(t, "test", statement)

		remove.Expect("From", "tests").
			Return(remove)
		assert.Equal(t, remove, remove.From("tests"))

		remove.Expect("Order", "tests").
			Return(remove)
		assert.Equal(t, remove, remove.Order("tests"))

		remove.Expect("First", 1).
			Return(remove)
		assert.Equal(t, remove, remove.First(1))

		remove.Expect("After", "count", 1).
			Return(remove)
		assert.Equal(t, remove, remove.After("count", 1))

		store := NewDisconnectedStoreWithFail(t)
		defer store.Assert(t)

		store.Expect("Remove").
			Return(remove)
		assert.Equal(t, remove, store.Remove())
	})

	t.Run("raises bad calls errors on remove", func(t *testing.T) {
		remove := NewRemoveWithFail(t, "unexpected length of return values")
		defer remove.Assert(t)

		filter := NewFilter(t)
		remove.Expect("Filter", filter).
			Return(remove, remove)
		remove.Filter(filter)

		remove.Expect("Execute", context.TODO()).
			Return(1, nil, nil)
		remove.Execute(context.TODO())

		remove.Expect("Statement").
			Return("test", []interface{}{"test"}, nil, nil)
		remove.Statement()

		remove.Expect("From", "tests").
			Return(remove, remove)
		remove.From("tests")

		remove.Expect("Order", "tests").
			Return(remove, remove)
		remove.Order("tests")

		remove.Expect("First", 1).
			Return(remove, remove)
		remove.First(1)

		remove.Expect("After", "count", 1).
			Return(remove, remove)
		remove.After("count", 1)

		store := NewDisconnectedStoreWithFail(t, "unexpected length of return values")
		defer store.Assert(t)

		store.Expect("Remove").
			Return(remove, remove)
		store.Remove()
	})

	t.Run("raises return value errors on remove", func(t *testing.T) {
		remove := NewRemoveWithFail(t, "unexpected type of return value")
		defer remove.Assert(t)

		filter := NewFilter(t)
		remove.Expect("Filter", filter).
			Return("")
		remove.Filter(filter)

		remove.Expect("Execute", context.TODO()).
			Return("", nil)
		remove.Execute(context.TODO())

		remove.Expect("Execute", context.TODO()).
			Return(0, "")
		remove.Execute(context.TODO())

		remove.Expect("Statement").
			Return(0, []interface{}{"test"}, nil)
		remove.Statement()

		remove.Expect("Statement").
			Return("", 0, nil)
		remove.Statement()

		remove.Expect("Statement").
			Return(0, []interface{}{"test"}, "")
		remove.Statement()

		remove.Expect("From", "tests").
			Return("")
		remove.From("tests")

		remove.Expect("Order", "tests").
			Return("")
		remove.Order("tests")

		remove.Expect("First", 1).
			Return("")
		remove.First(1)

		remove.Expect("After", "count", 1).
			Return("")
		remove.After("count", 1)

		store := NewDisconnectedStoreWithFail(t, "unexpected type of return value")
		defer store.Assert(t)

		store.Expect("Remove").
			Return("")
		store.Remove()
	})

	t.Run("can create transaction instance", func(t *testing.T) {
		transaction := NewTransactionWithFail(t)
		assert.NotNil(t, transaction)
		defer transaction.Assert(t)

		transaction.Expect("Commit").
			Return(nil)
		err := transaction.Commit()
		assert.Nil(t, err)

		transaction.Expect("Commit").
			Return(errors.New("an error has occurred"))
		err = transaction.Commit()
		assert.NotNil(t, err)

		transaction.Expect("Rollback").
			Return(nil)
		err = transaction.Rollback()
		assert.Nil(t, err)

		transaction.Expect("Rollback").
			Return(errors.New("an error has occurred"))
		err = transaction.Rollback()
		assert.NotNil(t, err)

		add := NewAdd(t)
		transaction.Expect("Execute", add).
			Return(1, nil)
		count, err := transaction.Execute(add)
		assert.Equal(t, 1, count)
		assert.Nil(t, err)

		transaction.Expect("Execute", add).
			Return(0, errors.New("an error has occurred"))
		count, err = transaction.Execute(add)
		assert.Equal(t, 0, count)
		assert.NotNil(t, err)

		store := NewDisconnectedStoreWithFail(t)
		defer store.Assert(t)

		store.Expect("Transaction", context.TODO()).
			Return(transaction, nil)
		tx, err := store.Transaction(context.TODO())
		assert.Nil(t, err)
		assert.Equal(t, tx, transaction)

		store.Expect("Transaction", context.TODO()).
			Return(nil, errors.New("an error has occurred"))
		tx, err = store.Transaction(context.TODO())
		assert.NotNil(t, err)
	})

	t.Run("raises bad calls errors on transaction", func(t *testing.T) {
		transaction := NewTransactionWithFail(t, "unexpected length of return values")
		defer transaction.Assert(t)

		transaction.Expect("Commit").
			Return(nil, nil)
		transaction.Commit()

		transaction.Expect("Rollback").
			Return(nil, nil)
		transaction.Rollback()

		add := NewAdd(t)
		transaction.Expect("Execute", add).
			Return(1, nil, nil)
		transaction.Execute(add)

		store := NewDisconnectedStoreWithFail(t, "unexpected length of return values")
		defer store.Assert(t)

		store.Expect("Transaction", context.TODO()).
			Return(nil, nil, nil)
		store.Transaction(context.TODO())
	})

	t.Run("raises return value errors on transaction", func(t *testing.T) {
		transaction := NewTransactionWithFail(t, "unexpected type of return value")
		defer transaction.Assert(t)

		transaction.Expect("Commit").
			Return("")
		transaction.Commit()

		transaction.Expect("Rollback").
			Return("")
		transaction.Rollback()

		add := NewAdd(t)
		transaction.Expect("Execute", add).
			Return("", nil)
		transaction.Execute(add)

		transaction.Expect("Execute", add).
			Return(0, "")
		transaction.Execute(add)

		store := NewDisconnectedStoreWithFail(t, "unexpected type of return value")
		defer store.Assert(t)

		store.Expect("Transaction", context.TODO()).
			Return("", nil)
		store.Transaction(context.TODO())

		store.Expect("Transaction", context.TODO()).
			Return(nil, "")
		store.Transaction(context.TODO())
	})

	t.Run("can create update instance", func(t *testing.T) {
		update := NewUpdateWithFail(t)
		assert.NotNil(t, update)
		defer update.Assert(t)

		filter := NewFilter(t)
		update.Expect("Filter", filter).
			Return(update)
		assert.Equal(t, update, update.Filter(filter))

		update.Expect("Execute", context.TODO()).
			Return(1, nil)
		count, err := update.Execute(context.TODO())
		assert.Equal(t, 1, count)
		assert.Nil(t, err)

		update.Expect("Execute", context.TODO()).
			Return(0, errors.New("an error has occurred"))
		count, err = update.Execute(context.TODO())
		assert.Equal(t, 0, count)
		assert.NotNil(t, err)

		update.Expect("Statement").
			Return("test", []interface{}{"test"}, nil)
		statement, args, err := update.Statement()
		assert.Equal(t, "test", statement)
		assert.Equal(t, []interface{}{"test"}, args)
		assert.Nil(t, err)

		update.Expect("Statement").
			Return("", nil, errors.New("an error has occurred"))
		statement, args, err = update.Statement()
		assert.NotNil(t, err)

		update.Expect("Statement").
			Return("test", nil, nil)
		statement, args, err = update.Statement()
		assert.Equal(t, "test", statement)

		update.Expect("In", "tests").
			Return(update)
		assert.Equal(t, update, update.In("tests"))

		update.Expect("Item", map[string]interface{}{"count": 1}).
			Return(update)
		assert.Equal(t, update, update.Item(map[string]interface{}{"count": 1}))

		store := NewDisconnectedStoreWithFail(t)
		defer store.Assert(t)

		store.Expect("Update").
			Return(update)
		assert.Equal(t, update, store.Update())
	})

	t.Run("raises bad calls errors on update", func(t *testing.T) {
		update := NewUpdateWithFail(t, "unexpected length of return values")
		defer update.Assert(t)

		filter := NewFilter(t)
		update.Expect("Filter", filter).
			Return(update, update)
		update.Filter(filter)

		update.Expect("Execute", context.TODO()).
			Return(1, nil, nil)
		update.Execute(context.TODO())

		update.Expect("Statement").
			Return("test", []interface{}{"test"}, nil, nil)
		update.Statement()

		update.Expect("In", "tests").
			Return(update, update)
		update.In("tests")

		update.Expect("Item", map[string]interface{}{"count": 1}).
			Return(update, update)
		update.Item(map[string]interface{}{"count": 1})

		store := NewDisconnectedStoreWithFail(t, "unexpected length of return values")
		defer store.Assert(t)

		store.Expect("Update").
			Return(update, update)
		store.Update()
	})

	t.Run("raises return value errors on update", func(t *testing.T) {
		update := NewUpdateWithFail(t, "unexpected type of return value")
		defer update.Assert(t)

		filter := NewFilter(t)
		update.Expect("Filter", filter).
			Return("")
		update.Filter(filter)

		update.Expect("Execute", context.TODO()).
			Return("", nil)
		update.Execute(context.TODO())

		update.Expect("Execute", context.TODO()).
			Return(0, "")
		update.Execute(context.TODO())

		update.Expect("Statement").
			Return(0, []interface{}{"test"}, nil)
		update.Statement()

		update.Expect("Statement").
			Return("", 0, nil)
		update.Statement()

		update.Expect("Statement").
			Return(0, []interface{}{"test"}, "")
		update.Statement()

		update.Expect("In", "tests").
			Return("")
		update.In("tests")

		update.Expect("Item", map[string]interface{}{"count": 1}).
			Return("")
		update.Item(map[string]interface{}{"count": 1})

		store := NewDisconnectedStoreWithFail(t, "unexpected type of return value")
		defer store.Assert(t)

		store.Expect("Update").
			Return("")
		store.Update()
	})
}
