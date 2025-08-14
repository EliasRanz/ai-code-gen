package observability

import (
	"context"
	"fmt"
	"sync"
)

// observabilityFactory implements ObservabilityFactory
type observabilityFactory struct {
	providers map[string]func(ObservabilityConfig) (ObservabilityProvider, error)
	mu        sync.RWMutex
}

// NewObservabilityFactory creates a new observability factory
func NewObservabilityFactory() ObservabilityFactory {
	factory := &observabilityFactory{
		providers: make(map[string]func(ObservabilityConfig) (ObservabilityProvider, error)),
	}

	// Register default providers
	factory.RegisterProvider("prometheus", NewPrometheusProvider)
	factory.RegisterProvider("opentelemetry", NewOpenTelemetryProvider)
	factory.RegisterProvider("multi", func(config ObservabilityConfig) (ObservabilityProvider, error) {
		return NewMultiProvider([]ObservabilityProvider{}), nil
	})

	return factory
}

// RegisterProvider registers a new observability provider
func (f *observabilityFactory) RegisterProvider(name string, constructor func(ObservabilityConfig) (ObservabilityProvider, error)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers[name] = constructor
}

// CreateProvider creates an observability provider by type
func (f *observabilityFactory) CreateProvider(providerType string, config ObservabilityConfig) (ObservabilityProvider, error) {
	f.mu.RLock()
	constructor, exists := f.providers[providerType]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}

	return constructor(config)
}

// CreateMultiProvider creates a multi-provider that combines multiple providers
func (f *observabilityFactory) CreateMultiProvider(providers []ObservabilityProvider) ObservabilityProvider {
	return NewMultiProvider(providers)
}

// ListAvailableProviders returns a list of available provider types
func (f *observabilityFactory) ListAvailableProviders() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	providers := make([]string, 0, len(f.providers))
	for name := range f.providers {
		providers = append(providers, name)
	}
	return providers
}

// multiProvider implements ObservabilityProvider by delegating to multiple providers
type multiProvider struct {
	providers []ObservabilityProvider
	mu        sync.RWMutex
}

// NewMultiProvider creates a new multi-provider
func NewMultiProvider(providers []ObservabilityProvider) ObservabilityProvider {
	return &multiProvider{
		providers: providers,
	}
}

// AddProvider adds a provider to the multi-provider
func (m *multiProvider) AddProvider(provider ObservabilityProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers = append(m.providers, provider)
}

// RecordMetric records a metric to all providers
func (m *multiProvider) RecordMetric(name string, value float64, labels map[string]string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastError error
	for _, provider := range m.providers {
		if err := provider.RecordMetric(name, value, labels); err != nil {
			lastError = err
		}
	}
	return lastError
}

// IncrementCounter increments a counter in all providers
func (m *multiProvider) IncrementCounter(name string, labels map[string]string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastError error
	for _, provider := range m.providers {
		if err := provider.IncrementCounter(name, labels); err != nil {
			lastError = err
		}
	}
	return lastError
}

// RecordHistogram records a histogram value in all providers
func (m *multiProvider) RecordHistogram(name string, value float64, labels map[string]string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastError error
	for _, provider := range m.providers {
		if err := provider.RecordHistogram(name, value, labels); err != nil {
			lastError = err
		}
	}
	return lastError
}

// RecordGauge records a gauge value in all providers
func (m *multiProvider) RecordGauge(name string, value float64, labels map[string]string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastError error
	for _, provider := range m.providers {
		if err := provider.RecordGauge(name, value, labels); err != nil {
			lastError = err
		}
	}
	return lastError
}

// StartSpan starts a span using the first provider that supports tracing
func (m *multiProvider) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, provider := range m.providers {
		if ctx, span := provider.StartSpan(ctx, name); span != nil {
			return ctx, span
		}
	}
	// Return a no-op span if no provider supports tracing
	return ctx, &noOpSpan{}
}

// CreateTracer creates a tracer using the first provider that supports tracing
func (m *multiProvider) CreateTracer(serviceName string) Tracer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, provider := range m.providers {
		if tracer := provider.CreateTracer(serviceName); tracer != nil {
			return tracer
		}
	}
	// Return a no-op tracer if no provider supports tracing
	return &noOpTracer{}
}

// RegisterHealthCheck registers a health check in all providers
func (m *multiProvider) RegisterHealthCheck(name string, check HealthCheckFunc) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastError error
	for _, provider := range m.providers {
		if err := provider.RegisterHealthCheck(name, check); err != nil {
			lastError = err
		}
	}
	return lastError
}

// GetHealthStatus gets health status from the first provider
func (m *multiProvider) GetHealthStatus() HealthStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, provider := range m.providers {
		status := provider.GetHealthStatus()
		if status.Status != "" {
			return status
		}
	}
	return HealthStatus{Status: "unknown"}
}

// Configure configures all providers
func (m *multiProvider) Configure(config ObservabilityConfig) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastError error
	for _, provider := range m.providers {
		if err := provider.Configure(config); err != nil {
			lastError = err
		}
	}
	return lastError
}

// Start starts all providers
func (m *multiProvider) Start(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastError error
	for _, provider := range m.providers {
		if err := provider.Start(ctx); err != nil {
			lastError = err
		}
	}
	return lastError
}

// Stop stops all providers
func (m *multiProvider) Stop(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var lastError error
	for _, provider := range m.providers {
		if err := provider.Stop(ctx); err != nil {
			lastError = err
		}
	}
	return lastError
}

// No-op implementations for when no provider supports a feature

type noOpSpan struct{}

func (s *noOpSpan) SetAttribute(key string, value interface{})    {}
func (s *noOpSpan) SetStatus(code StatusCode, description string) {}
func (s *noOpSpan) RecordError(err error)                         {}
func (s *noOpSpan) End()                                          {}

type noOpTracer struct{}

func (t *noOpTracer) StartSpan(ctx context.Context, operationName string) (context.Context, Span) {
	return ctx, &noOpSpan{}
}

func (t *noOpTracer) InjectHeaders(span Span, headers map[string]string) error {
	return nil
}

func (t *noOpTracer) ExtractSpan(ctx context.Context, headers map[string]string) (context.Context, Span) {
	return ctx, &noOpSpan{}
}
