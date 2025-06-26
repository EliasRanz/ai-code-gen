package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Prometheus metrics
var (
	// HTTP request duration histogram
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTP request counter
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	// HTTP request size histogram
	httpRequestSizeBytes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "Size of HTTP requests in bytes",
			Buckets: prometheus.ExponentialBuckets(1, 2, 20),
		},
		[]string{"method", "path"},
	)

	// HTTP response size histogram
	httpResponseSizeBytes = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Size of HTTP responses in bytes",
			Buckets: prometheus.ExponentialBuckets(1, 2, 20),
		},
		[]string{"method", "path", "status_code"},
	)

	// Currently active HTTP requests
	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)
)

// RequestLogger creates a structured logging middleware
func RequestLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom log format using zerolog
		log.Info().
			Str("method", param.Method).
			Str("path", param.Path).
			Int("status", param.StatusCode).
			Dur("latency", param.Latency).
			Str("client_ip", param.ClientIP).
			Str("user_agent", param.Request.UserAgent()).
			Msg("HTTP Request")
		return ""
	})
}

// TracingMiddleware adds distributed tracing to requests
func TracingMiddleware(serviceName string) gin.HandlerFunc {
	tracer := otel.Tracer(serviceName)

	return func(c *gin.Context) {
		// Start a new span
		ctx, span := tracer.Start(c.Request.Context(), c.Request.Method+" "+c.Request.URL.Path)
		defer span.End()

		// Set span attributes
		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.route", c.FullPath()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
			attribute.String("http.client_ip", c.ClientIP()),
		)

		// Update context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		// Set response attributes
		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
			attribute.Int64("http.response_size", int64(c.Writer.Size())),
			attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1e6),
		)

		// Set span status based on HTTP status code
		if c.Writer.Status() >= 400 {
			span.SetStatus(codes.Error, "HTTP Error")
		}
	}
}

// MetricsMiddleware adds Prometheus metrics collection
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Increment in-flight requests counter
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()

		// Record request size if available
		if c.Request.ContentLength > 0 {
			httpRequestSizeBytes.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(float64(c.Request.ContentLength))
		}

		// Process request
		c.Next()

		// Calculate metrics
		duration := time.Since(start)
		statusCode := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()
		
		// If no route matched, use the raw path
		if path == "" {
			path = c.Request.URL.Path
		}

		// Record HTTP request duration
		httpRequestDuration.WithLabelValues(
			method,
			path,
			statusCode,
		).Observe(duration.Seconds())

		// Increment HTTP request counter
		httpRequestsTotal.WithLabelValues(
			method,
			path,
			statusCode,
		).Inc()

		// Record response size
		responseSize := c.Writer.Size()
		if responseSize > 0 {
			httpResponseSizeBytes.WithLabelValues(
				method,
				path,
				statusCode,
			).Observe(float64(responseSize))
		}

		log.Debug().
			Str("method", method).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Int("response_size", responseSize).
			Msg("Request metrics recorded")
	}
}

// ErrorHandler middleware for consistent error handling
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			log.Error().
				Err(err.Err).
				Str("path", c.Request.URL.Path).
				Str("method", c.Request.Method).
				Msg("Request error")

			// Return appropriate error response
			switch err.Type {
			case gin.ErrorTypeBind:
				c.JSON(400, gin.H{"error": "Invalid request data"})
			case gin.ErrorTypePublic:
				c.JSON(500, gin.H{"error": err.Error()})
			default:
				c.JSON(500, gin.H{"error": "Internal server error"})
			}
		}
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate a simple request ID (in production, use UUID)
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		// Add to logger context
		logger := log.With().Str("request_id", requestID).Logger()
		c.Set("logger", &logger)

		c.Next()
	}
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + "abc123" // Simplified for now
}
