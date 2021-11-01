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
	"sync"
)

// lock provides safe concurrent access for the global logger
var lock sync.Mutex

// Writer sets the Writer for the global Logger
func Writer(w io.Writer) {
	lock.Lock()
	defer lock.Unlock()

	l := CurrentLogger()
	l.Writer(w)
}

// Level sets the default log level for the global Logger
func Level(level string) {
	lock.Lock()
	defer lock.Unlock()

	l := CurrentLogger()
	l.Level(level)
}

// Debug sends a debug level message
func Debug(v ...interface{}) *Logger {
	lock.Lock()
	defer lock.Unlock()

	l := CurrentLogger()
	return l.Debug(fmt.Sprint(v...))
}

// Debugf sends a formatted debug level message
func Debugf(format string, args ...interface{}) *Logger {
	return Debug(fmt.Sprintf(format, args...))
}

// Info sends an info level message
func Info(v ...interface{}) *Logger {
	lock.Lock()
	defer lock.Unlock()

	l := CurrentLogger()
	return l.Info(fmt.Sprint(v...))
}

// Infof sends a formatted info level message
func Infof(format string, args ...interface{}) *Logger {
	return Info(fmt.Sprintf(format, args...))
}

// Warn sends a warning level message
func Warn(v ...interface{}) *Logger {
	lock.Lock()
	defer lock.Unlock()

	l := CurrentLogger()
	return l.Warn(fmt.Sprint(v...))
}

// Warnf sends a formatted warning level message
func Warnf(format string, args ...interface{}) *Logger {
	return Warn(fmt.Sprintf(format, args...))
}

// Reset sets the global logger to default values
func Reset() {
	lock.Lock()
	defer lock.Unlock()

	l := CurrentLogger()
	l.Writer(NewLogger().w)
}
