// Copyright 2021 PGHQ. All Rights Reserved.
//
// Licensed under the GNU General Public License, Version 3 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package errors provides resources for dealing with errors.
package errors

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"

	"github.com/pghq/go-museum/museum/diagnostic/log"
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

// Wrap creates an internal error
func Wrap(err error) error {
	if Is(err, context.Canceled) {
		return runtimeError(err, http.StatusBadRequest)
	}

	if Is(err, context.DeadlineExceeded) {
		return runtimeError(err, http.StatusRequestTimeout)
	}

	if ae, ok := err.(*applicationError); ok {
		return runtimeError(err, ae.code)
	}

	return runtimeError(err, http.StatusInternalServerError)
}

// New creates an internal error from an error message
func New(v ...interface{}) error {
	return Wrap(errors.New(fmt.Sprint(v...)))
}

// Newf creates an internal error from a formatted error message
func Newf(format string, v ...interface{}) error {
	return Wrap(errors.New(fmt.Sprintf(format, v...)))
}

// HTTP creates a http error
func HTTP(code int, err error) error {
	return runtimeError(err, code)
}

// NewHTTP creates a http error from a msg
func NewHTTP(code int, v ...interface{}) error {
	return HTTP(code, New(v...))
}

// BadRequest creates a bad request error
func BadRequest(err error) error {
	return HTTP(http.StatusBadRequest, err)
}

// NewBadRequest creates a bad request error from a msg
func NewBadRequest(v ...interface{}) error {
	return NewHTTP(http.StatusBadRequest, v...)
}

// NoContent creates a no content error
func NoContent(err error) error {
	return HTTP(http.StatusNoContent, err)
}

// NewNoContent creates a no content error from a msg
func NewNoContent(v ...interface{}) error {
	return NewHTTP(http.StatusNoContent, v...)
}

// As finds first error in chain matching target.
func As(err error, target interface{}) bool {
	// why would we ever want to panic on this call???
	defer func() {
		recover()
	}()

	return errors.As(err, target)
}

// Is finds first error in chain matching target.
func Is(err error, target error) bool {
	return errors.Is(err, target)
}

// IsFatal checks if an error is a fatal application error
func IsFatal(err error) bool {
	return StatusCode(err) >= http.StatusInternalServerError
}

// StatusCode gets an HTTP error code from an error
func StatusCode(err error) int {
	if Is(err, context.Canceled) {
		return http.StatusBadRequest
	}

	if Is(err, context.DeadlineExceeded) {
		return http.StatusRequestTimeout
	}

	if e, ok := err.(*applicationError); ok {
		return e.code
	}

	return http.StatusInternalServerError
}

// Send emits fatal errors to global log and monitor.
func Send(err error) {
	if !IsFatal(err) {
		return
	}

	l := log.CurrentLogger()
	l.Error(err)
	monitor.Send(err)
}

// SendHTTP replies to the request with an error
// and emits fatal http errors to global log and monitor.
func SendHTTP(w http.ResponseWriter, r *http.Request, err error) {
	msg := err.Error()
	status := StatusCode(err)
	if IsFatal(err) {
		m := CurrentMonitor()
		l := log.CurrentLogger()
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

// runtimeError creates an error with a given code and stack trace.
func runtimeError(err error, code int) error {
	ae := &applicationError{
		cause: errors.WithStack(err),
		code:  code,
	}

	return ae
}
