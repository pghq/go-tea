package trail

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

// Stacktrace adds stacktrace to an error
func Stacktrace(err error) error {
	if err == nil {
		return nil
	}

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

// NewError creates an internal error from an error message
func NewError(msg string) error {
	return Stacktrace(errors.New(msg))
}

// NewErrorf creates an internal error from a formatted error message
func NewErrorf(format string, v ...interface{}) error {
	return NewError(fmt.Sprintf(format, v...))
}

// NewErrorWithCode creates an error with code from a msg
func NewErrorWithCode(msg string, code int) error {
	return errorTransfer(code, NewError(msg))
}

// ErrorBadRequest creates a bad request error
func ErrorBadRequest(err error) error {
	return errorTransfer(http.StatusBadRequest, err)
}

// NewErrorBadRequest creates a bad request error from a msg
func NewErrorBadRequest(msg string) error {
	return NewErrorWithCode(msg, http.StatusBadRequest)
}

// IsBadRequest checks if an error is a bad request application error
func IsBadRequest(err error) bool {
	return IsError(err, context.Canceled) || err != nil && StatusCode(err) == http.StatusBadRequest
}

// ErrorNoContent creates a no content error
func ErrorNoContent(err error) error {
	return errorTransfer(http.StatusNoContent, err)
}

// NewErrorNoContent creates a no content error from a msg
func NewErrorNoContent(msg string) error {
	return NewErrorWithCode(msg, http.StatusNoContent)
}

// IsNoContent checks if an error is a no content application error
func IsNoContent(err error) bool {
	return err != nil && StatusCode(err) == http.StatusNoContent
}

// ErrorNotFound creates a not found error
func ErrorNotFound(err error) error {
	return errorTransfer(http.StatusNotFound, err)
}

// NewErrorNotFound creates a not found error from a msg
func NewErrorNotFound(msg string) error {
	return NewErrorWithCode(msg, http.StatusNotFound)
}

// IsNotFound checks if an error is a not found application error
func IsNotFound(err error) bool {
	return err != nil && StatusCode(err) == http.StatusNotFound
}

// ErrorConflict creates a conflict error
func ErrorConflict(err error) error {
	return errorTransfer(http.StatusConflict, err)
}

// NewErrorConflict creates a conflict error from a msg
func NewErrorConflict(msg string) error {
	return NewErrorWithCode(msg, http.StatusConflict)
}

// IsConflict checks if an error is a conflict application error
func IsConflict(err error) bool {
	return err != nil && StatusCode(err) == http.StatusConflict
}

// ErrorTooManyRequests creates a too many requests error
func ErrorTooManyRequests(err error) error {
	return errorTransfer(http.StatusTooManyRequests, err)
}

// NewErrorTooManyRequests creates a too many requests error from a msg
func NewErrorTooManyRequests(msg string) error {
	return NewErrorWithCode(msg, http.StatusTooManyRequests)
}

// IsTooManyRequests checks if an error is a too many requests application error
func IsTooManyRequests(err error) bool {
	return err != nil && StatusCode(err) == http.StatusTooManyRequests
}

// ErrorNotAuthorized creates an unauthorized error
func ErrorNotAuthorized(err error) error {
	return errorTransfer(http.StatusUnauthorized, err)
}

// NewErrorNotAuthorized creates an unauthorized error from a msg
func NewErrorNotAuthorized(msg string) error {
	return NewErrorWithCode(msg, http.StatusUnauthorized)
}

// IsNotAuthorized checks if an error is an unauthorized application error
func IsNotAuthorized(err error) bool {
	return IsError(err, context.Canceled) || err != nil && StatusCode(err) == http.StatusUnauthorized
}

// AsError finds first error in chain matching target.
func AsError(err error, target interface{}) bool {
	// why would we ever want to panic on this call???
	defer func() {
		_ = recover()
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

// IsFatal checks if an error is a fatal application error
func IsFatal(err error) bool {
	return err != nil && StatusCode(err) >= http.StatusInternalServerError
}

// StatusCode gets an HTTP status code from an error
func StatusCode(err error) int {
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

// errorTransfer creates a http error
func errorTransfer(code int, err error) error {
	return stack(err, code)
}

// stack creates an error with a given code and stack trail.
func stack(err error, code int) error {
	return &stacktrace{
		cause: err,
		stack: errors.WithStack(err),
		code:  code,
	}
}
