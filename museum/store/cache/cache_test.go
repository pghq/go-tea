package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/internal/clock"
)

func TestLRU_Insert(t *testing.T) {
	t.Run("raises encode key errors", func(t *testing.T) {
		c := NewLRU()
		err := c.Insert(func() {}, "test", time.Minute)
		assert.NotNil(t, err)
	})

	t.Run("can insert", func(t *testing.T) {
		c := NewLRU()
		err := c.Insert("item", "test", time.Minute)
		assert.Nil(t, err)
		i, _ := c.Get("item")
		assert.NotNil(t, i)
	})
}

func TestLRU_Remove(t *testing.T) {
	t.Run("raises encode key errors", func(t *testing.T) {
		c := NewLRU()
		_ = c.Insert(func() {}, "test", time.Minute)
		err := c.Remove(func() {})
		assert.NotNil(t, err)
	})

	t.Run("can remove", func(t *testing.T) {
		c := NewLRU()
		_ = c.Insert("item", "test", time.Minute)
		err := c.Remove("item")
		assert.Nil(t, err)
		i, _ := c.Get("item")
		assert.Nil(t, i)
	})
}

func TestGet(t *testing.T) {
	t.Run("raises encode key errors", func(t *testing.T) {
		c := NewLRU()
		_ = c.Insert(func() {}, "test", time.Minute)
		_, err := c.Get(func() {})
		assert.NotNil(t, err)
	})

	t.Run("raises not found errors", func(t *testing.T) {
		c := NewLRU()
		_, err := c.Get("item")
		assert.NotNil(t, err)
		assert.False(t, errors.IsFatal(err))
	})

	t.Run("raises casting errors", func(t *testing.T) {
		c := NewLRU()
		key, _ := encodeKey("item")
		c.lru.Add(key, "test")
		_, err := c.Get("item")
		assert.NotNil(t, err)
		assert.True(t, errors.IsFatal(err))
	})

	t.Run("raises expiration errors", func(t *testing.T) {
		c := NewLRU()
		_ = c.Insert("item", "test", time.Nanosecond)
		time.Sleep(time.Nanosecond)
		_, err := c.Get("item")
		assert.NotNil(t, err)
		assert.False(t, errors.IsFatal(err))
	})

	t.Run("can retrieve values", func(t *testing.T) {
		c := NewLRU()
		_ = c.Insert("item", "test", time.Minute)
		i, _ := c.Get("item")
		assert.NotNil(t, i)
		assert.Equal(t, i.Value(), "test")
	})
}

func TestLRU_Len(t *testing.T) {
	t.Run("calculates length", func(t *testing.T) {
		c := NewLRU()
		c.SetCapacity(1)
		_ = c.Insert("item1", "test", time.Minute)
		_ = c.Insert("item2", "test", time.Minute)
		assert.Equal(t, c.Len(), 1)
	})
}

func TestItem_CachedAt(t *testing.T) {
	t.Run("keeps track of cache time", func(t *testing.T) {
		c := NewLRU()
		now := time.Now()
		c.setClock(clock.New(now).From(func() time.Time {
			return now
		}))
		_ = c.Insert("item", "test", time.Minute)
		i, _ := c.Get("item")
		assert.NotNil(t, i)
		assert.Equal(t, i.CachedAt(), now)
	})
}

func TestItem_Value(t *testing.T) {
	t.Run("can retrieve underlying value", func(t *testing.T) {
		c := NewLRU()
		_ = c.Insert("item", "test", time.Minute)
		i, _ := c.Get("item")
		assert.NotNil(t, i)
		assert.Equal(t, i.Value(), "test")
	})
}
