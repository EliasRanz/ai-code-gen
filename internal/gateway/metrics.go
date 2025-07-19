package gateway

import (
	"sync"
	"time"

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

// GetRequestCount returns current request count (for testing)
func (m *MetricsCollector) GetRequestCount() int {
	// This is a simplified implementation for testing
	// In practice, you'd need to extract the actual counter value
	return 0
}
