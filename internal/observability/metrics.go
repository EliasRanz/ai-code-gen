package observability

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code", "service"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "service"},
	)

	// Service metrics
	serviceUptime = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_uptime_seconds_total",
			Help: "Total uptime of the service in seconds",
		},
		[]string{"service", "version"},
	)

	activeConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
		[]string{"service"},
	)
)

// InitMetrics initializes the metrics system
func InitMetrics(serviceName string) error {
	// Register metrics
	prometheus.MustRegister(
		httpRequestsTotal,
		httpRequestDuration,
		serviceUptime,
		activeConnections,
	)

	return nil
}

// GetMetricsHandler returns the Prometheus metrics handler
func GetMetricsHandler() http.Handler {
	return promhttp.Handler()
}

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, endpoint, statusCode, service string, duration float64) {
	httpRequestsTotal.WithLabelValues(method, endpoint, statusCode, service).Inc()
	httpRequestDuration.WithLabelValues(method, endpoint, service).Observe(duration)
}

// IncrementServiceUptime increments service uptime
func IncrementServiceUptime(service, version string) {
	serviceUptime.WithLabelValues(service, version).Inc()
}

// SetActiveConnections sets the number of active connections
func SetActiveConnections(service string, count float64) {
	activeConnections.WithLabelValues(service).Set(count)
}
