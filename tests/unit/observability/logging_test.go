package observability_test

import (
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggingInterface(t *testing.T) {
	t.Run("NewLogger Creation", func(t *testing.T) {
		logger := observability.NewLogger("info", "json")
		require.NotNil(t, logger)
	})

	t.Run("Logger Debug Messages", func(t *testing.T) {
		logger := observability.NewLogger("debug", "console")
		require.NotNil(t, logger)

		// Test debug logging with fields
		fields := map[string]interface{}{
			"key":   "value",
			"count": 42,
		}

		// Should not panic
		assert.NotPanics(t, func() {
			logger.Debug("debug message", fields)
		})

		// Test debug without fields
		assert.NotPanics(t, func() {
			logger.Debug("debug message without fields")
		})
	})

	t.Run("Logger Info Messages", func(t *testing.T) {
		logger := observability.NewLogger("info", "json")
		require.NotNil(t, logger)

		fields := map[string]interface{}{
			"service": "test-service",
			"module":  "logging",
		}

		assert.NotPanics(t, func() {
			logger.Info("info message", fields)
		})

		assert.NotPanics(t, func() {
			logger.Info("info message without fields")
		})
	})

	t.Run("Logger Warning Messages", func(t *testing.T) {
		logger := observability.NewLogger("warn", "console")
		require.NotNil(t, logger)

		fields := map[string]interface{}{
			"warning_type": "deprecated",
			"module":       "test",
		}

		assert.NotPanics(t, func() {
			logger.Warn("warning message", fields)
		})

		assert.NotPanics(t, func() {
			logger.Warn("warning message without fields")
		})
	})

	t.Run("Logger Error Messages", func(t *testing.T) {
		logger := observability.NewLogger("error", "json")
		require.NotNil(t, logger)

		testErr := assert.AnError
		fields := map[string]interface{}{
			"error_code": "E001",
			"context":    "test",
		}

		assert.NotPanics(t, func() {
			logger.Error("error message", testErr, fields)
		})

		assert.NotPanics(t, func() {
			logger.Error("error message without error", nil, fields)
		})

		assert.NotPanics(t, func() {
			logger.Error("error message without fields", testErr)
		})
	})

	t.Run("Logger With Method", func(t *testing.T) {
		logger := observability.NewLogger("info", "json")
		require.NotNil(t, logger)

		fields := map[string]interface{}{
			"component": "test",
			"version":   "1.0.0",
		}

		childLogger := logger.With(fields)
		require.NotNil(t, childLogger)

		// Child logger should work independently
		assert.NotPanics(t, func() {
			childLogger.Info("message from child logger")
		})

		// Test nested With calls
		moreFields := map[string]interface{}{
			"request_id": "123",
		}

		grandchildLogger := childLogger.With(moreFields)
		require.NotNil(t, grandchildLogger)

		assert.NotPanics(t, func() {
			grandchildLogger.Info("message from grandchild logger")
		})
	})

	t.Run("Logger Levels", func(t *testing.T) {
		levels := []string{"debug", "info", "warn", "error"}
		formats := []string{"json", "console"}

		for _, level := range levels {
			for _, format := range formats {
				t.Run(level+"_"+format, func(t *testing.T) {
					logger := observability.NewLogger(level, format)
					require.NotNil(t, logger)

					// Test all methods work regardless of level
					assert.NotPanics(t, func() {
						logger.Debug("debug test")
						logger.Info("info test")
						logger.Warn("warn test")
						logger.Error("error test", nil)
					})
				})
			}
		}
	})

	t.Run("InitLogging Function", func(t *testing.T) {
		// Test different combinations
		testCases := []struct {
			level   string
			format  string
			service string
		}{
			{"debug", "console", "test-service"},
			{"info", "json", "another-service"},
			{"warn", "console", "warn-service"},
			{"error", "json", "error-service"},
			{"invalid", "json", "default-service"}, // Should default to info
		}

		for _, tc := range testCases {
			t.Run(tc.level+"_"+tc.format, func(t *testing.T) {
				assert.NotPanics(t, func() {
					observability.InitLogging(tc.level, tc.format, tc.service)
				})

				// Test that GetLogger works after init
				logger := observability.GetLogger("component")
				assert.NotNil(t, logger)
			})
		}
	})

	t.Run("Utility Logging Functions", func(t *testing.T) {
		// Initialize logging first
		observability.InitLogging("info", "json", "test-service")

		assert.NotPanics(t, func() {
			observability.LogStartup("test-service", "1.0.0", 8080)
		})

		assert.NotPanics(t, func() {
			observability.LogShutdown("test-service", "graceful shutdown")
		})
	})

	t.Run("Field Processing Edge Cases", func(t *testing.T) {
		logger := observability.NewLogger("debug", "json")
		require.NotNil(t, logger)

		// Test with empty fields
		emptyFields := map[string]interface{}{}
		assert.NotPanics(t, func() {
			logger.Info("message with empty fields", emptyFields)
		})

		// Test with nil values in fields
		nilFields := map[string]interface{}{
			"nil_value":  nil,
			"empty_str":  "",
			"zero_int":   0,
			"false_bool": false,
		}
		assert.NotPanics(t, func() {
			logger.Debug("message with nil fields", nilFields)
		})

		// Test with complex nested data
		complexFields := map[string]interface{}{
			"nested": map[string]interface{}{
				"inner": "value",
				"count": 123,
			},
			"array": []string{"item1", "item2"},
		}
		assert.NotPanics(t, func() {
			logger.Warn("message with complex fields", complexFields)
		})
	})
}

func TestLoggerFieldAddition(t *testing.T) {
	logger := observability.NewLogger("debug", "json")
	require.NotNil(t, logger)

	// Test multiple field maps in one call
	fields1 := map[string]interface{}{
		"field1": "value1",
	}
	fields2 := map[string]interface{}{
		"field2": "value2",
	}

	assert.NotPanics(t, func() {
		logger.Info("message with multiple field maps", fields1, fields2)
	})

	// Test with no fields at all
	assert.NotPanics(t, func() {
		logger.Error("message with no fields", nil)
	})
}

func TestLoggerInterfaceCompliance(t *testing.T) {
	logger := observability.NewLogger("info", "json")

	// Verify it implements the Logger interface
	var _ observability.Logger = logger

	// Test that all interface methods are available
	assert.NotPanics(t, func() {
		logger.Debug("debug")
		logger.Info("info")
		logger.Warn("warn")
		logger.Error("error", nil)
		childLogger := logger.With(map[string]interface{}{"key": "value"})
		childLogger.Info("child message")
	})
}
