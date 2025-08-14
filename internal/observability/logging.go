package observability

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger interface for dependency injection (compatible with infrastructure/observability)
type Logger interface {
	Debug(msg string, fields ...map[string]interface{})
	Info(msg string, fields ...map[string]interface{})
	Warn(msg string, fields ...map[string]interface{})
	Error(msg string, err error, fields ...map[string]interface{})
	Fatal(msg string, err error, fields ...map[string]interface{})
	With(fields map[string]interface{}) Logger
}

// GlobalLogger provides structured logging
var GlobalLogger zerolog.Logger

// InitLogging initializes the logging system
func InitLogging(level string, format string, serviceName string) {
	// Set log level
	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Configure output format
	if format == "console" {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		GlobalLogger = zerolog.New(output).With().
			Timestamp().
			Str("service", serviceName).
			Logger()
	} else {
		GlobalLogger = zerolog.New(os.Stdout).With().
			Timestamp().
			Str("service", serviceName).
			Logger()
	}
}

// GetLogger returns a logger with additional context
func GetLogger(component string) zerolog.Logger {
	return GlobalLogger.With().Str("component", component).Logger()
}

// ZerologLogger wraps zerolog for structured logging (compatible with infrastructure interface)
type ZerologLogger struct {
	logger zerolog.Logger
}

// NewLogger creates a new structured logger compatible with infrastructure/observability interface
func NewLogger(level string, format string) Logger {
	InitLogging(level, format, "service")
	return &ZerologLogger{logger: GlobalLogger}
}

// Debug logs a debug message
func (l *ZerologLogger) Debug(msg string, fields ...map[string]interface{}) {
	event := l.logger.Debug()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Info logs an info message
func (l *ZerologLogger) Info(msg string, fields ...map[string]interface{}) {
	event := l.logger.Info()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Warn logs a warning message
func (l *ZerologLogger) Warn(msg string, fields ...map[string]interface{}) {
	event := l.logger.Warn()
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Error logs an error message
func (l *ZerologLogger) Error(msg string, err error, fields ...map[string]interface{}) {
	event := l.logger.Error()
	if err != nil {
		event = event.Err(err)
	}
	l.addFields(event, fields...)
	event.Msg(msg)
}

// Fatal logs a fatal message and exits
func (l *ZerologLogger) Fatal(msg string, err error, fields ...map[string]interface{}) {
	event := l.logger.Fatal()
	if err != nil {
		event = event.Err(err)
	}
	l.addFields(event, fields...)
	event.Msg(msg)
}

// With returns a new logger with additional fields
func (l *ZerologLogger) With(fields map[string]interface{}) Logger {
	ctx := l.logger.With()
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}
	return &ZerologLogger{logger: ctx.Logger()}
}

// addFields adds fields to a zerolog event
func (l *ZerologLogger) addFields(event *zerolog.Event, fields ...map[string]interface{}) {
	for _, fieldMap := range fields {
		for key, value := range fieldMap {
			event.Interface(key, value)
		}
	}
}

// LogStartup logs service startup information
func LogStartup(serviceName string, version string, port int) {
	GlobalLogger.Info().
		Str("service", serviceName).
		Str("version", version).
		Int("port", port).
		Msg("Service starting up")
}

// LogShutdown logs service shutdown information
func LogShutdown(serviceName string, reason string) {
	GlobalLogger.Info().
		Str("service", serviceName).
		Str("reason", reason).
		Msg("Service shutting down")
}
