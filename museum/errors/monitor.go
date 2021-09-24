package errors

import (
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/pghq/go-museum/museum/log"
)

// monitor the initial error monitor with sensible defaults.
var monitor = NewMonitor()

const (
	// defaultFlushTimeout is the default time to wait for panic errors to be sent
	defaultFlushTimeout = 5 * time.Second
)

// Monitor is an instance of a sentry based Monitor
type Monitor struct {
	flushTimeout time.Duration
}

func (m *Monitor) Emit(err error){
	hub := sentry.CurrentHub().Clone()
	hub.CaptureException(err)
}

func (m *Monitor) EmitHTTP(r *http.Request, err error){
	hub := sentry.CurrentHub().Clone()
	hub.Scope().SetRequest(r)
	hub.CaptureException(err)
}

func (m *Monitor) Recover(err interface{}){
	sentry.CurrentHub().Recover(err)
	sentry.Flush(m.flushTimeout)
	log.Error(fmt.Errorf("%+v", err))
}

func NewMonitor() *Monitor{
	return &Monitor{
		flushTimeout: defaultFlushTimeout,
	}
}

// MonitorConfig is the configuration for initializing the monitor
type MonitorConfig struct {
	Dsn string
	Version string
	Environment string
	FlushTimeout time.Duration
}

// CurrentMonitor returns an instance of the global monitor.
func CurrentMonitor() *Monitor{
	return monitor
}

// Init initializes the global Monitor
func Init(conf MonitorConfig) error {
	m := CurrentMonitor()

	sentryOpts := sentry.ClientOptions{
		Dsn: conf.Dsn,
		AttachStacktrace: true,
		Release:          conf.Version,
		Environment:     conf.Environment,
	}

	if conf.FlushTimeout != 0{
		m.flushTimeout = conf.FlushTimeout
	}

	if err := sentry.Init(sentryOpts); err != nil{
		return err
	}

	return nil
}
