package log

import (
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	// DefaultLogLevel is the default log level for the logger
	DefaultLogLevel = zerolog.InfoLevel
)

// Logger is a global Logger
var logger = NewLogger()

// Logger is an instance of the zerolog based Logger
type Logger struct {
	w    zerolog.Logger
	lock sync.RWMutex
}

// Debug sends a debug level message
func (l *Logger) Debug(msg string) *Logger {
	l.lock.RLock()
	defer l.lock.RUnlock()

	l.w.Debug().Msg(msg)
	return l
}

// Info sends an info level message
func (l *Logger) Info(msg string) *Logger {
	l.lock.RLock()
	defer l.lock.RUnlock()

	l.w.Info().Msg(msg)
	return l
}

// Warn sends a warning level message
func (l *Logger) Warn(msg string) *Logger {
	l.lock.RLock()
	defer l.lock.RUnlock()

	l.w.Warn().Msg(msg)
	return l
}

// Error sends a error level message
func (l *Logger) Error(err error) *Logger {
	l.lock.RLock()
	defer l.lock.RUnlock()

	l.w.Error().Msgf("%+v", err)
	return l
}

// HTTPError sends a http error level message
func (l *Logger) HTTPError(r *http.Request, status int, err error) *Logger {
	l.lock.RLock()
	defer l.lock.RUnlock()

	l.w.Error().
		Str("method", r.Method).
		Stringer("url", r.URL).
		Int("status", status).
		Msgf("%+v", err)

	return l
}

// Writer sets the io writer for the global logger
func (l *Logger) Writer(w io.Writer) {
	l.lock.Lock()
	defer l.lock.Unlock()

	l.w = l.w.Output(w)
}

// Level sets the log level for the global logger
func (l *Logger) Level(level string) {
	l.lock.Lock()
	defer l.lock.Unlock()

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
}

// NewLogger creates a Logger with sane defaults.
func NewLogger() *Logger {
	// https://github.com/rs/zerolog/issues/213
	zerolog.TimeFieldFormat = time.RFC3339Nano
	cw := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339Nano}
	return &Logger{
		w: zerolog.New(cw).With().Timestamp().Logger().Level(DefaultLogLevel),
	}
}

// Reset sets the global logger to default values
func Reset() {
	l := CurrentLogger()
	l.lock.Lock()
	defer l.lock.Unlock()

	logger = NewLogger()
}

// CurrentLogger returns an instance of the global Logger.
func CurrentLogger() *Logger {
	logger.lock.RLock()
	defer logger.lock.RUnlock()

	return logger
}
