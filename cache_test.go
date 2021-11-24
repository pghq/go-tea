package tea

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-tea/internal"
	"github.com/pghq/go-tea/internal/clock"
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
		assert.False(t, IsFatal(err))
	})

	t.Run("raises casting errors", func(t *testing.T) {
		c := NewLRU()
		key, _ := CacheKey("item")
		c.lru.Add(key, "test")
		_, err := c.Get("item")
		assert.NotNil(t, err)
		assert.True(t, IsFatal(err))
	})

	t.Run("raises expiration errors", func(t *testing.T) {
		c := NewLRU()
		_ = c.Insert("item", "test", time.Nanosecond)
		time.Sleep(time.Nanosecond)
		_, err := c.Get("item")
		assert.NotNil(t, err)
		assert.False(t, IsFatal(err))
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

func TestCacheMiddleware_Handle(t *testing.T) {
	c := NewLRU()
	r := httptest.NewRequest("GET", "/tests?name=foo", nil)

	t.Run("can create instance", func(t *testing.T) {
		w := httptest.NewRecorder()
		res := NewCacheResponseWatcher(c, &CacheConfig{
			PositiveTTL: time.Second,
			NegativeTTL: time.Second,
		}, w, CacheRequestKey(r, "name"))
		assert.NotNil(t, res)

		opts := []CacheOption{
			PositiveCacheFor(time.Second),
			NegativeCacheFor(time.Second),
		}
		m := NewCacheMiddleware(c).With(opts...)
		assert.NotNil(t, m)
		assert.Equal(t, opts, m.opts)
	})

	t.Run("calls origin if no cache is present", func(t *testing.T) {
		w := httptest.NewRecorder()
		m := NewCacheMiddleware(nil)
		m.Handle(internal.NoopHandler).ServeHTTP(w, r)
	})

	t.Run("raises fatal errors", func(t *testing.T) {
		SetGlobalLogWriter(io.Discard)
		defer ResetGlobalLogger()
		w := httptest.NewRecorder()
		m := NewCacheMiddleware(c)
		c.lru.Add(CacheRequestKey(r), "test")
		defer c.lru.Remove(CacheRequestKey(r))
		m.Handle(internal.NoopHandler).ServeHTTP(w, r)
	})

	t.Run("raises cached response errors", func(t *testing.T) {
		SetGlobalLogWriter(io.Discard)
		defer ResetGlobalLogger()
		w := httptest.NewRecorder()
		m := NewCacheMiddleware(c)
		_ = c.Insert(CacheRequestKey(r), "test", time.Minute)
		defer c.lru.Remove(CacheRequestKey(r))
		m.Handle(internal.NoopHandler).ServeHTTP(w, r)
	})

	t.Run("sends response", func(t *testing.T) {
		w := httptest.NewRecorder()
		m := NewCacheMiddleware(c)
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		})).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})

	t.Run("caches response", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/tests?name=bar&cachebuster=1892312938", nil)
		w := httptest.NewRecorder()
		m := NewCacheMiddleware(c).With(NegativeCacheFor(time.Minute), PositiveCacheFor(time.Minute), UseCacheQuery("name"))
		m.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("ok"))
		})).ServeHTTP(w, r)

		r = httptest.NewRequest("GET", "/tests?name=bar&cachebuster=123y19238y", nil)
		w = httptest.NewRecorder()
		m.Handle(internal.NoopHandler).ServeHTTP(w, r)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "ok", w.Body.String())
		assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
		assert.NotEmpty(t, w.Header().Get("Cached-At"))
	})
}
