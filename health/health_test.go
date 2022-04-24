package health

import (
	"net/http"
	"net/http/httptest"
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

	t.Run("handles status requests with dependencies", func(t *testing.T) {
		t.Run("bad dependency url", func(t *testing.T) {
			dep := NewDependencyCheck(time.Now(), "http//")
			assert.Equal(t, StatusUnhealthy, dep.Status)

			s := NewService("0.0.1")
			s.AddDependency("dep", "http//")
			assert.Equal(t, StatusHealthyWithConcerns, s.Status().Status)
		})

		t.Run("bad dependency response", func(t *testing.T) {
			dep := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{bad}`))
			}))
			defer dep.Close()
			assert.Equal(t, StatusUnhealthy, NewDependencyCheck(time.Now(), dep.URL).Status)
		})

		now := time.Now()
		s := NewService("0.0.1")
		s.now = func() time.Time { return now }
		s.start = now

		dep := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{
				"version": "0.1.0",
				"status": "healthy", 
			    "checks": {
					"uptime": [{
						"time": "2022-04-05T23:16:10.658971001Z",
						"status":"healthy",
						"observedValue":947804.46370581,
						"observedUnit":"s"
					}]
				}
			}`))
		}))
		defer dep.Close()

		s.AddDependency("dep", dep.URL)

		resp := s.Status()
		assert.Equal(t, "0.0.1", resp.Version)
		assert.Equal(t, StatusHealthy, resp.Status)
		assert.Equal(t, map[string][]*Check{
			"uptime": {{
				Time:   now,
				Status: StatusHealthy,
				Value:  (now.Sub(now) / (1000 * 1000 * 1000)).Seconds(),
				Unit:   "s",
			}},
			"dep": {{
				Time:   now,
				Status: StatusHealthy,
				Value: map[string]interface{}{
					"version": "0.1.0",
					"status":  "healthy",
					"checks": map[string]interface{}{
						"uptime": []interface{}{
							map[string]interface{}{
								"time":          "2022-04-05T23:16:10.658971001Z",
								"status":        "healthy",
								"observedValue": 947804.46370581,
								"observedUnit":  "s",
							},
						},
					},
				},
				Unit: "application/health+json",
			}},
		}, resp.Checks)
	})
}
