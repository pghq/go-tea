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

// Package logger provides a global logger.
package log

import (
	"fmt"
	"io"
	"net/http"
)

// Writer sets the Writer for the global Logger
func Writer(w io.Writer) {
	l := CurrentLogger()
	l.Writer(w)
}

// Level sets the default log level for the global Logger
func Level(level string){
	l := CurrentLogger()
	l.Level(level)
}

// Debug sends a debug level message
func Debug(msg string) *Logger {
	l := CurrentLogger()
	return l.Debug(msg)
}

// Debugf sends a formatted debug level message
func Debugf(format string, args ...interface{}) *Logger {
	return Debug(fmt.Sprintf(format, args...))
}

// Info sends an info level message
func Info(msg string) *Logger {
	l := CurrentLogger()
	return l.Info(msg)
}

// Infof sends a formatted info level message
func Infof(format string, args ...interface{}) *Logger {
	return Info(fmt.Sprintf(format, args...))
}

// Warn sends a warning level message
func Warn(msg string) *Logger {
	l := CurrentLogger()
	return l.Warn(msg)
}

// Warnf sends a formatted warning level message
func Warnf(format string, args ...interface{}) *Logger {
	return Warn(fmt.Sprintf(format, args...))
}

// Error sends a error level message
func Error(err error) *Logger {
	l := CurrentLogger()
	return l.Error(err)
}

// HTTPError sends a http error level message
func HTTPError(r *http.Request, status int, err error) *Logger {
	l := CurrentLogger()
	return l.HTTPError(r, status, err)
}