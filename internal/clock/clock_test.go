package clock

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		now := time.Now()
		c := New(now)
		assert.NotNil(t, c)
	})
}

func TestClock_From(t *testing.T) {
	t.Run("can set new value", func(t *testing.T) {
		now := time.Now()
		c := New(now).From(func() time.Time {
			return now
		})
		assert.NotNil(t, c)
		assert.Equal(t, c.Now(), now)
	})
}

func TestClock_Now(t *testing.T) {
	t.Run("calculates current time", func(t *testing.T) {
		now := time.Now()
		c := New(now)
		assert.NotNil(t, c)
		assert.True(t, c.Now().After(now))
	})
}

func TestClock_Start(t *testing.T) {
	t.Run("keeps track of start time", func(t *testing.T) {
		now := time.Now()
		c := New(now)
		assert.NotNil(t, c)
		assert.Equal(t, c.Start(), now)
	})
}
