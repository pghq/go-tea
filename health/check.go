package health

import "time"

const (
	// StatusHealthy represents a healthy application state
	StatusHealthy Status = "healthy"
)

// Check is an object representing health of an app component
type Check struct {
	Time   time.Time   `json:"time"`
	Status Status      `json:"status,omitempty"`
	Value  interface{} `json:"observedValue"`
	Unit   string      `json:"observedUnit"`
}

// NewHealthyCheck creates a check, denoting it as unhealthy
func NewHealthyCheck(observedAt time.Time, value interface{}, unit string) *Check {
	c := &Check{
		Time:   observedAt,
		Status: StatusHealthy,
		Value:  value,
		Unit:   unit,
	}

	return c
}
