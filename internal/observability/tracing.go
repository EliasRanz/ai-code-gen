package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

// InitTracing initializes OpenTelemetry tracing
func InitTracing(serviceName, jaegerEndpoint string) error {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		return fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)

	// Get tracer
	tracer = otel.Tracer(serviceName)

	return nil
}

// StartSpan starts a new tracing span
func StartSpan(ctx context.Context, operationName string) (context.Context, trace.Span) {
	return tracer.Start(ctx, operationName)
}

// StartSpanWithOptions starts a new tracing span with options
func StartSpanWithOptions(ctx context.Context, operationName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tracer.Start(ctx, operationName, opts...)
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(ctx context.Context, name string, attributes map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		// Convert attributes to the correct type
		var attrs []attribute.KeyValue
		for key, value := range attributes {
			// This is a simplified conversion - in practice you'd want proper type handling
			switch v := value.(type) {
			case string:
				attrs = append(attrs, attribute.String(key, v))
			case int:
				attrs = append(attrs, attribute.Int(key, v))
			case bool:
				attrs = append(attrs, attribute.Bool(key, v))
			}
		}
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// SetSpanError sets an error on the current span
func SetSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span != nil && err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// SetSpanAttributes sets attributes on the current span
func SetSpanAttributes(ctx context.Context, attributes map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		// Convert and set attributes
		var attrs []attribute.KeyValue
		for key, value := range attributes {
			// This is a simplified conversion - in practice you'd want proper type handling
			switch v := value.(type) {
			case string:
				attrs = append(attrs, attribute.String(key, v))
			case int:
				attrs = append(attrs, attribute.Int(key, v))
			case bool:
				attrs = append(attrs, attribute.Bool(key, v))
			}
		}
		span.SetAttributes(attrs...)
	}
}

// FinishSpan finishes the current span
func FinishSpan(span trace.Span) {
	if span != nil {
		span.End()
	}
}

// GetTracer returns the global tracer
func GetTracer() trace.Tracer {
	return tracer
}
