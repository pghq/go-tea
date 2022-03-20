package tea

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// logger is the global logger
	logger *Logger

	// exit is the function that is called on fatal log
	exit func(int)
)

func init() {
	logger = NewLogger()
	exit = os.Exit
}

// Testing configures logging for testing
func Testing() {
	exit = func(int) {}
	logger.zap = zap.NewNop()
}

// SetVerbosity sets the global log level
func SetVerbosity(level string) {
	level = strings.ToLower(level)
	switch strings.ToLower(level) {
	case "debug":
		logger.atom.SetLevel(zap.DebugLevel)
	case "info":
		logger.atom.SetLevel(zap.InfoLevel)
	case "warn":
		logger.atom.SetLevel(zap.WarnLevel)
	case "error":
		logger.atom.SetLevel(zap.ErrorLevel)
	case "fatal":
		logger.atom.SetLevel(zap.FatalLevel)
	default:
		logger.atom.SetLevel(zap.DebugLevel)
		return
	}
}

// Log a series of values at a given level
func Log(ctx context.Context, level string, v ...interface{}) {
	if ctx.Err() != nil || len(v) == 0 || v[0] == nil {
		return
	}

	level = strings.ToLower(level)
	switch level {
	case "test":
		logger := NewLogger()
		logger.zap.Info(fmt.Sprint(v...))
	case "debug":
		logger.zap.Debug(fmt.Sprint(v...))
	case "info":
		logger.zap.Info(fmt.Sprint(v...))
	case "warn":
		logger.zap.Warn(fmt.Sprint(v...))
	case "error", "fatal":
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

		span := Nest(ctx, level)
		defer span.End()
		span.Capture(err)

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
	zap  *zap.Logger
	atom zap.AtomicLevel
}

func (l Logger) Error(err interface{}) {
	l.zap.Error(fmt.Sprintf("%+v", err))
}

func (l Logger) ErrorWithStacktrace(err interface{}) {
	l.zap.Error(fmt.Sprintf("%+v\n%s", err, string(debug.Stack())))
}

func (l Logger) Tag(k, v string) {
	l.zap = l.zap.With(zap.String(k, v))
}

func (l Logger) Flush() {
	_ = l.zap.Sync()
}

// NewLogger creates a Logger with sane defaults.
func NewLogger() *Logger {
	atom := zap.NewAtomicLevel()
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapLogger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(config), zapcore.Lock(os.Stdout), atom))
	return &Logger{
		zap:  zapLogger,
		atom: atom,
	}
}
