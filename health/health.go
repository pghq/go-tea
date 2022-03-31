package health

import (
	"time"
)

// Service is a shared service for all health services
type Service struct {
	now     func() time.Time
	start   time.Time
	version string
}

// NewService creates a new health client instance
func NewService(version string) Service {
	return Service{
		version: version,
		now:     time.Now,
		start:   time.Now(),
	}
}
