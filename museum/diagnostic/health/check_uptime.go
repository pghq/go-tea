package health

// Uptime is an API endpoint that determines the uptime of the application (in seconds)
func (s *CheckService) Uptime() *Check {
	now := s.clock.Now()
	uptime := now.Sub(s.clock.Start()).Seconds()

	return NewHealthyCheck(now, uptime, "s")
}
