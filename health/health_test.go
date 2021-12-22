package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		s := NewService("0.0.1")
		assert.NotNil(t, s)
	})
}

func TestNewHealthyCheck(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		now := time.Now()
		check := NewHealthyCheck(now, "1", "ms")
		assert.NotNil(t, check)
		assert.Equal(t, check.Status, StatusHealthy)
		assert.Equal(t, check.Time, now)
		assert.Equal(t, check.Value, "1")
		assert.Equal(t, check.Unit, "ms")
	})
}

func TestStatusResponse_WithCheck(t *testing.T) {
	t.Run("can add check", func(t *testing.T) {
		health := &StatusResponse{Checks: make(map[string][]*Check)}
		now := time.Now()
		check := NewHealthyCheck(now, "1", "ms")
		health = health.WithCheck("uptime", check)
		assert.Equal(t, map[string][]*Check{
			"uptime": {check},
		}, health.Checks)
	})
}

func TestCheckService_Status(t *testing.T) {
	t.Run("handles status requests", func(t *testing.T) {
		health := &StatusResponse{Checks: make(map[string][]*Check)}
		now := time.Now()
		check := NewHealthyCheck(now, "1", "ms")
		health = health.WithCheck("uptime", check)
		assert.Equal(t, map[string][]*Check{
			"uptime": {check},
		}, health.Checks)
	})
}

func TestService_Status(t *testing.T) {
	t.Run("handles status requests", func(t *testing.T) {
		now := time.Now()
		s := NewService("0.0.1")
		s.now = func() time.Time { return now }
		s.start = now
		resp := s.Status()
		assert.Equal(t, "0.0.1", resp.Version)
		assert.Equal(t, StatusHealthy, resp.Status)
		assert.Equal(t, map[string][]*Check{"uptime": {{
			Time:   now,
			Status: StatusHealthy,
			Value:  (now.Sub(now) / (1000 * 1000 * 1000)).Seconds(),
			Unit:   "s",
		}}}, resp.Checks)
	})
}
