package health

import (
	"net/http"

	"github.com/pghq/go-museum/museum/transmit/response"
)

const (
	// UptimeCheckKey is the key for the uptime health measurement
	UptimeCheckKey = "uptime"
)

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
func (s *CheckService) Status() *StatusResponse {
	status := StatusResponse{
		Version: s.version,
		Checks:  make(map[string][]*Check),
		Status:  StatusHealthy,
	}

	status.WithCheck(UptimeCheckKey, s.Uptime())

	return &status
}

// Status is the corresponding HTTP handler for the status API
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	status := h.client.Checks.Status()
	response.New(w, r).Body(status).Send()
}
