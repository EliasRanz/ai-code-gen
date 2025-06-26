// Package observability provides logging, metrics, and tracing infrastructure
package observability

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger interface for dependency injection
type Logger interface {
	Debug(msg string, fields ...map[string]interface{})
	Info(msg string, fields ...map[string]interface{})
	Warn(msg string, fields ...map[string]interface{})
	Error(msg string, err error, fields ...map[string]interface{})
	Fatal(msg string, err error, fields ...map[string]interface{})
	With(fields map[string]interface{}) Logger
}

// ZerologLogger wraps zerolog for structured logging
type ZerologLogger struct {
	logger zerolog.Logger
}

// NewLogger creates a new structured logger
func NewLogger(level string, format string) Logger {
	// Configure log level
	logLevel := zerolog.InfoLevel
	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	// Configure output format
	var logger zerolog.Logger
	if format == "console" {
		logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	} else {
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}

	return &ZerologLogger{logger: logger}
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

// RequestLogger middleware for HTTP request logging
type RequestLogger struct {
	logger Logger
}

// NewRequestLogger creates a new request logger middleware
func NewRequestLogger(logger Logger) *RequestLogger {
	return &RequestLogger{logger: logger}
}

// LogRequest logs HTTP request details
func (rl *RequestLogger) LogRequest(ctx context.Context, method, path string, statusCode int, duration time.Duration) {
	fields := map[string]interface{}{
		"method":      method,
		"path":        path,
		"status_code": statusCode,
		"duration_ms": duration.Milliseconds(),
	}

	if statusCode >= 500 {
		rl.logger.Error("HTTP request failed", nil, fields)
	} else if statusCode >= 400 {
		rl.logger.Warn("HTTP request warning", fields)
	} else {
		rl.logger.Info("HTTP request completed", fields)
	}
}

// ContextLogger adds request-scoped logging context
type ContextLogger struct {
	logger Logger
}

// NewContextLogger creates a new context logger
func NewContextLogger(logger Logger) *ContextLogger {
	return &ContextLogger{logger: logger}
}

// FromContext retrieves logger from context
func (cl *ContextLogger) FromContext(ctx context.Context) Logger {
	// In a real implementation, you would extract correlation ID, user ID, etc. from context
	// For now, return the base logger
	return cl.logger
}

// WithContext adds logger to context
func (cl *ContextLogger) WithContext(ctx context.Context, fields map[string]interface{}) context.Context {
	// In a real implementation, you would add the logger with fields to context
	// For now, return the context as-is
	return ctx
}

// MetricsCollector interface for metrics collection
type MetricsCollector interface {
	IncrementCounter(name string, tags map[string]string)
	RecordHistogram(name string, value float64, tags map[string]string)
	RecordGauge(name string, value float64, tags map[string]string)
}

// NoOpMetricsCollector is a no-op implementation of MetricsCollector
type NoOpMetricsCollector struct{}

// NewNoOpMetricsCollector creates a new no-op metrics collector
func NewNoOpMetricsCollector() MetricsCollector {
	return &NoOpMetricsCollector{}
}

// IncrementCounter does nothing
func (m *NoOpMetricsCollector) IncrementCounter(name string, tags map[string]string) {}

// RecordHistogram does nothing
func (m *NoOpMetricsCollector) RecordHistogram(name string, value float64, tags map[string]string) {}

// RecordGauge does nothing
func (m *NoOpMetricsCollector) RecordGauge(name string, value float64, tags map[string]string) {}
