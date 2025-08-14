package observability_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObservabilityInterfaces(t *testing.T) {
	t.Run("Prometheus Provider", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
			MetricsEnabled: true,
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)

		// Test metric recording
		labels := map[string]string{
			"service": "test",
			"env":     "test",
		}

		err = provider.IncrementCounter("test_counter", labels)
		assert.NoError(t, err)

		err = provider.RecordHistogram("test_histogram", 1.5, labels)
		assert.NoError(t, err)

		err = provider.RecordGauge("test_gauge", 42.0, labels)
		assert.NoError(t, err)

		// Test health check
		healthCheck := func(ctx context.Context) error {
			return nil
		}
		err = provider.RegisterHealthCheck("test_check", healthCheck)
		assert.NoError(t, err)

		status := provider.GetHealthStatus()
		assert.Equal(t, "healthy", status.Status)
		assert.Contains(t, status.Checks, "test_check")
		assert.Equal(t, "healthy", status.Checks["test_check"])
	})

	t.Run("OpenTelemetry Provider", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:     "test-service",
			ServiceVersion:  "1.0.0",
			TracingEnabled:  false, // Disable to avoid external dependencies
			TracingEndpoint: "",
		}

		provider, err := observability.NewOpenTelemetryProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)

		// Test span creation (should return no-op when tracing disabled)
		ctx := context.Background()
		ctx, span := provider.StartSpan(ctx, "test_operation")
		assert.NotNil(t, span)
		span.End()

		tracer := provider.CreateTracer("test-service")
		assert.NotNil(t, tracer)
	})

	t.Run("Multi Provider", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
		}

		prometheus, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		otel, err := observability.NewOpenTelemetryProvider(config)
		require.NoError(t, err)

		multiProvider := observability.NewMultiProvider([]observability.ObservabilityProvider{prometheus, otel})
		require.NotNil(t, multiProvider)

		// Test that operations work with multi-provider
		labels := map[string]string{
			"service": "test",
		}

		err = multiProvider.IncrementCounter("multi_counter", labels)
		assert.NoError(t, err)

		err = multiProvider.RecordHistogram("multi_histogram", 2.5, labels)
		assert.NoError(t, err)
	})

	t.Run("Observability Factory", func(t *testing.T) {
		factory := observability.NewObservabilityFactory()
		require.NotNil(t, factory)

		// Test available providers
		providers := factory.ListAvailableProviders()
		assert.Contains(t, providers, "prometheus")
		assert.Contains(t, providers, "opentelemetry")

		// Test creating provider
		config := observability.ObservabilityConfig{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
		}

		provider, err := factory.CreateProvider("prometheus", config)
		require.NoError(t, err)
		assert.NotNil(t, provider)

		// Test unknown provider
		_, err = factory.CreateProvider("unknown", config)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown provider type")
	})

	t.Run("Service Observability", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		serviceObs := observability.NewServiceObservability(provider, "test-service")
		require.NotNil(t, serviceObs)

		assert.Equal(t, provider, serviceObs.GetProvider())
		assert.NotNil(t, serviceObs.GetTracer())
		assert.NotNil(t, serviceObs.GetMetrics())

		// Test recording request
		err = serviceObs.RecordRequest("GET", "/api/test", 150*time.Millisecond, 200)
		assert.NoError(t, err)

		// Test starting trace
		ctx := context.Background()
		ctx, span := serviceObs.StartTrace(ctx, "test_operation")
		assert.NotNil(t, span)
		span.End()
	})
}

func TestMonitoringDecorators(t *testing.T) {
	t.Run("Generic Monitoring Decorator", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		serviceObs := observability.NewServiceObservability(provider, "test-service")
		monitoringConfig := observability.MonitoringConfig{
			Enabled:        true,
			MetricsEnabled: true,
			TracingEnabled: true,
		}

		decorator := observability.NewGenericMonitoringDecorator(
			"test-component",
			serviceObs,
			"repository",
			monitoringConfig,
		)

		require.NotNil(t, decorator)
		assert.Equal(t, "repository", decorator.GetComponentType())

		// Test wrapping a method
		ctx := context.Background()
		result, err := decorator.WrapMethod(ctx, "testMethod", func() (interface{}, error) {
			return "success", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "success", result)

		// Test wrapping a void method
		err = decorator.WrapVoidMethod(ctx, "testVoidMethod", func() error {
			return nil
		})

		assert.NoError(t, err)

		// Test error handling
		_, err = decorator.WrapMethod(ctx, "errorMethod", func() (interface{}, error) {
			return nil, fmt.Errorf("test error")
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test error")

		// Test metrics
		metrics := decorator.GetMetrics()
		assert.NotNil(t, metrics)

		// Should have recorded operations
		assert.Greater(t, metrics.GetOperationCount("repository.testMethod"), int64(0))
		assert.Greater(t, metrics.GetOperationCount("repository.testVoidMethod"), int64(0))
		assert.Greater(t, metrics.GetErrorCount("repository.errorMethod"), int64(0))
	})

	t.Run("Monitoring Decorator Factory", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		serviceObs := observability.NewServiceObservability(provider, "test-service")
		monitoringConfig := observability.MonitoringConfig{
			Enabled: true,
		}

		factory := observability.NewMonitoringDecoratorFactory(serviceObs, monitoringConfig)
		require.NotNil(t, factory)

		// Test creating generic decorator
		decorator := factory.CreateGenericDecorator("test-component", "service")
		assert.NotNil(t, decorator)
		assert.Equal(t, "service", decorator.GetComponentType())

		// Test updating config
		newConfig := observability.MonitoringConfig{
			Enabled:        false,
			MetricsEnabled: false,
		}
		factory.UpdateConfig(newConfig)

		// Create decorator with updated config
		disabledDecorator := factory.CreateGenericDecorator("test-component-2", "cache")
		assert.NotNil(t, disabledDecorator)

		// Test disabled decorator (should pass through without monitoring)
		ctx := context.Background()
		result, err := disabledDecorator.WrapMethod(ctx, "passthrough", func() (interface{}, error) {
			return "no monitoring", nil
		})

		assert.NoError(t, err)
		assert.Equal(t, "no monitoring", result)
	})

	t.Run("Metrics Collection", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:    "metrics-test",
			ServiceVersion: "1.0.0",
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		serviceObs := observability.NewServiceObservability(provider, "metrics-test")
		monitoringConfig := observability.MonitoringConfig{
			Enabled:             true,
			MetricsEnabled:      true,
			MaxOperationHistory: 100,
		}

		decorator := observability.NewGenericMonitoringDecorator(
			"test-component",
			serviceObs,
			"test",
			monitoringConfig,
		)

		ctx := context.Background()

		// Execute multiple operations to generate metrics
		for i := 0; i < 5; i++ {
			_, err := decorator.WrapMethod(ctx, "operation", func() (interface{}, error) {
				time.Sleep(10 * time.Millisecond) // Simulate work
				return i, nil
			})
			assert.NoError(t, err)
		}

		// Generate some errors
		for i := 0; i < 2; i++ {
			_, err := decorator.WrapMethod(ctx, "errorOperation", func() (interface{}, error) {
				return nil, fmt.Errorf("error %d", i)
			})
			assert.Error(t, err)
		}

		metrics := decorator.GetMetrics()

		// Check operation counts
		assert.Equal(t, int64(5), metrics.GetOperationCount("test.operation"))
		assert.Equal(t, int64(2), metrics.GetOperationCount("test.errorOperation"))
		assert.Equal(t, int64(2), metrics.GetErrorCount("test.errorOperation"))

		// Check average duration is reasonable
		avgDuration := metrics.GetAverageDuration("test.operation")
		assert.Greater(t, avgDuration, 5*time.Millisecond)
		assert.Less(t, avgDuration, 50*time.Millisecond)

		// Check metrics by labels
		metricsMap := metrics.GetMetricsByLabels(map[string]string{"service": "test"})
		assert.NotEmpty(t, metricsMap)
		assert.Contains(t, metricsMap, "test.operation.count")
		assert.Contains(t, metricsMap, "test.errorOperation.errors")
	})
}
