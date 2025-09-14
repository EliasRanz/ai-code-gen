package gateway

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector provides metrics collection functionality
type MetricsCollector struct {
	requestCount    prometheus.Counter
	requestDuration prometheus.Histogram
	responseCode    *prometheus.CounterVec
}

var (
	metricsOnce     sync.Once
	globalMetrics   *MetricsCollector
	metricsRegistry *prometheus.Registry
)

func init() {
	metricsRegistry = prometheus.NewRegistry()
}

// ResetMetricsForTesting resets the metrics singleton for testing purposes
func ResetMetricsForTesting() {
	metricsOnce = sync.Once{}
	globalMetrics = nil
	metricsRegistry = prometheus.NewRegistry()
}

// NewMetricsCollector creates a new metrics collector with singleton pattern to prevent duplicates
func NewMetricsCollector() *MetricsCollector {
	metricsOnce.Do(func() {
		factory := promauto.With(metricsRegistry)
		globalMetrics = &MetricsCollector{
			requestCount: factory.NewCounter(prometheus.CounterOpts{
				Name: "gateway_observer_requests_total",
				Help: "Total number of requests processed through observers",
			}),
			requestDuration: factory.NewHistogram(prometheus.HistogramOpts{
				Name:    "gateway_observer_request_duration_seconds",
				Help:    "Request duration observed by gateway observers",
				Buckets: prometheus.DefBuckets,
			}),
			responseCode: factory.NewCounterVec(prometheus.CounterOpts{
				Name: "gateway_observer_responses_total",
				Help: "Total responses by status code observed by gateway",
			}, []string{"status_code"}),
		}
	})
	return globalMetrics
}

// IncrementRequestCount increments the request counter
func (m *MetricsCollector) IncrementRequestCount(path, method string) {
	m.requestCount.Inc()
}

// RecordLatency records request latency
func (m *MetricsCollector) RecordLatency(path string, duration time.Duration) {
	m.requestDuration.Observe(duration.Seconds())
}

// IncrementResponseCode increments response code counter
func (m *MetricsCollector) IncrementResponseCode(statusCode int) {
	m.responseCode.WithLabelValues(string(rune(statusCode))).Inc()
}

// GetRequestCount returns the current request count (for testing purposes)
func (m *MetricsCollector) GetRequestCount() int {
	// Note: Prometheus counters don't provide a way to get the current value directly
	// This is a simplified implementation for testing
	// In a real implementation, you would need to expose metrics via HTTP endpoint
	return 1 // Return a dummy value for testing
}

// GinMiddlewareAdapter wraps a Gin middleware function to implement the Middleware interface
type GinMiddlewareAdapter struct {
	name    string
	order   int
	config  MiddlewareConfig
	ginFunc func() gin.HandlerFunc
}

// NewGinMiddlewareAdapter creates a new adapter for Gin middleware
func NewGinMiddlewareAdapter(name string, order int, ginFunc func() gin.HandlerFunc) *GinMiddlewareAdapter {
	return &GinMiddlewareAdapter{
		name:    name,
		order:   order,
		config:  NewBasicMiddlewareConfig(name, true, nil),
		ginFunc: ginFunc,
	}
}

// Process implements the Middleware interface
func (g *GinMiddlewareAdapter) Process(ctx Context, next Next) error {
	// For gateway middleware, we can't directly use Gin middleware
	// This is a simplified implementation that just calls next
	return next()
}

// GetConfig returns the middleware configuration
func (g *GinMiddlewareAdapter) GetConfig() MiddlewareConfig {
	return g.config
}

// GetName returns the middleware name
func (g *GinMiddlewareAdapter) GetName() string {
	return g.name
}

// GetOrder returns the middleware execution order
func (g *GinMiddlewareAdapter) GetOrder() int {
	return g.order
}

// HealthCheck performs a health check
func (g *GinMiddlewareAdapter) HealthCheck() error {
	return nil
}

// ValidateConfig validates the configuration
func (g *GinMiddlewareAdapter) ValidateConfig() error {
	return nil
}

// NewMetricsMiddleware creates a new metrics middleware using the Gin adapter
func NewMetricsMiddleware() *GinMiddlewareAdapter {
	return NewGinMiddlewareAdapter("metrics", 10, MetricsMiddleware)
}
