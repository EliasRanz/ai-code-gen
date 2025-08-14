package observability

import (
	"context"
	"fmt"
	"time"
)

// MonitoringDecorator defines the interface for adding monitoring to existing components
type MonitoringDecorator interface {
	WrapComponent(component interface{}) interface{}
	GetMetrics() MonitoringMetrics
	Configure(config MonitoringConfig) error
}

// MonitoringMetrics provides access to collected monitoring metrics
type MonitoringMetrics interface {
	GetOperationCount(operation string) int64
	GetErrorCount(operation string) int64
	GetAverageDuration(operation string) time.Duration
	GetMetricsByLabels(labels map[string]string) map[string]interface{}
}

// MonitoringConfig holds configuration for monitoring decorators
type MonitoringConfig struct {
	Enabled             bool              `json:"enabled"`
	MetricsEnabled      bool              `json:"metrics_enabled"`
	TracingEnabled      bool              `json:"tracing_enabled"`
	LoggingEnabled      bool              `json:"logging_enabled"`
	SampleRate          float64           `json:"sample_rate"`
	DetailLevel         string            `json:"detail_level"` // "basic", "detailed", "verbose"
	Labels              map[string]string `json:"labels"`
	IncludeStackTrace   bool              `json:"include_stack_trace"`
	MaxOperationHistory int               `json:"max_operation_history"`
}

// Generic component monitoring decorator that can wrap any interface
type GenericMonitoringDecorator struct {
	component     interface{}
	observability *ServiceObservability
	config        MonitoringConfig
	componentType string
	metrics       *genericMetrics
}

// genericMetrics tracks operation metrics for any component
type genericMetrics struct {
	operationCounts map[string]int64
	errorCounts     map[string]int64
	durations       map[string][]time.Duration
}

// NewGenericMonitoringDecorator creates a new generic monitoring decorator
func NewGenericMonitoringDecorator(component interface{}, observability *ServiceObservability, componentType string, config MonitoringConfig) *GenericMonitoringDecorator {
	return &GenericMonitoringDecorator{
		component:     component,
		observability: observability,
		componentType: componentType,
		config:        config,
		metrics: &genericMetrics{
			operationCounts: make(map[string]int64),
			errorCounts:     make(map[string]int64),
			durations:       make(map[string][]time.Duration),
		},
	}
}

// WrapMethod wraps a method call with monitoring
func (g *GenericMonitoringDecorator) WrapMethod(ctx context.Context, methodName string, fn func() (interface{}, error)) (interface{}, error) {
	if !g.config.Enabled {
		return fn()
	}

	operation := fmt.Sprintf("%s.%s", g.componentType, methodName)

	// Start span for tracing
	ctx, span := g.observability.StartTrace(ctx, operation)
	defer span.End()

	// Add span attributes
	span.SetAttribute("component_type", g.componentType)
	span.SetAttribute("method", methodName)

	// Record metrics
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		labels := map[string]string{
			"component": g.componentType,
			"method":    methodName,
		}
		g.observability.GetProvider().RecordHistogram("component_operation_duration", duration.Seconds(), labels)
		g.metrics.durations[operation] = append(g.metrics.durations[operation], duration)
	}()

	// Execute operation with error tracking
	g.metrics.operationCounts[operation]++
	result, err := fn()

	if err != nil {
		g.metrics.errorCounts[operation]++
		labels := map[string]string{
			"component":  g.componentType,
			"method":     methodName,
			"error_type": g.getErrorType(err),
		}
		g.observability.GetProvider().IncrementCounter("component_errors", labels)
		span.SetStatus(StatusCodeError, err.Error())
		span.RecordError(err)
	} else {
		labels := map[string]string{
			"component": g.componentType,
			"method":    methodName,
		}
		g.observability.GetProvider().IncrementCounter("component_operations", labels)
		span.SetStatus(StatusCodeOK, "success")
	}

	return result, err
}

// WrapVoidMethod wraps a method call that doesn't return a value with monitoring
func (g *GenericMonitoringDecorator) WrapVoidMethod(ctx context.Context, methodName string, fn func() error) error {
	if !g.config.Enabled {
		return fn()
	}

	operation := fmt.Sprintf("%s.%s", g.componentType, methodName)

	// Start span for tracing
	ctx, span := g.observability.StartTrace(ctx, operation)
	defer span.End()

	// Add span attributes
	span.SetAttribute("component_type", g.componentType)
	span.SetAttribute("method", methodName)

	// Record metrics
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		labels := map[string]string{
			"component": g.componentType,
			"method":    methodName,
		}
		g.observability.GetProvider().RecordHistogram("component_operation_duration", duration.Seconds(), labels)
		g.metrics.durations[operation] = append(g.metrics.durations[operation], duration)
	}()

	// Execute operation with error tracking
	g.metrics.operationCounts[operation]++
	err := fn()

	if err != nil {
		g.metrics.errorCounts[operation]++
		labels := map[string]string{
			"component":  g.componentType,
			"method":     methodName,
			"error_type": g.getErrorType(err),
		}
		g.observability.GetProvider().IncrementCounter("component_errors", labels)
		span.SetStatus(StatusCodeError, err.Error())
		span.RecordError(err)
	} else {
		labels := map[string]string{
			"component": g.componentType,
			"method":    methodName,
		}
		g.observability.GetProvider().IncrementCounter("component_operations", labels)
		span.SetStatus(StatusCodeOK, "success")
	}

	return err
}

// WrapComponent implements MonitoringDecorator interface
func (g *GenericMonitoringDecorator) WrapComponent(component interface{}) interface{} {
	g.component = component
	return g
}

// GetMetrics implements MonitoringDecorator interface
func (g *GenericMonitoringDecorator) GetMetrics() MonitoringMetrics {
	return g.metrics
}

// Configure implements MonitoringDecorator interface
func (g *GenericMonitoringDecorator) Configure(config MonitoringConfig) error {
	g.config = config
	return nil
}

// MonitoringDecoratorFactory creates monitoring decorators
type MonitoringDecoratorFactory struct {
	observability *ServiceObservability
	config        MonitoringConfig
}

// NewMonitoringDecoratorFactory creates a new monitoring decorator factory
func NewMonitoringDecoratorFactory(observability *ServiceObservability, config MonitoringConfig) *MonitoringDecoratorFactory {
	return &MonitoringDecoratorFactory{
		observability: observability,
		config:        config,
	}
}

// CreateGenericDecorator creates a generic monitoring decorator
func (f *MonitoringDecoratorFactory) CreateGenericDecorator(component interface{}, componentType string) *GenericMonitoringDecorator {
	return NewGenericMonitoringDecorator(component, f.observability, componentType, f.config)
}

// UpdateConfig updates the factory configuration
func (f *MonitoringDecoratorFactory) UpdateConfig(config MonitoringConfig) {
	f.config = config
}

// Implement MonitoringMetrics for genericMetrics
func (m *genericMetrics) GetOperationCount(operation string) int64 {
	return m.operationCounts[operation]
}

func (m *genericMetrics) GetErrorCount(operation string) int64 {
	return m.errorCounts[operation]
}

func (m *genericMetrics) GetAverageDuration(operation string) time.Duration {
	durations := m.durations[operation]
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

func (m *genericMetrics) GetMetricsByLabels(labels map[string]string) map[string]interface{} {
	metrics := make(map[string]interface{})

	// Filter metrics based on labels (simplified implementation)
	for operation, count := range m.operationCounts {
		metrics[fmt.Sprintf("%s.count", operation)] = count
	}

	for operation, count := range m.errorCounts {
		metrics[fmt.Sprintf("%s.errors", operation)] = count
	}

	for operation := range m.durations {
		metrics[fmt.Sprintf("%s.avg_duration", operation)] = m.GetAverageDuration(operation)
	}

	return metrics
}

// Helper methods
func (g *GenericMonitoringDecorator) getErrorType(err error) string {
	return fmt.Sprintf("%T", err)
}

// GetComponent returns the wrapped component
func (g *GenericMonitoringDecorator) GetComponent() interface{} {
	return g.component
}

// GetComponentType returns the component type name
func (g *GenericMonitoringDecorator) GetComponentType() string {
	return g.componentType
}

// Example usage pattern for specific decorators:
//
// // Repository decorator example
// type MonitoredRepository struct {
//     *GenericMonitoringDecorator
//     repo Repository
// }
//
// func (m *MonitoredRepository) Create(ctx context.Context, entity interface{}) error {
//     return m.WrapVoidMethod(ctx, "Create", func() error {
//         return m.repo.Create(ctx, entity)
//     })
// }
//
// // Cache decorator example
// type MonitoredCache struct {
//     *GenericMonitoringDecorator
//     cache CacheProvider
// }
//
// func (m *MonitoredCache) Get(ctx context.Context, key string) (string, error) {
//     result, err := m.WrapMethod(ctx, "Get", func() (interface{}, error) {
//         return m.cache.Get(ctx, key)
//     })
//     if err != nil {
//         return "", err
//     }
//     return result.(string), nil
// }
