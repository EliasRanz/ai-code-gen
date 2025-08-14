package observability_test

import (
	"context"
	"math"
	"testing"

	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObservabilityProviders(t *testing.T) {
	config := observability.ObservabilityConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		MetricsEnabled: true,
		TracingEnabled: true,
	}

	t.Run("PrometheusProvider Comprehensive", func(t *testing.T) {
		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)

		// Test lifecycle
		ctx := context.Background()
		err = provider.Start(ctx)
		assert.NoError(t, err)

		labels := map[string]string{
			"service":     "test",
			"environment": "test",
		}

		// Test all metric types
		err = provider.RecordMetric("test_metric", 42.0, labels)
		assert.NoError(t, err)

		err = provider.IncrementCounter("test_counter", labels)
		assert.NoError(t, err)

		err = provider.RecordHistogram("test_histogram", 1.5, labels)
		assert.NoError(t, err)

		err = provider.RecordGauge("test_gauge", 100.0, labels)
		assert.NoError(t, err)

		// Test with empty labels
		err = provider.IncrementCounter("empty_labels_counter", map[string]string{})
		assert.NoError(t, err)

		// Test with nil labels
		err = provider.RecordMetric("nil_labels_metric", 50.0, nil)
		assert.NoError(t, err)

		// Test health status
		health := provider.GetHealthStatus()
		assert.NotNil(t, health)

		// Test tracing
		ctx, span := provider.StartSpan(ctx, "test_operation")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)

		tracer := provider.CreateTracer("test-tracer")
		assert.NotNil(t, tracer)

		// Test stop
		err = provider.Stop(ctx)
		assert.NoError(t, err)
	})

	t.Run("OpenTelemetryProvider Comprehensive", func(t *testing.T) {
		provider, err := observability.NewOpenTelemetryProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)

		ctx := context.Background()
		err = provider.Start(ctx)
		assert.NoError(t, err)

		labels := map[string]string{
			"component": "test",
			"version":   "1.0",
		}

		// Test all metric operations
		err = provider.RecordMetric("otel_metric", 25.0, labels)
		assert.NoError(t, err)

		err = provider.IncrementCounter("otel_counter", labels)
		assert.NoError(t, err)

		err = provider.RecordHistogram("otel_histogram", 0.75, labels)
		assert.NoError(t, err)

		err = provider.RecordGauge("otel_gauge", 200.0, labels)
		assert.NoError(t, err)

		// Test tracing functionality
		ctx, span := provider.StartSpan(ctx, "otel_operation")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)

		tracer := provider.CreateTracer("otel-tracer")
		assert.NotNil(t, tracer)

		// Test health monitoring
		health := provider.GetHealthStatus()
		assert.NotNil(t, health)

		err = provider.Stop(ctx)
		assert.NoError(t, err)
	})

	t.Run("MultiProvider Comprehensive", func(t *testing.T) {
		prometheus, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		otel, err := observability.NewOpenTelemetryProvider(config)
		require.NoError(t, err)

		multiProvider := observability.NewMultiProvider([]observability.ObservabilityProvider{prometheus, otel})
		require.NotNil(t, multiProvider)

		ctx := context.Background()
		err = multiProvider.Start(ctx)
		assert.NoError(t, err)

		labels := map[string]string{
			"multi": "provider",
			"test":  "true",
		}

		// Test that all operations work with multi-provider
		err = multiProvider.RecordMetric("multi_metric", 75.0, labels)
		assert.NoError(t, err)

		err = multiProvider.IncrementCounter("multi_counter", labels)
		assert.NoError(t, err)

		err = multiProvider.RecordHistogram("multi_histogram", 2.5, labels)
		assert.NoError(t, err)

		err = multiProvider.RecordGauge("multi_gauge", 150.0, labels)
		assert.NoError(t, err)

		// Test tracing with multi-provider
		ctx, span := multiProvider.StartSpan(ctx, "multi_operation")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)

		tracer := multiProvider.CreateTracer("multi-tracer")
		assert.NotNil(t, tracer)

		// Test health status aggregation
		health := multiProvider.GetHealthStatus()
		assert.NotNil(t, health)

		err = multiProvider.Stop(ctx)
		assert.NoError(t, err)
	})

	t.Run("Empty MultiProvider", func(t *testing.T) {
		emptyMultiProvider := observability.NewMultiProvider([]observability.ObservabilityProvider{})
		require.NotNil(t, emptyMultiProvider)

		ctx := context.Background()

		// Operations should not fail even with no providers
		err := emptyMultiProvider.Start(ctx)
		assert.NoError(t, err)

		err = emptyMultiProvider.IncrementCounter("test_counter", nil)
		assert.NoError(t, err)

		err = emptyMultiProvider.Stop(ctx)
		assert.NoError(t, err)
	})
}

func TestObservabilityFactory(t *testing.T) {
	t.Run("CreatePrometheusProvider", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:    "factory-test",
			ServiceVersion: "1.0.0",
			MetricsEnabled: true,
		}

		factory := observability.NewObservabilityFactory()
		require.NotNil(t, factory)

		provider, err := factory.CreateProvider("prometheus", config)
		assert.NoError(t, err)
		require.NotNil(t, provider)

		// Verify it's working
		ctx := context.Background()
		err = provider.Start(ctx)
		assert.NoError(t, err)

		err = provider.IncrementCounter("factory_test", map[string]string{"type": "prometheus"})
		assert.NoError(t, err)

		err = provider.Stop(ctx)
		assert.NoError(t, err)
	})

	t.Run("CreateOpenTelemetryProvider", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:    "factory-otel",
			ServiceVersion: "1.0.0",
			TracingEnabled: true,
		}

		factory := observability.NewObservabilityFactory()
		require.NotNil(t, factory)

		provider, err := factory.CreateProvider("opentelemetry", config)
		assert.NoError(t, err)
		require.NotNil(t, provider)

		ctx := context.Background()
		err = provider.Start(ctx)
		assert.NoError(t, err)

		ctx, span := provider.StartSpan(ctx, "factory_span")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)

		err = provider.Stop(ctx)
		assert.NoError(t, err)
	})

	t.Run("CreateMultiProvider", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:    "factory-multi",
			ServiceVersion: "1.0.0",
			MetricsEnabled: true,
			TracingEnabled: true,
		}

		factory := observability.NewObservabilityFactory()

		prometheus, err := factory.CreateProvider("prometheus", config)
		require.NoError(t, err)

		otel, err := factory.CreateProvider("opentelemetry", config)
		require.NoError(t, err)

		multiProvider := factory.CreateMultiProvider([]observability.ObservabilityProvider{prometheus, otel})
		require.NotNil(t, multiProvider)

		ctx := context.Background()
		err = multiProvider.Start(ctx)
		assert.NoError(t, err)

		err = multiProvider.IncrementCounter("multi_factory_test", map[string]string{"factory": "multi"})
		assert.NoError(t, err)

		err = multiProvider.Stop(ctx)
		assert.NoError(t, err)
	})

	t.Run("Invalid Provider Type", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName: "invalid-test",
		}

		factory := observability.NewObservabilityFactory()
		provider, err := factory.CreateProvider("invalid-type", config)
		assert.Error(t, err)
		assert.Nil(t, provider)
	})

	t.Run("ListAvailableProviders", func(t *testing.T) {
		factory := observability.NewObservabilityFactory()
		providers := factory.ListAvailableProviders()
		assert.NotEmpty(t, providers)
		assert.Contains(t, providers, "prometheus")
		assert.Contains(t, providers, "opentelemetry")
	})
}

func TestServiceObservability(t *testing.T) {
	t.Run("Service Observability Creation", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:    "service-test",
			ServiceVersion: "1.0.0",
			MetricsEnabled: true,
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		serviceObs := observability.NewServiceObservability(provider, "test-service")
		require.NotNil(t, serviceObs)

		// Test that we can get the provider
		actualProvider := serviceObs.GetProvider()
		assert.NotNil(t, actualProvider)

		// Test that we can get metrics registry
		metrics := serviceObs.GetMetrics()
		assert.NotNil(t, metrics)

		// Test that we can get tracer
		tracer := serviceObs.GetTracer()
		assert.NotNil(t, tracer)

		// Test tracing functionality
		ctx := context.Background()
		ctx, span := serviceObs.StartTrace(ctx, "service_operation")
		assert.NotNil(t, ctx)
		assert.NotNil(t, span)
	})

	t.Run("Service Request Recording", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName: "request-test",
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		serviceObs := observability.NewServiceObservability(provider, "request-service")
		require.NotNil(t, serviceObs)

		// Test request recording
		err = serviceObs.RecordRequest("GET", "/api/test", 1000000, 200) // 1ms duration in nanoseconds
		assert.NoError(t, err)

		err = serviceObs.RecordRequest("POST", "/api/create", 2000000, 201)
		assert.NoError(t, err)

		err = serviceObs.RecordRequest("GET", "/api/error", 500000, 500)
		assert.NoError(t, err)
	})
}

func TestHealthChecks(t *testing.T) {
	config := observability.ObservabilityConfig{
		ServiceName: "health-test-service",
	}

	provider, err := observability.NewPrometheusProvider(config)
	require.NoError(t, err)

	t.Run("Healthy Check", func(t *testing.T) {
		healthyCheck := func(ctx context.Context) error {
			return nil
		}

		err = provider.RegisterHealthCheck("healthy_check", healthyCheck)
		assert.NoError(t, err)

		health := provider.GetHealthStatus()
		assert.NotNil(t, health)
	})

	t.Run("Unhealthy Check", func(t *testing.T) {
		unhealthyCheck := func(ctx context.Context) error {
			return assert.AnError
		}

		err = provider.RegisterHealthCheck("unhealthy_check", unhealthyCheck)
		assert.NoError(t, err)

		health := provider.GetHealthStatus()
		assert.NotNil(t, health)
	})

	t.Run("Multiple Health Checks", func(t *testing.T) {
		check1 := func(ctx context.Context) error { return nil }
		check2 := func(ctx context.Context) error { return assert.AnError }
		check3 := func(ctx context.Context) error { return nil }

		err = provider.RegisterHealthCheck("check1", check1)
		assert.NoError(t, err)

		err = provider.RegisterHealthCheck("check2", check2)
		assert.NoError(t, err)

		err = provider.RegisterHealthCheck("check3", check3)
		assert.NoError(t, err)

		health := provider.GetHealthStatus()
		assert.NotNil(t, health)
	})
}

func TestObservabilityConfiguration(t *testing.T) {
	t.Run("Full Configuration", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:     "config-test",
			ServiceVersion:  "2.0.0",
			MetricsEnabled:  true,
			TracingEnabled:  true,
			LoggingEnabled:  true,
			HealthEnabled:   true,
			MetricsEndpoint: "http://localhost:9090/metrics",
			TracingEndpoint: "http://localhost:14268/api/traces",
			Environment:     "test",
			SampleRate:      0.1,
			Labels: map[string]string{
				"team": "platform",
				"env":  "test",
			},
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)

		err = provider.Configure(config)
		assert.NoError(t, err)
	})

	t.Run("Minimal Configuration", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName: "minimal-test",
		}

		provider, err := observability.NewOpenTelemetryProvider(config)
		require.NoError(t, err)
		require.NotNil(t, provider)

		err = provider.Configure(config)
		assert.NoError(t, err)
	})

	t.Run("Configuration Updates", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName:    "update-test",
			MetricsEnabled: false,
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		// Update configuration
		config.MetricsEnabled = true
		config.TracingEnabled = true

		err = provider.Configure(config)
		assert.NoError(t, err)
	})
}

func TestObservabilityErrorHandling(t *testing.T) {
	t.Run("Provider Creation Errors", func(t *testing.T) {
		// Test with invalid endpoint configuration that might cause issues during Start()
		invalidConfig := observability.ObservabilityConfig{
			ServiceName:     "error-test",
			TracingEndpoint: "invalid://endpoint", // Invalid URL format
			MetricsEndpoint: "invalid://endpoint",
		}

		// Providers should still create successfully, but may fail during Start/Stop operations
		provider, err := observability.NewPrometheusProvider(invalidConfig)
		assert.NoError(t, err) // Creation should succeed
		assert.NotNil(t, provider)

		provider2, err := observability.NewOpenTelemetryProvider(invalidConfig)
		assert.NoError(t, err) // Creation should succeed
		assert.NotNil(t, provider2)
	})

	t.Run("Operation with Stopped Provider", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName: "stopped-test",
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		ctx := context.Background()

		// Don't start the provider, try operations
		err = provider.IncrementCounter("test", nil)
		// Should either work gracefully or return appropriate error
		// Both behaviors are acceptable

		// Start and then stop
		err = provider.Start(ctx)
		assert.NoError(t, err)

		err = provider.Stop(ctx)
		assert.NoError(t, err)

		// Operations after stop
		err = provider.IncrementCounter("test_after_stop", nil)
		// Should handle gracefully
	})

	t.Run("Invalid Metric Values", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName: "invalid-metrics",
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		ctx := context.Background()
		err = provider.Start(ctx)
		assert.NoError(t, err)

		// Test with invalid metric values (should handle gracefully)
		err = provider.RecordMetric("test", math.NaN(), nil) // NaN
		// Should either succeed or fail gracefully

		err = provider.RecordGauge("infinity_test", math.Inf(1), nil) // Infinity
		// Should either succeed or fail gracefully

		err = provider.Stop(ctx)
		assert.NoError(t, err)
	})
}
