package tea

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

// monitor the initial error monitor with sensible defaults.
var monitor = NewMonitor()

// exitFunc is the function that is called on exit
var exitFunc = os.Exit

// exitMutex is the mutex for exit func
var exitMutex = sync.RWMutex{}

const (
	// defaultFlushTimeout is the default time to wait for panic errors to be sent
	defaultFlushTimeout = 5 * time.Second
)

// applicationError is an error type containing an error code string.
type applicationError struct {
	code  int
	cause error
}

// Error implements the error interface
// it represents the message associated with the error
func (e *applicationError) Error() string {
	return e.cause.Error()
}

func (e *applicationError) Format(s fmt.State, verb rune) {
	if fe, ok := e.cause.(fmt.Formatter); ok {
		fe.Format(s, verb)
	}
}

// Error adds stacktrace to an error
func Error(err error) error {
	if IsError(err, context.Canceled) {
		return newApplicationError(err, http.StatusBadRequest)
	}

	if IsError(err, context.DeadlineExceeded) {
		return newApplicationError(err, http.StatusRequestTimeout)
	}

	if ae, ok := err.(*applicationError); ok {
		return newApplicationError(err, ae.code)
	}

	return newApplicationError(err, http.StatusInternalServerError)
}

// NewError creates an internal error from an error message
func NewError(v ...interface{}) error {
	return Error(errors.New(fmt.Sprint(v...)))
}

// NewErrorf creates an internal error from a formatted error message
func NewErrorf(format string, v ...interface{}) error {
	return Error(errors.New(fmt.Sprintf(format, v...)))
}

// HTTPError creates a http error
func HTTPError(code int, err error) error {
	return newApplicationError(err, code)
}

// NewHTTPError creates a http error from a msg
func NewHTTPError(code int, v ...interface{}) error {
	return HTTPError(code, NewError(v...))
}

// BadRequest creates a bad request error
func BadRequest(err error) error {
	return HTTPError(http.StatusBadRequest, err)
}

// NewBadRequest creates a bad request error from a msg
func NewBadRequest(v ...interface{}) error {
	return NewHTTPError(http.StatusBadRequest, v...)
}

// NoContent creates a no content error
func NoContent(err error) error {
	return HTTPError(http.StatusNoContent, err)
}

// NewNoContent creates a no content error from a msg
func NewNoContent(v ...interface{}) error {
	return NewHTTPError(http.StatusNoContent, v...)
}

// AsError finds first error in chain matching target.
func AsError(err error, target interface{}) bool {
	// why would we ever want to panic on this call???
	defer func() {
		recover()
	}()

	return errors.As(err, target)
}

// IsError finds first error in chain matching target.
func IsError(err error, target error) bool {
	return errors.Is(err, target)
}

// IsFatal checks if an error is a fatal application error
func IsFatal(err error) bool {
	return err != nil && StatusCode(err) >= http.StatusInternalServerError
}

// StatusCode gets an HTTP error code from an error
func StatusCode(err error) int {
	if IsError(err, context.Canceled) {
		return http.StatusBadRequest
	}

	if IsError(err, context.DeadlineExceeded) {
		return http.StatusRequestTimeout
	}

	if e, ok := err.(*applicationError); ok {
		return e.code
	}

	return http.StatusInternalServerError
}

// SendError emits fatal errors to global log and monitor.
func SendError(err error) {
	if !IsFatal(err) {
		return
	}

	l := CurrentLogger()
	l.Error(err)
	monitor.Send(err)
}

// Fatal sends a fatal error and exits
func Fatal(err error) {
	exitMutex.RLock()
	defer exitMutex.RUnlock()
	SendError(err)
	exitFunc(1)
}

// SetGlobalExitFunc sets the global exit function
func SetGlobalExitFunc(fn func(code int)) {
	exitMutex.Lock()
	defer exitMutex.Unlock()
	exitFunc = fn
}

// ResetGlobalExitFunc resets the global exit function
func ResetGlobalExitFunc() {
	exitMutex.Lock()
	defer exitMutex.Unlock()
	exitFunc = os.Exit
}

// SendHTTP replies to the request with an error
// and emits fatal http errors to global log and monitor.
func SendHTTP(w http.ResponseWriter, r *http.Request, err error) {
	msg := err.Error()
	status := StatusCode(err)
	if IsFatal(err) {
		m := CurrentMonitor()
		l := CurrentLogger()
		l.HTTPError(r, status, err)
		m.SendHTTP(r, err)
		msg = http.StatusText(status)
	}

	http.Error(w, msg, status)
}

// Recover recovers panics
func Recover(err interface{}) {
	if err != nil {
		m := CurrentMonitor()
		m.Recover(err)
	}
}

// NotFound creates a not found error
func NotFound(err error) error {
	return HTTPError(http.StatusNotFound, err)
}

// NewNotFound creates a not found error from a msg
func NewNotFound(v ...interface{}) error {
	return NewHTTPError(http.StatusNotFound, v...)
}

// IsNotFound checks if an error is a not found application error
func IsNotFound(err error) bool {
	return err != nil && StatusCode(err) == http.StatusNotFound
}

// IsBadRequest checks if an error is a bad request application error
func IsBadRequest(err error) bool {
	return err != nil && StatusCode(err) == http.StatusBadRequest
}

// SendNotAuthorized sends a not authorized error
func SendNotAuthorized(w http.ResponseWriter, r *http.Request, err error) {
	if IsFatal(err) {
		SendHTTP(w, r, err)
		return
	}

	SendHTTP(w, r, HTTPError(http.StatusUnauthorized, err))
}

// SendNewNotAuthorized sends a not authorized error message
func SendNewNotAuthorized(w http.ResponseWriter, r *http.Request, v ...interface{}) {
	SendNotAuthorized(w, r, NewHTTPError(http.StatusUnauthorized, v...))
}

// Monitor is an instance of a sentry based Monitor
type Monitor struct {
	flushTimeout time.Duration
}

// Send sends an error to the backend monitor
func (m *Monitor) Send(err error) {
	hub := sentry.CurrentHub().Clone()
	hub.CaptureException(err)
}

// SendHTTP sends an error decorated with a http request to the backend monitor
func (m *Monitor) SendHTTP(r *http.Request, err error) {
	hub := sentry.CurrentHub().Clone()
	hub.Scope().SetRequest(r)
	hub.CaptureException(err)
}

// Recover sends a panic to the backend monitor
func (m *Monitor) Recover(err interface{}) {
	sentry.CurrentHub().Recover(err)
	sentry.Flush(m.flushTimeout)
	l := CurrentLogger()
	l.Error(fmt.Errorf("%+v", err))
}

// NewMonitor creates a new monitor for handling errors
func NewMonitor() *Monitor {
	return &Monitor{
		flushTimeout: defaultFlushTimeout,
	}
}

// MonitorConfig is the configuration for initializing the monitor
type MonitorConfig struct {
	Dsn          string
	Version      string
	Environment  string
	FlushTimeout time.Duration
}

// CurrentMonitor returns an instance of the global monitor.
func CurrentMonitor() *Monitor {
	return monitor
}

// Init initializes the global Monitor
func Init(conf MonitorConfig) error {
	m := CurrentMonitor()

	sentryOpts := sentry.ClientOptions{
		Dsn:              conf.Dsn,
		AttachStacktrace: true,
		Release:          conf.Version,
		Environment:      conf.Environment,
	}

	if conf.FlushTimeout != 0 {
		m.flushTimeout = conf.FlushTimeout
	}

	if err := sentry.Init(sentryOpts); err != nil {
		return err
	}

	return nil
}

// newApplicationError creates an error with a given code and stack trace.
func newApplicationError(err error, code int) error {
	ae := &applicationError{
		cause: errors.WithStack(err),
		code:  code,
	}

	return ae
}
