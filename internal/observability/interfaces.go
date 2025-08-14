package observability

import (
	"context"
	"time"
)

// ObservabilityProvider defines the core interface that all observability providers must implement
type ObservabilityProvider interface {
	// Metrics collection
	RecordMetric(name string, value float64, labels map[string]string) error
	IncrementCounter(name string, labels map[string]string) error
	RecordHistogram(name string, value float64, labels map[string]string) error
	RecordGauge(name string, value float64, labels map[string]string) error

	// Tracing
	StartSpan(ctx context.Context, name string) (context.Context, Span)
	CreateTracer(serviceName string) Tracer

	// Health checking
	RegisterHealthCheck(name string, check HealthCheckFunc) error
	GetHealthStatus() HealthStatus

	// Configuration and lifecycle
	Configure(config ObservabilityConfig) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// ObservabilityFactory creates observability providers
type ObservabilityFactory interface {
	CreateProvider(providerType string, config ObservabilityConfig) (ObservabilityProvider, error)
	CreateMultiProvider(providers []ObservabilityProvider) ObservabilityProvider
	ListAvailableProviders() []string
}

// ServiceObservability manages observability for a specific service
type ServiceObservability struct {
	provider    ObservabilityProvider
	serviceName string
	tracer      Tracer
	metrics     MetricRegistry
}

// MetricRegistry provides consistent metric collection interfaces
type MetricRegistry interface {
	Counter(name string) Counter
	Histogram(name string) Histogram
	Gauge(name string) Gauge
	Timer(name string) Timer
}

// Tracer interface for distributed tracing
type Tracer interface {
	StartSpan(ctx context.Context, operationName string) (context.Context, Span)
	InjectHeaders(span Span, headers map[string]string) error
	ExtractSpan(ctx context.Context, headers map[string]string) (context.Context, Span)
}

// Span represents a trace span
type Span interface {
	SetAttribute(key string, value interface{})
	SetStatus(code StatusCode, description string)
	RecordError(err error)
	End()
}

// Counter interface for counter metrics
type Counter interface {
	Inc()
	Add(delta float64)
	WithLabels(labels map[string]string) Counter
}

// Histogram interface for histogram metrics
type Histogram interface {
	Observe(value float64)
	WithLabels(labels map[string]string) Histogram
}

// Gauge interface for gauge metrics
type Gauge interface {
	Set(value float64)
	Inc()
	Dec()
	Add(delta float64)
	WithLabels(labels map[string]string) Gauge
}

// Timer interface for timing metrics
type Timer interface {
	Start() TimerContext
	ObserveDuration(start time.Time)
	WithLabels(labels map[string]string) Timer
}

// TimerContext represents an active timer
type TimerContext interface {
	Stop()
}

// HealthCheckFunc defines a health check function
type HealthCheckFunc func(ctx context.Context) error

// HealthStatus represents the overall health status
type HealthStatus struct {
	Status  string            `json:"status"`
	Checks  map[string]string `json:"checks"`
	Uptime  time.Duration     `json:"uptime"`
	Version string            `json:"version"`
}

// StatusCode represents span status codes
type StatusCode int

const (
	StatusCodeUnset StatusCode = iota
	StatusCodeOK
	StatusCodeError
)

// ObservabilityConfig holds configuration for observability providers
type ObservabilityConfig struct {
	ServiceName     string            `json:"service_name"`
	ServiceVersion  string            `json:"service_version"`
	Environment     string            `json:"environment"`
	MetricsEnabled  bool              `json:"metrics_enabled"`
	TracingEnabled  bool              `json:"tracing_enabled"`
	LoggingEnabled  bool              `json:"logging_enabled"`
	HealthEnabled   bool              `json:"health_enabled"`
	MetricsEndpoint string            `json:"metrics_endpoint"`
	TracingEndpoint string            `json:"tracing_endpoint"`
	SampleRate      float64           `json:"sample_rate"`
	Labels          map[string]string `json:"labels"`
}

// NewServiceObservability creates a new service observability manager
func NewServiceObservability(provider ObservabilityProvider, serviceName string) *ServiceObservability {
	return &ServiceObservability{
		provider:    provider,
		serviceName: serviceName,
		tracer:      provider.CreateTracer(serviceName),
		metrics:     &defaultMetricRegistry{provider: provider},
	}
}

// GetProvider returns the underlying observability provider
func (s *ServiceObservability) GetProvider() ObservabilityProvider {
	return s.provider
}

// GetTracer returns the service tracer
func (s *ServiceObservability) GetTracer() Tracer {
	return s.tracer
}

// GetMetrics returns the metric registry
func (s *ServiceObservability) GetMetrics() MetricRegistry {
	return s.metrics
}

// RecordRequest records a request metric
func (s *ServiceObservability) RecordRequest(method, endpoint string, duration time.Duration, statusCode int) error {
	labels := map[string]string{
		"service":     s.serviceName,
		"method":      method,
		"endpoint":    endpoint,
		"status_code": string(rune(statusCode)),
	}

	// Record multiple metrics
	if err := s.provider.IncrementCounter("http_requests_total", labels); err != nil {
		return err
	}

	return s.provider.RecordHistogram("http_request_duration_seconds", duration.Seconds(), labels)
}

// StartTrace starts a new trace span
func (s *ServiceObservability) StartTrace(ctx context.Context, operationName string) (context.Context, Span) {
	return s.tracer.StartSpan(ctx, operationName)
}

// defaultMetricRegistry provides a default implementation of MetricRegistry
type defaultMetricRegistry struct {
	provider ObservabilityProvider
}

func (r *defaultMetricRegistry) Counter(name string) Counter {
	return &defaultCounter{provider: r.provider, name: name}
}

func (r *defaultMetricRegistry) Histogram(name string) Histogram {
	return &defaultHistogram{provider: r.provider, name: name}
}

func (r *defaultMetricRegistry) Gauge(name string) Gauge {
	return &defaultGauge{provider: r.provider, name: name}
}

func (r *defaultMetricRegistry) Timer(name string) Timer {
	return &defaultTimer{provider: r.provider, name: name}
}

// Default metric implementations
type defaultCounter struct {
	provider ObservabilityProvider
	name     string
	labels   map[string]string
}

func (c *defaultCounter) Inc() {
	c.provider.IncrementCounter(c.name, c.labels)
}

func (c *defaultCounter) Add(delta float64) {
	c.provider.RecordMetric(c.name, delta, c.labels)
}

func (c *defaultCounter) WithLabels(labels map[string]string) Counter {
	newLabels := make(map[string]string)
	for k, v := range c.labels {
		newLabels[k] = v
	}
	for k, v := range labels {
		newLabels[k] = v
	}
	return &defaultCounter{provider: c.provider, name: c.name, labels: newLabels}
}

type defaultHistogram struct {
	provider ObservabilityProvider
	name     string
	labels   map[string]string
}

func (h *defaultHistogram) Observe(value float64) {
	h.provider.RecordHistogram(h.name, value, h.labels)
}

func (h *defaultHistogram) WithLabels(labels map[string]string) Histogram {
	newLabels := make(map[string]string)
	for k, v := range h.labels {
		newLabels[k] = v
	}
	for k, v := range labels {
		newLabels[k] = v
	}
	return &defaultHistogram{provider: h.provider, name: h.name, labels: newLabels}
}

type defaultGauge struct {
	provider ObservabilityProvider
	name     string
	labels   map[string]string
}

func (g *defaultGauge) Set(value float64) {
	g.provider.RecordGauge(g.name, value, g.labels)
}

func (g *defaultGauge) Inc() {
	g.provider.RecordGauge(g.name, 1, g.labels)
}

func (g *defaultGauge) Dec() {
	g.provider.RecordGauge(g.name, -1, g.labels)
}

func (g *defaultGauge) Add(delta float64) {
	g.provider.RecordGauge(g.name, delta, g.labels)
}

func (g *defaultGauge) WithLabels(labels map[string]string) Gauge {
	newLabels := make(map[string]string)
	for k, v := range g.labels {
		newLabels[k] = v
	}
	for k, v := range labels {
		newLabels[k] = v
	}
	return &defaultGauge{provider: g.provider, name: g.name, labels: newLabels}
}

type defaultTimer struct {
	provider ObservabilityProvider
	name     string
	labels   map[string]string
}

func (t *defaultTimer) Start() TimerContext {
	return &defaultTimerContext{
		timer: t,
		start: time.Now(),
	}
}

func (t *defaultTimer) ObserveDuration(start time.Time) {
	duration := time.Since(start)
	t.provider.RecordHistogram(t.name, duration.Seconds(), t.labels)
}

func (t *defaultTimer) WithLabels(labels map[string]string) Timer {
	newLabels := make(map[string]string)
	for k, v := range t.labels {
		newLabels[k] = v
	}
	for k, v := range labels {
		newLabels[k] = v
	}
	return &defaultTimer{provider: t.provider, name: t.name, labels: newLabels}
}

type defaultTimerContext struct {
	timer *defaultTimer
	start time.Time
}

func (tc *defaultTimerContext) Stop() {
	tc.timer.ObserveDuration(tc.start)
}
