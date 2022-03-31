package trail

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
	// globalLogger is the global logger
	globalLogger *logger

	// exit is the function that is called on fatal log
	exit func(int)
)

func init() {
	globalLogger = newLogger()
	exit = os.Exit
}

// Testing configures logging for testing
func Testing() {
	exit = func(int) {}
	globalLogger.zap = zap.NewNop()
}

// SetVerbosity sets the global log level
func SetVerbosity(level string) {
	level = strings.ToLower(level)
	switch strings.ToLower(level) {
	case "debug":
		globalLogger.atom.SetLevel(zap.DebugLevel)
	case "info":
		globalLogger.atom.SetLevel(zap.InfoLevel)
	case "warn":
		globalLogger.atom.SetLevel(zap.WarnLevel)
	case "error":
		globalLogger.atom.SetLevel(zap.ErrorLevel)
	case "fatal":
		globalLogger.atom.SetLevel(zap.FatalLevel)
	default:
		globalLogger.atom.SetLevel(zap.DebugLevel)
		return
	}
}

// log a series of values at a given level
func log(level string, v interface{}) {
	if v == nil {
		return
	}

	level = strings.ToLower(level)
	switch level {
	case "test":
		newLogger().zap.Info(fmt.Sprint(v))
	case "debug":
		globalLogger.zap.Debug(fmt.Sprint(v))
	case "info":
		globalLogger.zap.Info(fmt.Sprint(v))
	case "warn":
		globalLogger.zap.Warn(fmt.Sprint(v))
	case "error", "fatal":
		err, ok := v.(error)
		if !ok {
			err = NewError(fmt.Sprint(v))
		}

		if !IsFatal(err) {
			return
		}

		span := StartSpan(context.Background(), level)
		defer span.Finish()

		if level == "fatal" {
			span.Recover(err)
			exit(1)
		} else {
			span.Capture(err)
		}
	}
}

// Debug prints a message at debug level
func Debug(v interface{}) {
	log("debug", v)
}

// Debugf prints a formatted message at debug level
func Debugf(format string, args ...interface{}) {
	Debug(fmt.Sprintf(format, args...))
}

// Info prints a message at info level
func Info(v interface{}) {
	log("info", v)
}

// Infof prints a formatted message at info level
func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

// Warn prints a message at warn level
func Warn(v interface{}) {
	log("warn", v)
}

// Warnf prints a formatted message at warn level
func Warnf(format string, args ...interface{}) {
	Warn(fmt.Sprintf(format, args...))
}

// Error prints a message at error level
func Error(v interface{}) {
	log("error", v)
}

// Errorf prints a formatted message at error level
func Errorf(format string, args ...interface{}) {
	Error(fmt.Sprintf(format, args...))
}

// Fatal prints a message at fatal level
func Fatal(v interface{}) {
	log("fatal", v)
}

// Fatalf prints a formatted message at fatal level
func Fatalf(format string, args ...interface{}) {
	Fatal(fmt.Sprintf(format, args...))
}

// OneOff prints a message at fatal level
func OneOff(v interface{}) {
	log("test", v)
}

// logger is an instance of the zap based Logger
type logger struct {
	zap  *zap.Logger
	atom zap.AtomicLevel
}

func (l logger) Error(err interface{}) {
	l.zap.Error(fmt.Sprintf("%+v", err))
}

func (l logger) ErrorWithStacktrace(err interface{}) {
	l.zap.Error(fmt.Sprintf("%+v\n%s", err, string(debug.Stack())))
}

func (l logger) Flush() {
	_ = l.zap.Sync()
}

// newLogger creates a Logger with sane defaults.
func newLogger() *logger {
	atom := zap.NewAtomicLevel()
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	config.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapLogger := zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(config), zapcore.Lock(os.Stdout), atom))
	return &logger{
		zap:  zapLogger,
		atom: atom,
	}
}
