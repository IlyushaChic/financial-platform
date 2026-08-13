package tracer

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Config содержит настройки для инициализации трейсера.
type Config struct {
	ServiceName  string        `env:"OTEL_SERVICE_NAME" default:"unknown-service"`
	CollectorURL string        `env:"OTEL_COLLECTOR_URL" default:"localhost:4317"`
	SamplingRate float64       `env:"OTEL_SAMPLING_RATE" default:"1.0"`
	Timeout      time.Duration `env:"OTEL_TIMEOUT" default:"5s"`
	Insecure     bool          `env:"OTEL_INSECURE" default:"true"`
}

// Init инициализирует глобальный TracerProvider и возвращает его.
func Init(ctx context.Context, cfg Config) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.CollectorURL),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(cfg.Timeout),
	)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, err
	}

	sampler := sdktrace.AlwaysSample()
	if cfg.SamplingRate < 1.0 {
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplingRate)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}
