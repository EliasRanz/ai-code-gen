package observability_test

import (
	"context"
	"testing"
	"time"

	"github.com/EliasRanz/ai-code-gen/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenericMonitoringDecorator(t *testing.T) {
	config := observability.ObservabilityConfig{
		ServiceName:    "decorator-test",
		ServiceVersion: "1.0.0",
		MetricsEnabled: true,
		TracingEnabled: true,
	}

	provider, err := observability.NewPrometheusProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.Start(ctx)
	require.NoError(t, err)

	serviceObs := observability.NewServiceObservability(provider, "decorator-test")

	t.Run("Basic Decorator Creation", func(t *testing.T) {
		component := "test-component"

		// Test configuration
		monitoringConfig := observability.MonitoringConfig{
			Enabled:        true,
			MetricsEnabled: true,
			TracingEnabled: true,
		}

		decorator := observability.NewGenericMonitoringDecorator(
			component,
			serviceObs,
			"test-component-type",
			monitoringConfig,
		)
		require.NotNil(t, decorator)

		// Test wrapping component
		wrappedComponent := decorator.WrapComponent(component)
		assert.NotNil(t, wrappedComponent)

		// Test metrics collection
		metrics := decorator.GetMetrics()
		assert.NotNil(t, metrics)

		// Test configuration
		newConfig := observability.MonitoringConfig{
			Enabled:        true,
			MetricsEnabled: false,
		}
		err := decorator.Configure(newConfig)
		assert.NoError(t, err)
	})

	t.Run("Decorator Method Wrapping", func(t *testing.T) {
		component := "test-component"

		decorator := observability.NewGenericMonitoringDecorator(
			component,
			serviceObs,
			"repository",
			observability.MonitoringConfig{Enabled: true},
		)

		// Test method wrapping with proper signature
		testMethod := func() (interface{}, error) {
			return "test-result", nil
		}

		result, err := decorator.WrapMethod(ctx, "TestMethod", testMethod)
		assert.NoError(t, err)
		assert.Equal(t, "test-result", result)
	})

	t.Run("Decorator Error Handling", func(t *testing.T) {
		component := "error-test-component"

		decorator := observability.NewGenericMonitoringDecorator(
			component,
			serviceObs,
			"error-test",
			observability.MonitoringConfig{Enabled: true},
		)

		// Test method that returns error
		errorMethod := func() (interface{}, error) {
			return nil, assert.AnError
		}

		result, err := decorator.WrapMethod(ctx, "ErrorMethod", errorMethod)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("Decorator Performance Tracking", func(t *testing.T) {
		component := "perf-component"

		decorator := observability.NewGenericMonitoringDecorator(
			component,
			serviceObs,
			"performance",
			observability.MonitoringConfig{
				Enabled:        true,
				MetricsEnabled: true,
			},
		)

		// Test method with some work
		performanceMethod := func() (interface{}, error) {
			time.Sleep(1 * time.Millisecond)
			return "performance-result", nil
		}

		result, err := decorator.WrapMethod(ctx, "PerformanceMethod", performanceMethod)
		assert.NoError(t, err)
		assert.Equal(t, "performance-result", result)

		// Test void method wrapping
		voidMethod := func() error {
			time.Sleep(1 * time.Millisecond)
			return nil
		}

		err = decorator.WrapVoidMethod(ctx, "VoidMethod", voidMethod)
		assert.NoError(t, err)
	})

	err = provider.Stop(ctx)
	assert.NoError(t, err)
}

func TestMonitoringDecoratorFactory(t *testing.T) {
	config := observability.ObservabilityConfig{
		ServiceName:    "factory-test",
		ServiceVersion: "1.0.0",
		MetricsEnabled: true,
	}

	provider, err := observability.NewPrometheusProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.Start(ctx)
	require.NoError(t, err)

	serviceObs := observability.NewServiceObservability(provider, "factory-test")

	t.Run("Factory Creation", func(t *testing.T) {
		monitoringConfig := observability.MonitoringConfig{
			Enabled:        true,
			MetricsEnabled: true,
			TracingEnabled: true,
		}

		factory := observability.NewMonitoringDecoratorFactory(serviceObs, monitoringConfig)
		require.NotNil(t, factory)
	})

	t.Run("Create Generic Decorator", func(t *testing.T) {
		factory := observability.NewMonitoringDecoratorFactory(serviceObs, observability.MonitoringConfig{
			Enabled:        true,
			MetricsEnabled: true,
		})

		component := "test-component"

		decorator := factory.CreateGenericDecorator(component, "test-component-type")
		require.NotNil(t, decorator)

		// Test that the decorator works
		testMethod := func() (interface{}, error) {
			return "factory-result", nil
		}

		result, err := decorator.WrapMethod(ctx, "TestMethod", testMethod)
		assert.NoError(t, err)
		assert.Equal(t, "factory-result", result)
	})

	t.Run("Factory Configuration Update", func(t *testing.T) {
		factory := observability.NewMonitoringDecoratorFactory(serviceObs, observability.MonitoringConfig{
			Enabled: false,
		})

		// Test configuration update
		newConfig := observability.MonitoringConfig{
			Enabled:        true,
			MetricsEnabled: true,
			TracingEnabled: true,
		}

		factory.UpdateConfig(newConfig)

		// Verify configuration was updated by creating a new decorator
		decorator := factory.CreateGenericDecorator("updated-component", "updated-type")
		assert.NotNil(t, decorator)
	})

	err = provider.Stop(ctx)
	assert.NoError(t, err)
}

func TestDecoratorConfiguration(t *testing.T) {
	t.Run("Monitoring Config Creation", func(t *testing.T) {
		config := observability.MonitoringConfig{
			Enabled:        true,
			MetricsEnabled: true,
			TracingEnabled: true,
			LoggingEnabled: true,
			SampleRate:     0.1,
			DetailLevel:    "detailed",
			Labels: map[string]string{
				"team":        "platform",
				"environment": "test",
			},
			IncludeStackTrace:   true,
			MaxOperationHistory: 1000,
		}

		assert.True(t, config.Enabled)
		assert.True(t, config.MetricsEnabled)
		assert.True(t, config.TracingEnabled)
		assert.True(t, config.LoggingEnabled)
		assert.Equal(t, 0.1, config.SampleRate)
		assert.Equal(t, "detailed", config.DetailLevel)
		assert.NotNil(t, config.Labels)
		assert.True(t, config.IncludeStackTrace)
		assert.Equal(t, 1000, config.MaxOperationHistory)
	})

	t.Run("Decorator Configuration Update", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName: "config-update-test",
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		serviceObs := observability.NewServiceObservability(provider, "config-update-test")

		decorator := observability.NewGenericMonitoringDecorator(
			"test-component",
			serviceObs,
			"test-type",
			observability.MonitoringConfig{Enabled: false},
		)

		// Test initial configuration
		monConfig1 := observability.MonitoringConfig{
			Enabled:        true,
			MetricsEnabled: false,
		}

		err = decorator.Configure(monConfig1)
		assert.NoError(t, err)

		// Test configuration update
		monConfig2 := observability.MonitoringConfig{
			Enabled:        true,
			MetricsEnabled: true,
			TracingEnabled: true,
		}

		err = decorator.Configure(monConfig2)
		assert.NoError(t, err)
	})
}

func TestDecoratorMetrics(t *testing.T) {
	config := observability.ObservabilityConfig{
		ServiceName:    "metrics-test",
		MetricsEnabled: true,
	}

	provider, err := observability.NewPrometheusProvider(config)
	require.NoError(t, err)

	ctx := context.Background()
	err = provider.Start(ctx)
	require.NoError(t, err)

	serviceObs := observability.NewServiceObservability(provider, "metrics-test")

	t.Run("Decorator Metrics Collection", func(t *testing.T) {
		decorator := observability.NewGenericMonitoringDecorator(
			"metrics-collector",
			serviceObs,
			"metrics-component",
			observability.MonitoringConfig{
				Enabled:        true,
				MetricsEnabled: true,
			},
		)

		// Test metrics collection
		metrics := decorator.GetMetrics()
		require.NotNil(t, metrics)

		// Execute some operations to generate metrics
		testMethod := func() (interface{}, error) {
			return "metrics-test", nil
		}

		_, err := decorator.WrapMethod(ctx, "MetricsTestMethod", testMethod)
		assert.NoError(t, err)

		// Check that metrics are recorded
		operationCount := metrics.GetOperationCount("metrics-component.MetricsTestMethod")
		assert.Greater(t, operationCount, int64(0))

		errorCount := metrics.GetErrorCount("metrics-component.MetricsTestMethod")
		assert.Equal(t, int64(0), errorCount)
	})

	t.Run("Decorator Metrics Labels", func(t *testing.T) {
		decorator := observability.NewGenericMonitoringDecorator(
			"labels-test",
			serviceObs,
			"labels-component",
			observability.MonitoringConfig{
				Enabled:        true,
				MetricsEnabled: true,
				Labels: map[string]string{
					"version":     "1.0.0",
					"environment": "test",
					"team":        "platform",
				},
			},
		)

		metrics := decorator.GetMetrics()
		assert.NotNil(t, metrics)

		// Test metrics with labels
		labelsFilter := map[string]string{
			"component": "labels-component",
		}
		metricsByLabels := metrics.GetMetricsByLabels(labelsFilter)
		assert.NotNil(t, metricsByLabels)
	})

	err = provider.Stop(ctx)
	assert.NoError(t, err)
}

func TestDecoratorIntegration(t *testing.T) {
	t.Run("End-to-End Decorator Flow", func(t *testing.T) {
		// Setup observability
		config := observability.ObservabilityConfig{
			ServiceName:    "integration-test",
			ServiceVersion: "1.0.0",
			MetricsEnabled: true,
			TracingEnabled: true,
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		ctx := context.Background()
		err = provider.Start(ctx)
		require.NoError(t, err)

		serviceObs := observability.NewServiceObservability(provider, "integration-test")

		// Create factory
		monConfig := observability.MonitoringConfig{
			Enabled:        true,
			MetricsEnabled: true,
			TracingEnabled: true,
		}

		factory := observability.NewMonitoringDecoratorFactory(serviceObs, monConfig)
		require.NotNil(t, factory)

		// Create and wrap component
		component := "integration-component"
		decorator := factory.CreateGenericDecorator(component, "integration-type")
		require.NotNil(t, decorator)

		// Test full operation cycle
		createMethod := func() (interface{}, error) {
			return "created", nil
		}

		result, err := decorator.WrapMethod(ctx, "Create", createMethod)
		assert.NoError(t, err)
		assert.Equal(t, "created", result)

		getMethod := func() (interface{}, error) {
			return "retrieved", nil
		}

		result, err = decorator.WrapMethod(ctx, "Get", getMethod)
		assert.NoError(t, err)
		assert.Equal(t, "retrieved", result)

		updateVoidMethod := func() error {
			return nil
		}

		err = decorator.WrapVoidMethod(ctx, "Update", updateVoidMethod)
		assert.NoError(t, err)

		deleteVoidMethod := func() error {
			return nil
		}

		err = decorator.WrapVoidMethod(ctx, "Delete", deleteVoidMethod)
		assert.NoError(t, err)

		// Verify metrics were collected
		metrics := decorator.GetMetrics()
		assert.NotNil(t, metrics)

		// Check operation counts
		createCount := metrics.GetOperationCount("integration-type.Create")
		assert.Equal(t, int64(1), createCount)

		getCount := metrics.GetOperationCount("integration-type.Get")
		assert.Equal(t, int64(1), getCount)

		err = provider.Stop(ctx)
		assert.NoError(t, err)
	})

	t.Run("Error Tracking Integration", func(t *testing.T) {
		config := observability.ObservabilityConfig{
			ServiceName: "error-tracking-test",
		}

		provider, err := observability.NewPrometheusProvider(config)
		require.NoError(t, err)

		ctx := context.Background()
		err = provider.Start(ctx)
		require.NoError(t, err)

		serviceObs := observability.NewServiceObservability(provider, "error-tracking-test")

		factory := observability.NewMonitoringDecoratorFactory(serviceObs, observability.MonitoringConfig{
			Enabled: true,
		})

		// Test with error-generating methods
		decorator := factory.CreateGenericDecorator("error-component", "error-type")

		// Test error tracking
		errorMethod := func() (interface{}, error) {
			return nil, assert.AnError
		}

		result, err := decorator.WrapMethod(ctx, "ErrorMethod", errorMethod)
		assert.Error(t, err)
		assert.Nil(t, result)

		voidErrorMethod := func() error {
			return assert.AnError
		}

		err = decorator.WrapVoidMethod(ctx, "VoidErrorMethod", voidErrorMethod)
		assert.Error(t, err)

		// Verify error metrics
		metrics := decorator.GetMetrics()
		errorCount := metrics.GetErrorCount("error-type.ErrorMethod")
		assert.Equal(t, int64(1), errorCount)

		voidErrorCount := metrics.GetErrorCount("error-type.VoidErrorMethod")
		assert.Equal(t, int64(1), voidErrorCount)

		err = provider.Stop(ctx)
		assert.NoError(t, err)
	})
}
