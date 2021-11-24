package tea

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	// DefaultSetGlobalLogLevel is the default log level for the logger
	DefaultSetGlobalLogLevel = zerolog.InfoLevel
)

// logLock provides safe concurrent access for the global logger
var logLock sync.Mutex

// Logger is a global Logger
var logger = NewLogger()

// SetGlobalLogWriter sets the Writer for the global Logger
func SetGlobalLogWriter(w io.Writer) {
	logLock.Lock()
	defer logLock.Unlock()

	l := CurrentLogger()
	l.Writer(w)
}

// SetGlobalLogLevel sets the default log level for the global Logger
func SetGlobalLogLevel(level string) {
	logLock.Lock()
	defer logLock.Unlock()

	l := CurrentLogger()
	l.Level(level)
}

// Debug prints a debug level message in isolation to the writer
func Debug(w io.Writer, v ...interface{}) {
	NewLogger().Writer(w).Level("debug").Debug(fmt.Sprint(v...))
}

// Debugf prints a debug level formatted message in isolation to the writer
func Debugf(w io.Writer, format string, args ...interface{}) {
	NewLogger().Writer(w).Level("debug").Debug(fmt.Sprintf(format, args...))
}

// Log a series of values at a given level
func Log(level string, v ...interface{}) *Logger {
	logLock.Lock()
	defer logLock.Unlock()

	l := CurrentLogger()
	switch strings.ToLower(level) {
	case "debug":
		l.Debug(fmt.Sprint(v...))
	case "info":
		l.Info(fmt.Sprint(v...))
	case "warn":
		l.Warn(fmt.Sprint(v...))
	}

	return l
}

// Logf prints a formatted message at a given level
func Logf(level, format string, args ...interface{}) *Logger {
	return Log(level, fmt.Sprintf(format, args...))
}

// ResetGlobalLogger sets the global logger to default values
func ResetGlobalLogger() {
	logLock.Lock()
	defer logLock.Unlock()

	l := CurrentLogger()
	l.w = NewLogger().w
}

// Logger is an instance of the zerolog based Logger
type Logger struct {
	w zerolog.Logger
}

// Debug sends a debug level message
func (l *Logger) Debug(msg string) *Logger {
	l.w.Debug().Msg(msg)
	return l
}

// Info sends an info level message
func (l *Logger) Info(msg string) *Logger {
	l.w.Info().Msg(msg)
	return l
}

// Warn sends a warning level message
func (l *Logger) Warn(msg string) *Logger {
	l.w.Warn().Msg(msg)
	return l
}

// Error sends a error level message
func (l *Logger) Error(err error) *Logger {
	l.w.Error().Msgf("%+v", err)
	return l
}

// HTTPError sends a http error level message
func (l *Logger) HTTPError(r *http.Request, status int, err error) *Logger {
	l.w.Error().
		Str("method", r.Method).
		Stringer("url", r.URL).
		Int("status", status).
		Msgf("%+v", err)

	return l
}

// Writer sets the io writer for the global logger
func (l *Logger) Writer(w io.Writer) *Logger {
	l.w = l.w.Output(w)
	return l
}

// Level sets the log level for the global logger
func (l *Logger) Level(level string) *Logger {
	switch strings.ToLower(level) {
	case "debug":
		l.w = l.w.Level(zerolog.DebugLevel)
	case "info":
		l.w = l.w.Level(zerolog.InfoLevel)
	case "warn":
		l.w = l.w.Level(zerolog.WarnLevel)
	default:
		l.w = l.w.Level(zerolog.ErrorLevel)
	}

	return l
}

// NewLogger creates a Logger with sane defaults.
func NewLogger() *Logger {
	// https://github.com/rs/zerolog/issues/213
	zerolog.TimeFieldFormat = time.RFC3339Nano
	cw := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339Nano}
	return &Logger{
		w: zerolog.New(cw).With().Timestamp().Logger().Level(DefaultSetGlobalLogLevel),
	}
}

// CurrentLogger returns an instance of the global Logger.
func CurrentLogger() *Logger {
	return logger
}
