package health

import "sync"

const (
	// UptimeCheckKey is the key for the uptime health measurement
	UptimeCheckKey = "uptime"
)

// Status is a nice name representing the state of the application
type Status string

// StatusResponse is the response for the health check status API
type StatusResponse struct {
	Version string              `json:"version"`
	Status  Status              `json:"status"`
	Checks  map[string][]*Check `json:"checks"`
}

// WithCheck adds a new check to the response
func (s *StatusResponse) WithCheck(key string, check *Check) *StatusResponse {
	s.Checks[key] = append(s.Checks[key], check)

	return s
}

// Status is an API endpoint that presents the health of the current application
// https://tools.ietf.org/id/draft-inadarei-api-health-check-05.html
func (s Service) Status() *StatusResponse {
	status := StatusResponse{
		Version: s.version,
		Checks:  make(map[string][]*Check),
		Status:  StatusHealthy,
	}

	mutex := sync.Mutex{}
	status.WithCheck(UptimeCheckKey, s.Uptime())
	wg := sync.WaitGroup{}
	for _, dep := range s.dependencies {
		wg.Add(1)
		go func(dep dependency) {
			defer wg.Done()
			check := NewDependencyCheck(s.now(), dep.url)
			mutex.Lock()
			defer mutex.Unlock()

			if check.Status != StatusHealthy {
				status.Status = StatusHealthyWithConcerns
			}

			status.WithCheck(dep.name, check)
		}(dep)
	}

	wg.Wait()

	return &status
}

func (s *Service) AddDependency(dependencyName string, dependencyURL string) {
	s.dependencies = append(s.dependencies, dependency{
		name: dependencyName,
		url:  dependencyURL,
	})
}
