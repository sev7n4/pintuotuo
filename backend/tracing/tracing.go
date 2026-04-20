// Package tracing configures OpenTelemetry for distributed tracing (Tempo / OTLP).
package tracing

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var enabled bool

// Enabled reports whether OTLP export is configured (Tempo reachable from backend).
func Enabled() bool {
	return enabled
}

// Init configures global propagators and, when OTEL_EXPORTER_OTLP_ENDPOINT is set, an OTLP gRPC exporter to Tempo.
func Init(ctx context.Context) (shutdown func(context.Context) error, err error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if isDisabled() {
		otel.SetTracerProvider(nooptrace.NewTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	endpoint := normalizeEndpoint(strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")))
	if endpoint == "" {
		otel.SetTracerProvider(nooptrace.NewTracerProvider())
		return func(context.Context) error { return nil }, nil
	}

	serviceName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	if serviceName == "" {
		serviceName = "pintuotuo-backend"
	}

	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(semconv.ServiceName(serviceName)),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	}
	exp, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("otlp trace exporter: %w", err)
	}

	sampler := buildSampler()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)
	otel.SetTracerProvider(tp)
	enabled = true

	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			return fmt.Errorf("tracer shutdown: %w", err)
		}
		return nil
	}, nil
}

func isDisabled() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_SDK_DISABLED")))
	return v == "1" || v == "true" || v == "yes"
}

func normalizeEndpoint(endpoint string) string {
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")
	return strings.TrimSpace(endpoint)
}

func buildSampler() sdktrace.Sampler {
	name := strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_TRACES_SAMPLER")))
	arg := 1.0
	if s := strings.TrimSpace(os.Getenv("OTEL_TRACES_SAMPLER_ARG")); s != "" {
		_, _ = fmt.Sscanf(s, "%f", &arg)
	}
	switch name {
	case "always_on":
		return sdktrace.AlwaysSample()
	case "always_off":
		return sdktrace.NeverSample()
	case "traceidratio":
		return sdktrace.TraceIDRatioBased(arg)
	case "parentbased_traceidratio":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(arg))
	default:
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(1.0))
	}
}

// SetEndUserID sets semantic end-user id on the current span when present.
func SetEndUserID(ctx context.Context, userID int) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() || userID <= 0 {
		return
	}
	span.SetAttributes(attribute.String("enduser.id", fmt.Sprintf("%d", userID)))
}
