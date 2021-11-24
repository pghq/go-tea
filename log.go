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
	// DefaultLogLevel is the default log level for the logger
	DefaultLogLevel = zerolog.InfoLevel
)

// logLock provides safe concurrent access for the global logger
var logLock sync.Mutex

// Logger is a global Logger
var logger = NewLogger()

// LogWriter sets the Writer for the global Logger
func LogWriter(w io.Writer) {
	logLock.Lock()
	defer logLock.Unlock()

	l := CurrentLogger()
	l.Writer(w)
}

// LogLevel sets the default log level for the global Logger
func LogLevel(level string) {
	logLock.Lock()
	defer logLock.Unlock()

	l := CurrentLogger()
	l.Level(level)
}

// Write prints a debug level message in isolation to stdout
func Write(w io.Writer, v ...interface{}){
	NewLogger().Writer(w).Level("debug").Debug(fmt.Sprint(v...))
}

// Writef prints a debug level formatted message in isolation to stdout
func Writef(w io.Writer, format string, args ...interface{}){
	NewLogger().Writer(w).Level("debug").Debug(fmt.Sprintf(format, args...))
}

// Print a series of values at a given level
func Print(level string, v ...interface{}) *Logger {
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

// Printf prints a formatted message at a given level
func Printf(level, format string, args ...interface{}) *Logger {
	return Print(level, fmt.Sprintf(format, args...))
}

// Debug sends a debug level message
func Debug(v ...interface{}) *Logger {
	return Print("debug", v...)
}

// Debugf sends a formatted debug level message
func Debugf(format string, args ...interface{}) *Logger {
	return Printf("debug", format, args...)
}

// Info sends an info level message
func Info(v ...interface{}) *Logger {
	return Print("info", v...)
}

// Infof sends a formatted info level message
func Infof(format string, args ...interface{}) *Logger {
	return Printf("info", format, args...)
}

// Warn sends a warning level message
func Warn(v ...interface{}) *Logger {
	return Print("warn", v...)
}

// Warnf sends a formatted warning level message
func Warnf(format string, args ...interface{}) *Logger {
	return Printf("warn", format, args...)
}

// ResetLog sets the global logger to default values
func ResetLog() {
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
func (l *Logger) Writer(w io.Writer) *Logger{
	l.w = l.w.Output(w)
	return l
}

// Level sets the log level for the global logger
func (l *Logger) Level(level string) *Logger{
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
		w: zerolog.New(cw).With().Timestamp().Logger().Level(DefaultLogLevel),
	}
}

// CurrentLogger returns an instance of the global Logger.
func CurrentLogger() *Logger {
	return logger
}
