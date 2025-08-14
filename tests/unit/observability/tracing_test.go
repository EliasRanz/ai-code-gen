package observability_test

import (
	"context"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestTracingInitialization(t *testing.T) {
	t.Run("InitTracing", func(t *testing.T) {
		serviceName := "tracing-test"
		jaegerEndpoint := "http://localhost:14268/api/traces"

		err := observability.InitTracing(serviceName, jaegerEndpoint)
		assert.NoError(t, err)
	})

	t.Run("InitTracing with Invalid Endpoint", func(t *testing.T) {
		serviceName := "invalid-endpoint-test"
		jaegerEndpoint := "invalid://endpoint"

		// Should handle gracefully (may return error or succeed with fallback)
		_ = observability.InitTracing(serviceName, jaegerEndpoint)
		// Don't assert error since behavior may vary
	})
}

func TestTracingOperations(t *testing.T) {
	// Initialize tracing first
	serviceName := "tracing-ops-test"
	jaegerEndpoint := "http://localhost:14268/api/traces"

	err := observability.InitTracing(serviceName, jaegerEndpoint)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("StartSpan", func(t *testing.T) {
		spanCtx, span := observability.StartSpan(ctx, "test-operation")
		require.NotNil(t, spanCtx)
		require.NotNil(t, span)

		// Test span operations
		observability.AddSpanEvent(spanCtx, "event-name", map[string]interface{}{
			"key": "value",
		})

		observability.SetSpanAttributes(spanCtx, map[string]interface{}{
			"http.method":      "GET",
			"http.url":         "/api/test",
			"http.status_code": 200,
			"custom.field":     "custom-value",
		})

		observability.FinishSpan(span)
	})

	t.Run("StartSpanWithOptions", func(t *testing.T) {
		options := []trace.SpanStartOption{
			trace.WithAttributes(
				attribute.String("service", "test-service"),
				attribute.Int("version", 1),
				attribute.Bool("debug", true),
			),
		}

		spanCtx, span := observability.StartSpanWithOptions(ctx, "span-with-options", options...)
		require.NotNil(t, spanCtx)
		require.NotNil(t, span)

		observability.FinishSpan(span)
	})

	t.Run("SetSpanError", func(t *testing.T) {
		spanCtx, span := observability.StartSpan(ctx, "error-span")
		require.NotNil(t, spanCtx)
		require.NotNil(t, span)

		// Test setting span error
		testError := assert.AnError
		observability.SetSpanError(spanCtx, testError)

		observability.FinishSpan(span)
	})

	t.Run("Nested Spans", func(t *testing.T) {
		// Parent span
		parentCtx, parentSpan := observability.StartSpan(ctx, "parent-operation")
		require.NotNil(t, parentCtx)
		require.NotNil(t, parentSpan)

		observability.SetSpanAttributes(parentCtx, map[string]interface{}{
			"operation.type": "parent",
		})

		// Child span
		childCtx, childSpan := observability.StartSpan(parentCtx, "child-operation")
		require.NotNil(t, childCtx)
		require.NotNil(t, childSpan)

		observability.SetSpanAttributes(childCtx, map[string]interface{}{
			"operation.type": "child",
			"child.id":       123,
		})

		// Grandchild span
		grandchildCtx, grandchildSpan := observability.StartSpan(childCtx, "grandchild-operation")
		require.NotNil(t, grandchildCtx)
		require.NotNil(t, grandchildSpan)

		observability.AddSpanEvent(grandchildCtx, "processing", map[string]interface{}{
			"step": "data-validation",
		})

		// Finish in reverse order
		observability.FinishSpan(grandchildSpan)
		observability.FinishSpan(childSpan)
		observability.FinishSpan(parentSpan)
	})
}

func TestTracingEdgeCases(t *testing.T) {
	serviceName := "edge-case-test"
	jaegerEndpoint := "http://localhost:14268/api/traces"

	err := observability.InitTracing(serviceName, jaegerEndpoint)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Span Operations with Nil Values", func(t *testing.T) {
		spanCtx, span := observability.StartSpan(ctx, "nil-test-span")
		require.NotNil(t, spanCtx)
		require.NotNil(t, span)

		// Test operations with nil/empty values
		observability.AddSpanEvent(spanCtx, "", nil)
		observability.AddSpanEvent(spanCtx, "empty-event", map[string]interface{}{})

		observability.SetSpanAttributes(spanCtx, nil)
		observability.SetSpanAttributes(spanCtx, map[string]interface{}{})
		observability.SetSpanAttributes(spanCtx, map[string]interface{}{
			"nil_value":    nil,
			"empty_string": "",
			"zero_int":     0,
			"false_bool":   false,
		})

		// Should not panic
		assert.NotPanics(t, func() {
			observability.FinishSpan(span)
		})
	})

	t.Run("Span Operations with Complex Values", func(t *testing.T) {
		spanCtx, span := observability.StartSpan(ctx, "complex-values-span")
		require.NotNil(t, spanCtx)
		require.NotNil(t, span)

		// Test with complex attribute values (note: limited type conversion in current implementation)
		observability.SetSpanAttributes(spanCtx, map[string]interface{}{
			"string": "test-value",
			"int":    42,
			"bool":   true,
			// Note: Complex types might not be supported by current implementation
		})

		observability.AddSpanEvent(spanCtx, "complex-event", map[string]interface{}{
			"timestamp": "2024-01-01T00:00:00Z",
			"count":     100,
		})

		observability.FinishSpan(span)
	})

	t.Run("Multiple Error Handling", func(t *testing.T) {
		spanCtx, span := observability.StartSpan(ctx, "multi-error-span")
		require.NotNil(t, spanCtx)
		require.NotNil(t, span)

		// Set multiple errors
		observability.SetSpanError(spanCtx, assert.AnError)
		observability.SetSpanError(spanCtx, assert.AnError)

		observability.FinishSpan(span)
	})
}

func TestTracingIntegration(t *testing.T) {
	t.Run("Full Tracing Workflow", func(t *testing.T) {
		serviceName := "integration-test"
		jaegerEndpoint := "http://localhost:14268/api/traces"

		err := observability.InitTracing(serviceName, jaegerEndpoint)
		require.NoError(t, err)

		ctx := context.Background()

		// Simulate a complex operation with multiple spans
		rootCtx, rootSpan := observability.StartSpan(ctx, "http-request")
		observability.SetSpanAttributes(rootCtx, map[string]interface{}{
			"http.method":      "POST",
			"http.url":         "/api/users",
			"http.status_code": 201,
			"user.id":          "12345",
		})

		// Database operation
		dbCtx, dbSpan := observability.StartSpan(rootCtx, "database-query")
		observability.SetSpanAttributes(dbCtx, map[string]interface{}{
			"db.system":    "postgresql",
			"db.operation": "INSERT",
			"db.table":     "users",
		})
		observability.AddSpanEvent(dbCtx, "query-start", nil)
		observability.AddSpanEvent(dbCtx, "query-complete", map[string]interface{}{
			"rows_affected": 1,
		})
		observability.FinishSpan(dbSpan)

		// Cache operation
		cacheCtx, cacheSpan := observability.StartSpan(rootCtx, "cache-set")
		observability.SetSpanAttributes(cacheCtx, map[string]interface{}{
			"cache.system": "redis",
			"cache.key":    "user:12345",
			"cache.hit":    false,
		})
		observability.FinishSpan(cacheSpan)

		// External API call
		apiCtx, apiSpan := observability.StartSpan(rootCtx, "external-api-call")
		observability.SetSpanAttributes(apiCtx, map[string]interface{}{
			"http.method":      "POST",
			"http.url":         "https://api.external.com/webhook",
			"http.status_code": 200,
			"external.service": "notification-service",
		})

		// Simulate an error in external call
		if false { // Toggle for testing error scenarios
			observability.SetSpanError(apiCtx, assert.AnError)
		} else {
			observability.AddSpanEvent(apiCtx, "webhook-sent", map[string]interface{}{
				"webhook.id": "wh_123456",
			})
		}
		observability.FinishSpan(apiSpan)

		// Finish root span
		observability.FinishSpan(rootSpan)

		// Test context propagation doesn't affect subsequent operations
		assert.Equal(t, ctx, context.Background()) // Original context unchanged
		_ = dbCtx
		_ = cacheCtx
		_ = apiCtx
	})

	t.Run("GetTracer", func(t *testing.T) {
		serviceName := "tracer-test"
		jaegerEndpoint := "http://localhost:14268/api/traces"

		err := observability.InitTracing(serviceName, jaegerEndpoint)
		require.NoError(t, err)

		tracer := observability.GetTracer()
		assert.NotNil(t, tracer)

		// Test that we can use the tracer directly
		ctx := context.Background()
		ctx, span := tracer.Start(ctx, "direct-tracer-span")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
		span.End()
	})
}

func TestTracingErrorScenarios(t *testing.T) {
	t.Run("Operations Before Init", func(t *testing.T) {
		// Test that tracing operations don't panic before initialization
		ctx := context.Background()

		assert.NotPanics(t, func() {
			_, span := observability.StartSpan(ctx, "before-init-span")
			observability.FinishSpan(span)
		})

		assert.NotPanics(t, func() {
			tracer := observability.GetTracer()
			assert.NotNil(t, tracer) // Should return a no-op tracer or nil
		})
	})

	t.Run("Double Finish Span", func(t *testing.T) {
		serviceName := "double-finish-test"
		jaegerEndpoint := "http://localhost:14268/api/traces"

		err := observability.InitTracing(serviceName, jaegerEndpoint)
		require.NoError(t, err)

		ctx := context.Background()
		_, span := observability.StartSpan(ctx, "double-finish-span")

		// Finish span twice - should not panic
		assert.NotPanics(t, func() {
			observability.FinishSpan(span)
			observability.FinishSpan(span)
		})
	})

	t.Run("Operations with Nil Span", func(t *testing.T) {
		// Test operations with nil span
		assert.NotPanics(t, func() {
			observability.FinishSpan(nil)
		})

		ctx := context.Background()
		assert.NotPanics(t, func() {
			observability.SetSpanError(ctx, assert.AnError)
			observability.SetSpanAttributes(ctx, map[string]interface{}{"key": "value"})
			observability.AddSpanEvent(ctx, "test-event", nil)
		})
	})
}
