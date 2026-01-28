package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds tracing configuration.
type Config struct {
	Enabled     bool
	ServiceName string
	Endpoint    string
	SampleRate  float64
}

// Tracer wraps OpenTelemetry tracer.
type Tracer struct {
	tracer   trace.Tracer
	provider *sdktrace.TracerProvider
}

// New creates a new tracer with OTLP HTTP exporter.
func New(ctx context.Context, cfg Config) (*Tracer, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	// Create OTLP HTTP exporter (compatible with Jaeger, Zipkin, etc.)
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.SampleRate)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Tracer{
		tracer:   tp.Tracer(cfg.ServiceName),
		provider: tp,
	}, nil
}

// Shutdown shuts down the tracer.
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t == nil || t.provider == nil {
		return nil
	}
	return t.provider.Shutdown(ctx)
}

// StartSpan starts a new span.
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if t == nil || t.tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return t.tracer.Start(ctx, name, opts...)
}

// AddEvent adds an event to the current span.
func AddEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// SetAttributes sets attributes on the current span.
func SetAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.SetAttributes(attrs...)
	}
}

// RecordError records an error on the current span.
func RecordError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.RecordError(err)
	}
}
