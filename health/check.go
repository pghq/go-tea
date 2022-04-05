package health

import (
	"encoding/json"
	"net/http"
	"time"
)

const (
	// StatusHealthy represents a healthy application state
	StatusHealthy Status = "healthy"

	// StatusUnhealthy represents an unhealthy application state
	StatusUnhealthy Status = "unhealthy"
)

// Check is an object representing health of an app component
type Check struct {
	Time   time.Time   `json:"time"`
	Status Status      `json:"status,omitempty"`
	Value  interface{} `json:"observedValue"`
	Unit   string      `json:"observedUnit"`
}

// NewHealthyCheck creates a check, denoting it as healthy
func NewHealthyCheck(observedAt time.Time, value interface{}, unit string) *Check {
	c := &Check{
		Time:   observedAt,
		Status: StatusHealthy,
		Value:  value,
		Unit:   unit,
	}

	return c
}

// NewDependencyCheck creates a dependency check
func NewDependencyCheck(observedAt time.Time, dependencyURL string) *Check {
	c := &Check{
		Time:   observedAt,
		Status: StatusHealthy,
		Unit:   "application/health+json",
	}

	response, err := http.Get(dependencyURL)
	if err != nil {
		c.Status = StatusUnhealthy
		return c
	}

	var check map[string]interface{}
	if err := json.NewDecoder(response.Body).Decode(&check); err != nil {
		c.Status = StatusUnhealthy
		return c
	}

	c.Value = check
	return c
}
