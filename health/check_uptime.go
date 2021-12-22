package health

// Uptime is an API endpoint that determines the uptime of the application (in seconds)
func (s Service) Uptime() *Check {
	now := s.now()
	uptime := now.Sub(s.start).Seconds()
	return NewHealthyCheck(now, uptime, "s")
}
