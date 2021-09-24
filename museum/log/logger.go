package log

import (
	"io"
	"net/http"
	"os"
	"strings"
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
type Logger struct{
	w zerolog.Logger
}

func (l *Logger) Debug(msg string) *Logger {
	l.w.Debug().Msg(msg)
	return l
}

func (l *Logger) Info(msg string) *Logger {
	l.w.Info().Msg(msg)
	return l
}

func (l *Logger) Warn(msg string) *Logger {
	l.w.Warn().Msg(msg)
	return l
}

func (l *Logger) Error(err error) *Logger {
	l.w.Error().Msgf("%+v", err)
	return l
}

func (l *Logger) HTTPError(r *http.Request, status int, err error) *Logger {
	l.w.Error().
		Str("method", r.Method).
		Stringer("url", r.URL).
		Int("status", status).
		Msgf("%+v", err)

	return l
}

func (l *Logger) Writer(w io.Writer){
	l.w = l.w.Output(w)
}

func (l *Logger) Level(level string){
	switch strings.ToLower(level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	}
}

// NewLogger creates a Logger with sane defaults.
func NewLogger() *Logger{
	return &Logger{
		w: zerolog.New(os.Stderr),
	}
}

// CurrentLogger returns an instance of the global Logger.
func CurrentLogger() *Logger{
	return logger
}

// Init initializes the global logger
func Init(){
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(DefaultLogLevel)
}
