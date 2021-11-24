package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-tea/internal/clock"
)

func TestNew(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		client := NewClient("0.0.1", nil)
		assert.NotNil(t, client)
	})
}

func TestClient_Handler(t *testing.T) {
	t.Run("can create handler instance", func(t *testing.T) {
		client := NewClient("0.0.1", nil)
		assert.NotNil(t, client)
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

func TestHandler_Status(t *testing.T) {
	t.Run("handles status requests", func(t *testing.T) {
		now := time.Now()
		client := NewClient("0.0.1", clock.New(now).From(func() time.Time {
			return now
		}))
		resp := client.Checks.Status()
		assert.Equal(t, "0.0.1", resp.Version)
		assert.Equal(t, StatusHealthy, resp.Status)
		assert.Equal(t, map[string][]*Check{"uptime": {{
			Time: now,
			Status: StatusHealthy,
			Value: (now.Sub(now)/(1000*1000*1000)).Seconds(),
			Unit: "s",
		}}}, resp.Checks)
	})
}
