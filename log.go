package tea

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var (
	// logger is the global logger
	logger Logger

	// exit is the function that is called on fatal log
	exit func(int)
)

func init() {
	logger = defaultLogger()
	exit = os.Exit
}

// Testing configures logging for testing
func Testing() {
	exit = func(int) {}
	logger.zerolog = logger.zerolog.Output(io.Discard)
}

// Verbosity gets the global log verbosity
func Verbosity() string {
	return logger.level
}

// SetVerbosity sets the global log level
func SetVerbosity(level string) {
	level = strings.ToLower(level)
	switch strings.ToLower(level) {
	case "trace":
		logger.zerolog = logger.zerolog.Level(zerolog.TraceLevel)
	case "debug":
		logger.zerolog = logger.zerolog.Level(zerolog.DebugLevel)
	case "info":
		logger.zerolog = logger.zerolog.Level(zerolog.InfoLevel)
	case "warn":
		logger.zerolog = logger.zerolog.Level(zerolog.WarnLevel)
	case "error":
		logger.zerolog = logger.zerolog.Level(zerolog.ErrorLevel)
	case "fatal":
		logger.zerolog = logger.zerolog.Level(zerolog.FatalLevel)
	default:
		logger.zerolog = logger.zerolog.Level(zerolog.NoLevel)
		return
	}

	logger.level = level
}

// Log a series of values at a given level
func Log(ctx context.Context, level string, v ...interface{}) {
	level = strings.ToLower(level)
	switch level {
	case "test":
		z := defaultLogger().zerolog
		z.Print(fmt.Sprint(v...))
	case "debug":
		logger.zerolog.Debug().Msg(fmt.Sprint(v...))
	case "info":
		logger.zerolog.Info().Msg(fmt.Sprint(v...))
	case "warn":
		logger.zerolog.Warn().Msg(fmt.Sprint(v...))
	case "error", "fatal", "trace":
		var err error
		if len(v) == 1 {
			err, _ = v[0].(error)
		}

		if err == nil {
			err = Err(v...)
		}

		if !IsFatal(err) {
			return
		}

		span := Start(ctx, level)
		defer span.End()
		span.Capture(err)
		logger.zerolog.Error().Msg(fmt.Sprint(v...))
		if level == "fatal" {
			Flush()
			exit(1)
		}
	}
}

// Logf prints a formatted message at a given level
func Logf(ctx context.Context, level, format string, args ...interface{}) {
	Log(ctx, level, fmt.Sprintf(format, args...))
}

// Logger is an instance of the zerolog based Logger
type Logger struct {
	level   string
	zerolog zerolog.Logger
}

// defaultLogger creates a Logger with sane defaults.
func defaultLogger() Logger {
	// https://github.com/rs/zerolog/issues/213
	zerolog.TimeFieldFormat = time.RFC3339Nano
	cw := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339Nano}
	return Logger{
		zerolog: zerolog.New(cw).With().Timestamp().Logger().Level(zerolog.TraceLevel),
		level:   "trace",
	}
}
