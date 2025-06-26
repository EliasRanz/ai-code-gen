package observability

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger provides structured logging
var Logger zerolog.Logger

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
		Logger = zerolog.New(output).With().
			Timestamp().
			Str("service", serviceName).
			Logger()
	} else {
		Logger = zerolog.New(os.Stdout).With().
			Timestamp().
			Str("service", serviceName).
			Logger()
	}
}

// GetLogger returns a logger with additional context
func GetLogger(component string) zerolog.Logger {
	return Logger.With().Str("component", component).Logger()
}

// GetLoggerWithContext returns a logger with custom context
func GetLoggerWithContext(fields map[string]interface{}) zerolog.Logger {
	logger := Logger
	for key, value := range fields {
		logger = logger.With().Interface(key, value).Logger()
	}
	return logger
}

// LogRequest logs HTTP request information
func LogRequest(method, path, userAgent string, statusCode int, duration time.Duration) {
	Logger.Info().
		Str("method", method).
		Str("path", path).
		Str("user_agent", userAgent).
		Int("status_code", statusCode).
		Dur("duration", duration).
		Msg("HTTP request processed")
}

// LogError logs error information with context
func LogError(err error, context string, fields map[string]interface{}) {
	event := Logger.Error().Err(err).Str("context", context)

	for key, value := range fields {
		event = event.Interface(key, value)
	}

	event.Msg("Error occurred")
}

// LogStartup logs service startup information
func LogStartup(serviceName string, version string, port int) {
	Logger.Info().
		Str("service", serviceName).
		Str("version", version).
		Int("port", port).
		Msg("Service starting up")
}

// LogShutdown logs service shutdown information
func LogShutdown(serviceName string, reason string) {
	Logger.Info().
		Str("service", serviceName).
		Str("reason", reason).
		Msg("Service shutting down")
}
