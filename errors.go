package tea

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// stacktrace is an error type containing an error code string.
type stacktrace struct {
	code  int
	cause error
	stack error
}

// Error implements the error interface
// it represents the message associated with the error
func (e *stacktrace) Error() string {
	return e.stack.Error()
}

func (e *stacktrace) Format(s fmt.State, verb rune) {
	if fe, ok := e.stack.(fmt.Formatter); ok {
		fe.Format(s, verb)
	}
}

// Stack adds stacktrace to an error
func Stack(err error) error {
	if IsError(err, context.Canceled) {
		return stack(err, http.StatusBadRequest)
	}

	if IsError(err, context.DeadlineExceeded) {
		return stack(err, http.StatusRequestTimeout)
	}

	if ae, ok := err.(*stacktrace); ok {
		return ae
	}

	return stack(err, http.StatusInternalServerError)
}

// Err creates an internal error from an error message
func Err(v ...interface{}) error {
	return Stack(errors.New(fmt.Sprint(v...)))
}

// Errf creates an internal error from a formatted error message
func Errf(format string, v ...interface{}) error {
	return Err(errors.New(fmt.Sprintf(format, v...)))
}

// AsErrTransfer creates a http error
func AsErrTransfer(code int, err error) error {
	return stack(err, code)
}

// ErrTransfer creates a http error from a msg
func ErrTransfer(code int, v ...interface{}) error {
	return AsErrTransfer(code, Err(v...))
}

// AsErrBadRequest creates a bad request error
func AsErrBadRequest(err error) error {
	return AsErrTransfer(http.StatusBadRequest, err)
}

// ErrBadRequest creates a bad request error from a msg
func ErrBadRequest(v ...interface{}) error {
	return ErrTransfer(http.StatusBadRequest, v...)
}

// AsErrNoContent creates a no content error
func AsErrNoContent(err error) error {
	return AsErrTransfer(http.StatusNoContent, err)
}

// ErrNoContent creates a no content error from a msg
func ErrNoContent(v ...interface{}) error {
	return ErrTransfer(http.StatusNoContent, v...)
}

// AsErrNotFound creates a not found error
func AsErrNotFound(err error) error {
	return AsErrTransfer(http.StatusNotFound, err)
}

// ErrNotFound creates a not found error from a msg
func ErrNotFound(v ...interface{}) error {
	return ErrTransfer(http.StatusNotFound, v...)
}

// AsError finds first error in chain matching target.
func AsError(err error, target interface{}) bool {
	// why would we ever want to panic on this call???
	defer func() {
		recover()
	}()

	stack, aok := err.(*stacktrace)
	if aok {
		err = stack.cause
	}

	against, tok := target.(*stacktrace)
	if tok {
		target = against.cause
	}

	return (aok && tok) || errors.As(err, target)
}

// IsError finds first error in chain matching target.
func IsError(err error, target error) bool {
	if ae, ok := err.(*stacktrace); ok {
		err = ae.cause
	}

	return errors.Is(err, target)
}

// ErrStatus gets an HTTP status code from an error
func ErrStatus(err error) int {
	if IsError(err, context.Canceled) {
		return http.StatusBadRequest
	}

	if IsError(err, context.DeadlineExceeded) {
		return http.StatusRequestTimeout
	}

	if e, ok := err.(*stacktrace); ok {
		return e.code
	}

	return http.StatusInternalServerError
}

// IsNotFound checks if an error is a not found application error
func IsNotFound(err error) bool {
	return err != nil && ErrStatus(err) == http.StatusNotFound
}

// IsBadRequest checks if an error is a bad request application error
func IsBadRequest(err error) bool {
	return err != nil && ErrStatus(err) == http.StatusBadRequest
}

// IsFatal checks if an error is a fatal application error
func IsFatal(err error) bool {
	return err != nil && ErrStatus(err) >= http.StatusInternalServerError
}

// stack creates an error with a given code and stack trace.
func stack(err error, code int) error {
	return &stacktrace{
		cause: err,
		stack: errors.WithStack(err),
		code:  code,
	}
}
