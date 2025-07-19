package gateway

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// Prometheus metrics for logging middleware
var (
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status_code"},
	)

	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "gateway_http_requests_in_flight",
			Help: "Number of HTTP requests currently being processed",
		},
	)
)

// LoggingMiddlewareImpl provides structured logging and metrics collection
type LoggingMiddlewareImpl struct {
	name        string
	order       int
	serviceName string
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware() *LoggingMiddlewareImpl {
	return &LoggingMiddlewareImpl{
		name:        "logging",
		order:       10, // Run early in chain
		serviceName: "api-gateway",
	}
}

// Process implements middleware processing with logging and metrics
func (l *LoggingMiddlewareImpl) Process(ctx Context, next Next) error {
	start := time.Now()

	// Increment in-flight requests
	httpRequestsInFlight.Inc()
	defer httpRequestsInFlight.Dec()

	// Add request ID if not present
	requestID := ctx.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = l.generateRequestID()
	}

	// Always set the request ID in response header and context
	if ginCtx, ok := ctx.(*GinContextWrapper); ok {
		ginCtx.Header("X-Request-ID", requestID)
	}
	ctx.Set("request_id", requestID)

	// Start tracing span
	tracer := otel.Tracer(l.serviceName)
	_, span := tracer.Start(ctx.Request().Context(), ctx.Request().Method+" "+ctx.Request().URL.Path)
	defer span.End()

	// Set span attributes
	l.setSpanAttributes(span, ctx, requestID)

	// Log request start
	l.logRequestStart(ctx, requestID)

	// Process request through chain
	err := next()

	duration := time.Since(start)

	// Get response information from context
	statusCode := 200
	responseSize := 0
	if ginCtx, ok := ctx.(*GinContextWrapper); ok {
		statusCode = ginCtx.Status()
		responseSize = ginCtx.Size()
	}

	// Record metrics
	l.recordMetrics(ctx, statusCode, duration)

	// Update span with response data
	l.updateSpanWithResponse(span, statusCode, responseSize, duration, err)

	// Log request completion
	l.logRequestComplete(ctx, requestID, statusCode, duration, responseSize, err)

	return err
}

// LogRequest logs the incoming request
func (l *LoggingMiddlewareImpl) LogRequest(ctx Context) error {
	log.Info().
		Str("method", ctx.Request().Method).
		Str("path", ctx.Request().URL.Path).
		Str("client_ip", ctx.ClientIP()).
		Str("user_agent", ctx.Request().UserAgent()).
		Msg("HTTP Request received")
	return nil
}

// LogResponse logs the outgoing response
func (l *LoggingMiddlewareImpl) LogResponse(ctx Context) error {
	statusCode := 200
	if ginCtx, ok := ctx.(*GinContextWrapper); ok {
		statusCode = ginCtx.Status()
	}

	log.Info().
		Str("path", ctx.Request().URL.Path).
		Int("status", statusCode).
		Msg("HTTP Response sent")
	return nil
}

// GetConfig returns middleware configuration
func (l *LoggingMiddlewareImpl) GetConfig() MiddlewareConfig {
	return NewBasicMiddlewareConfig(l.name, true, map[string]interface{}{
		"service_name":    l.serviceName,
		"metrics_enabled": true,
		"tracing_enabled": true,
	})
}

// GetName returns middleware name
func (l *LoggingMiddlewareImpl) GetName() string {
	return l.name
}

// GetOrder returns middleware execution order
func (l *LoggingMiddlewareImpl) GetOrder() int {
	return l.order
}

// HealthCheck validates middleware health
func (l *LoggingMiddlewareImpl) HealthCheck() error {
	return nil // Logging middleware is always healthy
}

// ValidateConfig validates middleware configuration
func (l *LoggingMiddlewareImpl) ValidateConfig() error {
	return nil // No configuration to validate
}

// generateRequestID creates a simple request ID
func (l *LoggingMiddlewareImpl) generateRequestID() string {
	return time.Now().Format("20060102150405.000000")
}

// setSpanAttributes adds attributes to the tracing span
func (l *LoggingMiddlewareImpl) setSpanAttributes(span any, ctx Context, requestID string) {
	// This would work with proper OpenTelemetry span interface
	if s, ok := span.(interface {
		SetAttributes(...attribute.KeyValue)
	}); ok {
		s.SetAttributes(
			attribute.String("http.method", ctx.Request().Method),
			attribute.String("http.url", ctx.Request().URL.String()),
			attribute.String("http.route", ctx.Request().URL.Path),
			attribute.String("http.user_agent", ctx.Request().UserAgent()),
			attribute.String("http.client_ip", ctx.ClientIP()),
			attribute.String("request.id", requestID),
		)
	}
}

// updateSpanWithResponse updates span with response information
func (l *LoggingMiddlewareImpl) updateSpanWithResponse(span any, statusCode, responseSize int, duration time.Duration, err error) {
	if s, ok := span.(interface {
		SetAttributes(...attribute.KeyValue)
		SetStatus(codes.Code, string)
	}); ok {
		s.SetAttributes(
			attribute.Int("http.status_code", statusCode),
			attribute.Int("http.response_size", responseSize),
			attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1e6),
		)

		if statusCode >= 400 || err != nil {
			s.SetStatus(codes.Error, "HTTP Error")
		}
	}
}

// recordMetrics records Prometheus metrics
func (l *LoggingMiddlewareImpl) recordMetrics(ctx Context, statusCode int, duration time.Duration) {
	method := ctx.Request().Method
	path := ctx.Request().URL.Path
	statusCodeStr := strconv.Itoa(statusCode)

	// Record request duration
	httpRequestDuration.WithLabelValues(method, path, statusCodeStr).Observe(duration.Seconds())

	// Increment request counter
	httpRequestsTotal.WithLabelValues(method, path, statusCodeStr).Inc()
}

// logRequestStart logs the start of request processing
func (l *LoggingMiddlewareImpl) logRequestStart(ctx Context, requestID string) {
	log.Debug().
		Str("request_id", requestID).
		Str("method", ctx.Request().Method).
		Str("path", ctx.Request().URL.Path).
		Str("client_ip", ctx.ClientIP()).
		Msg("Gateway: Processing request")
}

// logRequestComplete logs the completion of request processing
func (l *LoggingMiddlewareImpl) logRequestComplete(ctx Context, requestID string, statusCode int, duration time.Duration, responseSize int, err error) {
	logger := log.Info().
		Str("request_id", requestID).
		Str("method", ctx.Request().Method).
		Str("path", ctx.Request().URL.Path).
		Int("status", statusCode).
		Dur("duration", duration).
		Int("response_size", responseSize)

	if err != nil {
		logger = logger.Err(err)
		logger.Msg("Gateway: Request completed with error")
	} else {
		logger.Msg("Gateway: Request completed successfully")
	}
}
