package observability

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

// prometheusProvider implements ObservabilityProvider using Prometheus for metrics
type prometheusProvider struct {
	config       ObservabilityConfig
	registry     *prometheus.Registry
	healthChecks map[string]HealthCheckFunc
	startTime    time.Time
	mu           sync.RWMutex

	// Prometheus metrics
	counters   map[string]*prometheus.CounterVec
	histograms map[string]*prometheus.HistogramVec
	gauges     map[string]*prometheus.GaugeVec
}

// NewPrometheusProvider creates a new Prometheus observability provider
func NewPrometheusProvider(config ObservabilityConfig) (ObservabilityProvider, error) {
	return &prometheusProvider{
		config:       config,
		registry:     prometheus.NewRegistry(),
		healthChecks: make(map[string]HealthCheckFunc),
		counters:     make(map[string]*prometheus.CounterVec),
		histograms:   make(map[string]*prometheus.HistogramVec),
		gauges:       make(map[string]*prometheus.GaugeVec),
		startTime:    time.Now(),
	}, nil
}

// RecordMetric records a generic metric (delegates to appropriate type)
func (p *prometheusProvider) RecordMetric(name string, value float64, labels map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Try to find existing gauge
	if gauge, exists := p.gauges[name]; exists {
		labelValues := p.extractLabelValues(labels)
		gauge.WithLabelValues(labelValues...).Set(value)
		return nil
	}

	// Create new gauge if not found
	labelNames := p.extractLabelNames(labels)
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: fmt.Sprintf("Generic metric: %s", name),
	}, labelNames)

	if err := p.registry.Register(gauge); err != nil {
		return fmt.Errorf("failed to register gauge %s: %w", name, err)
	}

	p.gauges[name] = gauge
	labelValues := p.extractLabelValues(labels)
	gauge.WithLabelValues(labelValues...).Set(value)
	return nil
}

// IncrementCounter increments a counter metric
func (p *prometheusProvider) IncrementCounter(name string, labels map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	counter, exists := p.counters[name]
	if !exists {
		labelNames := p.extractLabelNames(labels)
		counter = prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: fmt.Sprintf("Counter metric: %s", name),
		}, labelNames)

		if err := p.registry.Register(counter); err != nil {
			return fmt.Errorf("failed to register counter %s: %w", name, err)
		}
		p.counters[name] = counter
	}

	labelValues := p.extractLabelValues(labels)
	counter.WithLabelValues(labelValues...).Inc()
	return nil
}

// RecordHistogram records a histogram value
func (p *prometheusProvider) RecordHistogram(name string, value float64, labels map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	histogram, exists := p.histograms[name]
	if !exists {
		labelNames := p.extractLabelNames(labels)
		histogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    name,
			Help:    fmt.Sprintf("Histogram metric: %s", name),
			Buckets: prometheus.DefBuckets,
		}, labelNames)

		if err := p.registry.Register(histogram); err != nil {
			return fmt.Errorf("failed to register histogram %s: %w", name, err)
		}
		p.histograms[name] = histogram
	}

	labelValues := p.extractLabelValues(labels)
	histogram.WithLabelValues(labelValues...).Observe(value)
	return nil
}

// RecordGauge records a gauge value
func (p *prometheusProvider) RecordGauge(name string, value float64, labels map[string]string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	gauge, exists := p.gauges[name]
	if !exists {
		labelNames := p.extractLabelNames(labels)
		gauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: fmt.Sprintf("Gauge metric: %s", name),
		}, labelNames)

		if err := p.registry.Register(gauge); err != nil {
			return fmt.Errorf("failed to register gauge %s: %w", name, err)
		}
		p.gauges[name] = gauge
	}

	labelValues := p.extractLabelValues(labels)
	gauge.WithLabelValues(labelValues...).Set(value)
	return nil
}

// StartSpan starts a tracing span (Prometheus doesn't support tracing)
func (p *prometheusProvider) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	return ctx, &noOpSpan{}
}

// CreateTracer creates a tracer (Prometheus doesn't support tracing)
func (p *prometheusProvider) CreateTracer(serviceName string) Tracer {
	return &noOpTracer{}
}

// RegisterHealthCheck registers a health check
func (p *prometheusProvider) RegisterHealthCheck(name string, check HealthCheckFunc) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.healthChecks[name] = check
	return nil
}

// GetHealthStatus returns the current health status
func (p *prometheusProvider) GetHealthStatus() HealthStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := HealthStatus{
		Status:  "healthy",
		Checks:  make(map[string]string),
		Uptime:  time.Since(p.startTime),
		Version: p.config.ServiceVersion,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for name, check := range p.healthChecks {
		if err := check(ctx); err != nil {
			status.Status = "unhealthy"
			status.Checks[name] = err.Error()
		} else {
			status.Checks[name] = "healthy"
		}
	}

	return status
}

// Configure configures the provider
func (p *prometheusProvider) Configure(config ObservabilityConfig) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = config
	return nil
}

// Start starts the provider
func (p *prometheusProvider) Start(ctx context.Context) error {
	return nil
}

// Stop stops the provider
func (p *prometheusProvider) Stop(ctx context.Context) error {
	return nil
}

// GetMetricsHandler returns the Prometheus metrics HTTP handler
func (p *prometheusProvider) GetMetricsHandler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}

// Helper methods
func (p *prometheusProvider) extractLabelNames(labels map[string]string) []string {
	names := make([]string, 0, len(labels))
	for name := range labels {
		names = append(names, name)
	}
	return names
}

func (p *prometheusProvider) extractLabelValues(labels map[string]string) []string {
	// We need to maintain consistent ordering based on label names
	labelNames := p.extractLabelNames(labels)
	values := make([]string, len(labelNames))
	for i, name := range labelNames {
		if value, exists := labels[name]; exists {
			values[i] = value
		} else {
			values[i] = ""
		}
	}
	return values
}

// openTelemetryProvider implements ObservabilityProvider using OpenTelemetry
type openTelemetryProvider struct {
	config       ObservabilityConfig
	tracer       trace.Tracer
	healthChecks map[string]HealthCheckFunc
	startTime    time.Time
	mu           sync.RWMutex
}

// NewOpenTelemetryProvider creates a new OpenTelemetry observability provider
func NewOpenTelemetryProvider(config ObservabilityConfig) (ObservabilityProvider, error) {
	provider := &openTelemetryProvider{
		config:       config,
		healthChecks: make(map[string]HealthCheckFunc),
		startTime:    time.Now(),
	}

	if config.TracingEnabled && config.TracingEndpoint != "" {
		if err := provider.initializeTracing(); err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
	}

	return provider, nil
}

func (o *openTelemetryProvider) initializeTracing() error {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(o.config.TracingEndpoint)))
	if err != nil {
		return fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(o.config.ServiceName),
			semconv.ServiceVersionKey.String(o.config.ServiceVersion),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(o.config.SampleRate)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Get tracer
	o.tracer = otel.Tracer(o.config.ServiceName)

	return nil
}

// RecordMetric records a generic metric (OpenTelemetry metrics not implemented in this example)
func (o *openTelemetryProvider) RecordMetric(name string, value float64, labels map[string]string) error {
	// OpenTelemetry metrics would be implemented here
	return nil
}

// IncrementCounter increments a counter (OpenTelemetry metrics not implemented in this example)
func (o *openTelemetryProvider) IncrementCounter(name string, labels map[string]string) error {
	// OpenTelemetry metrics would be implemented here
	return nil
}

// RecordHistogram records a histogram (OpenTelemetry metrics not implemented in this example)
func (o *openTelemetryProvider) RecordHistogram(name string, value float64, labels map[string]string) error {
	// OpenTelemetry metrics would be implemented here
	return nil
}

// RecordGauge records a gauge (OpenTelemetry metrics not implemented in this example)
func (o *openTelemetryProvider) RecordGauge(name string, value float64, labels map[string]string) error {
	// OpenTelemetry metrics would be implemented here
	return nil
}

// StartSpan starts a tracing span
func (o *openTelemetryProvider) StartSpan(ctx context.Context, name string) (context.Context, Span) {
	if o.tracer == nil {
		return ctx, &noOpSpan{}
	}

	ctx, span := o.tracer.Start(ctx, name)
	return ctx, &otelSpan{span: span}
}

// CreateTracer creates a tracer
func (o *openTelemetryProvider) CreateTracer(serviceName string) Tracer {
	if o.tracer == nil {
		return &noOpTracer{}
	}
	return &otelTracer{tracer: otel.Tracer(serviceName)}
}

// RegisterHealthCheck registers a health check
func (o *openTelemetryProvider) RegisterHealthCheck(name string, check HealthCheckFunc) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.healthChecks[name] = check
	return nil
}

// GetHealthStatus returns the current health status
func (o *openTelemetryProvider) GetHealthStatus() HealthStatus {
	o.mu.RLock()
	defer o.mu.RUnlock()

	status := HealthStatus{
		Status:  "healthy",
		Checks:  make(map[string]string),
		Uptime:  time.Since(o.startTime),
		Version: o.config.ServiceVersion,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for name, check := range o.healthChecks {
		if err := check(ctx); err != nil {
			status.Status = "unhealthy"
			status.Checks[name] = err.Error()
		} else {
			status.Checks[name] = "healthy"
		}
	}

	return status
}

// Configure configures the provider
func (o *openTelemetryProvider) Configure(config ObservabilityConfig) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.config = config
	return nil
}

// Start starts the provider
func (o *openTelemetryProvider) Start(ctx context.Context) error {
	return nil
}

// Stop stops the provider
func (o *openTelemetryProvider) Stop(ctx context.Context) error {
	return nil
}

// OpenTelemetry span wrapper
type otelSpan struct {
	span trace.Span
}

func (s *otelSpan) SetAttribute(key string, value interface{}) {
	s.span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", value)))
}

func (s *otelSpan) SetStatus(code StatusCode, description string) {
	// Convert custom status code to OpenTelemetry codes
	var otelCode codes.Code
	switch code {
	case StatusCodeOK:
		otelCode = codes.Ok
	case StatusCodeError:
		otelCode = codes.Error
	default:
		otelCode = codes.Unset
	}
	s.span.SetStatus(otelCode, description)
}

func (s *otelSpan) RecordError(err error) {
	s.span.RecordError(err)
}

func (s *otelSpan) End() {
	s.span.End()
}

// OpenTelemetry tracer wrapper
type otelTracer struct {
	tracer trace.Tracer
}

func (t *otelTracer) StartSpan(ctx context.Context, operationName string) (context.Context, Span) {
	ctx, span := t.tracer.Start(ctx, operationName)
	return ctx, &otelSpan{span: span}
}

func (t *otelTracer) InjectHeaders(span Span, headers map[string]string) error {
	// Header injection would be implemented here
	return nil
}

func (t *otelTracer) ExtractSpan(ctx context.Context, headers map[string]string) (context.Context, Span) {
	// Header extraction would be implemented here
	return ctx, &noOpSpan{}
}
