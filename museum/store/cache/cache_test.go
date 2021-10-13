package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/diagnostic/errors"
	"github.com/pghq/go-museum/museum/internal/clock"
)

func TestInsert(t *testing.T) {
	t.Run("EncodeKeyError", func(t *testing.T) {
		c := New()
		err := c.Insert(func() {}, "test", time.Minute)
		assert.NotNil(t, err)
	})

	t.Run("NoError", func(t *testing.T) {
		c := New()
		err := c.Insert("item", "test", time.Minute)
		assert.Nil(t, err)
		i, _ := c.Get("item")
		assert.NotNil(t, i)
	})
}

func TestRemove(t *testing.T) {
	t.Run("EncodeKeyError", func(t *testing.T) {
		c := New()
		_ = c.Insert(func() {}, "test", time.Minute)
		err := c.Remove(func() {})
		assert.NotNil(t, err)
	})

	t.Run("NoError", func(t *testing.T) {
		c := New()
		_ = c.Insert("item", "test", time.Minute)
		err := c.Remove("item")
		assert.Nil(t, err)
		i, _ := c.Get("item")
		assert.Nil(t, i)
	})
}

func TestGet(t *testing.T) {
	t.Run("EncodeKeyError", func(t *testing.T) {
		c := New()
		_ = c.Insert(func() {}, "test", time.Minute)
		_, err := c.Get(func() {})
		assert.NotNil(t, err)
	})

	t.Run("NotFoundError", func(t *testing.T) {
		c := New()
		_, err := c.Get("item")
		assert.NotNil(t, err)
		assert.False(t, errors.IsFatal(err))
	})

	t.Run("CastError", func(t *testing.T) {
		c := New()
		key, _ := encodeKey("item")
		c.lru.Add(key, "test")
		_, err := c.Get("item")
		assert.NotNil(t, err)
		assert.True(t, errors.IsFatal(err))
	})

	t.Run("ExpiredError", func(t *testing.T) {
		c := New()
		_ = c.Insert("item", "test", time.Nanosecond)
		time.Sleep(time.Nanosecond)
		_, err := c.Get("item")
		assert.NotNil(t, err)
		assert.False(t, errors.IsFatal(err))
	})

	t.Run("NoError", func(t *testing.T) {
		c := New()
		_ = c.Insert("item", "test", time.Minute)
		i, _ := c.Get("item")
		assert.NotNil(t, i)
		assert.Equal(t, i.Value(), "test")
	})
}

func TestLen(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		c := New()
		c.SetCapacity(1)
		_ = c.Insert("item1", "test", time.Minute)
		_ = c.Insert("item2", "test", time.Minute)
		assert.Equal(t, c.Len(), 1)
	})
}

func TestItemCachedAt(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		c := New()
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

func TestItemValue(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		c := New()
		_ = c.Insert("item", "test", time.Minute)
		i, _ := c.Get("item")
		assert.NotNil(t, i)
		assert.Equal(t, i.Value(), "test")
	})
}
