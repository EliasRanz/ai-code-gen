package observability_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Use a sync.Once to ensure metrics are only initialized once across all tests
var (
	metricsInitOnce sync.Once
	metricsInitErr  error
)

func initMetricsOnce() error {
	metricsInitOnce.Do(func() {
		metricsInitErr = observability.InitMetrics("test-service-global")
	})
	return metricsInitErr
}

func TestMetricsInitializationFixed(t *testing.T) {
	t.Run("InitMetrics", func(t *testing.T) {
		err := initMetricsOnce()
		assert.NoError(t, err)

		// Test getting metrics handler
		handler := observability.GetMetricsHandler()
		assert.NotNil(t, handler)

		// Test that handler works
		req := httptest.NewRequest("GET", "/metrics", nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Header().Get("Content-Type"), "text/plain")
	})
}

func TestMetricsRecordingFixed(t *testing.T) {
	// Initialize metrics once
	err := initMetricsOnce()
	require.NoError(t, err)

	serviceName := "metrics-recording-test"

	t.Run("RecordHTTPRequest", func(t *testing.T) {
		// Record some HTTP requests
		observability.RecordHTTPRequest("GET", "/api/users", "200", serviceName, 0.150)
		observability.RecordHTTPRequest("POST", "/api/users", "201", serviceName, 0.250)
		observability.RecordHTTPRequest("GET", "/api/users", "404", serviceName, 0.050)
		observability.RecordHTTPRequest("DELETE", "/api/users/123", "200", serviceName, 0.100)

		// Verify metrics are recorded by checking the metrics endpoint
		handler := observability.GetMetricsHandler()
		req := httptest.NewRequest("GET", "/metrics", nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		metricsOutput := recorder.Body.String()

		// Check that HTTP metrics are present
		assert.Contains(t, metricsOutput, "http_requests_total")
		assert.Contains(t, metricsOutput, "http_request_duration_seconds")

		// Check specific method/status combinations
		assert.Contains(t, metricsOutput, `method="GET"`)
		assert.Contains(t, metricsOutput, `method="POST"`)
		assert.Contains(t, metricsOutput, `method="DELETE"`)
		assert.Contains(t, metricsOutput, `status_code="200"`)
		assert.Contains(t, metricsOutput, `status_code="201"`)
		assert.Contains(t, metricsOutput, `status_code="404"`)
	})

	t.Run("IncrementServiceUptime", func(t *testing.T) {
		// Test uptime increment
		observability.IncrementServiceUptime(serviceName, "1.0.0")
		observability.IncrementServiceUptime(serviceName, "1.0.0")
		observability.IncrementServiceUptime(serviceName, "1.0.0")

		// Check metrics output
		handler := observability.GetMetricsHandler()
		req := httptest.NewRequest("GET", "/metrics", nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		metricsOutput := recorder.Body.String()
		assert.Contains(t, metricsOutput, "service_uptime_seconds")
		assert.Contains(t, metricsOutput, `service="`+serviceName+`"`)
		assert.Contains(t, metricsOutput, `version="1.0.0"`)
	})

	t.Run("SetActiveConnections", func(t *testing.T) {
		// Test setting active connections
		observability.SetActiveConnections(serviceName, 10)
		observability.SetActiveConnections(serviceName, 25)
		observability.SetActiveConnections(serviceName, 15)

		// Check metrics output
		handler := observability.GetMetricsHandler()
		req := httptest.NewRequest("GET", "/metrics", nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		metricsOutput := recorder.Body.String()
		assert.Contains(t, metricsOutput, "active_connections")
		assert.Contains(t, metricsOutput, "15") // Should show the latest value
		assert.Contains(t, metricsOutput, `service="`+serviceName+`"`)
	})
}

func TestMetricsEdgeCasesFixed(t *testing.T) {
	err := initMetricsOnce()
	require.NoError(t, err)

	t.Run("GetMetricsHandler", func(t *testing.T) {
		// Test getting handler
		handler := observability.GetMetricsHandler()
		assert.NotNil(t, handler)
	})

	t.Run("Record Metrics with Edge Cases", func(t *testing.T) {
		serviceName := "edge-case-test"

		// Test with edge case values
		observability.RecordHTTPRequest("", "", "0", serviceName, 0.0)
		observability.RecordHTTPRequest("VERY_LONG_METHOD_NAME", "/very/long/endpoint/path/that/might/cause/issues", "999", serviceName, 9999.999)
		observability.SetActiveConnections(serviceName, 0)
		observability.SetActiveConnections(serviceName, -1) // Negative values

		// Should not panic or error
		handler := observability.GetMetricsHandler()
		req := httptest.NewRequest("GET", "/metrics", nil)
		recorder := httptest.NewRecorder()

		// This should not panic
		assert.NotPanics(t, func() {
			handler.ServeHTTP(recorder, req)
		})
	})

	t.Run("Metrics with Special Characters", func(t *testing.T) {
		serviceName := "special-chars-test"

		// Test with special characters that might cause issues
		observability.RecordHTTPRequest("GET", "/api/users?query=test&filter=active", "200", serviceName, 0.100)
		observability.RecordHTTPRequest("POST", "/api/data with spaces", "201", serviceName, 0.200)
		observability.RecordHTTPRequest("GET", "/api/unicode/测试", "200", serviceName, 0.150)

		// Should handle gracefully
		handler := observability.GetMetricsHandler()
		req := httptest.NewRequest("GET", "/metrics", nil)
		recorder := httptest.NewRecorder()

		assert.NotPanics(t, func() {
			handler.ServeHTTP(recorder, req)
		})

		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestMetricsIntegrationFixed(t *testing.T) {
	err := initMetricsOnce()
	require.NoError(t, err)

	t.Run("Full Metrics Workflow", func(t *testing.T) {
		serviceName := "integration-test"

		// Simulate typical application metrics
		for i := 0; i < 10; i++ {
			observability.RecordHTTPRequest("GET", "/api/health", "200", serviceName, 0.001)
		}

		for i := 0; i < 5; i++ {
			observability.RecordHTTPRequest("POST", "/api/users", "201", serviceName, 0.150)
		}

		for i := 0; i < 2; i++ {
			observability.RecordHTTPRequest("GET", "/api/users/missing", "404", serviceName, 0.050)
		}

		// Update service metrics
		observability.IncrementServiceUptime(serviceName, "2.0.0")
		observability.SetActiveConnections(serviceName, 42)

		// Verify comprehensive metrics output
		handler := observability.GetMetricsHandler()
		req := httptest.NewRequest("GET", "/custom-metrics", nil)
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
		metricsOutput := recorder.Body.String()

		// Verify all our metrics are present
		assert.Contains(t, metricsOutput, "http_requests_total")
		assert.Contains(t, metricsOutput, "http_request_duration_seconds")
		assert.Contains(t, metricsOutput, "service_uptime_seconds")
		assert.Contains(t, metricsOutput, "active_connections")

		// Verify service-specific metrics
		assert.Contains(t, metricsOutput, serviceName)
		assert.Contains(t, metricsOutput, "2.0.0")
		assert.Contains(t, metricsOutput, "42")
	})
}
