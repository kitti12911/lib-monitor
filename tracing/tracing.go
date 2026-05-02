package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

type Config struct {
	Enabled     bool    `mapstructure:"enabled"      env:"TRACING_ENABLED"`
	Endpoint    string  `mapstructure:"endpoint"     env:"TRACING_ENDPOINT"     validate:"required_if=Enabled true,omitempty,hostname_port"`
	Protocol    string  `mapstructure:"protocol"     env:"TRACING_PROTOCOL"     validate:"omitempty,oneof=grpc http"`
	Insecure    bool    `mapstructure:"insecure"     env:"TRACING_INSECURE"`
	SampleRatio float64 `mapstructure:"sample_ratio" env:"TRACING_SAMPLE_RATIO" validate:"omitempty,gte=0,lte=1"`
}

func New(ctx context.Context, serviceName, collectorEndpoint string) (*sdktrace.TracerProvider, error) {
	return NewFromConfig(ctx, serviceName, Config{
		Enabled:  true,
		Endpoint: collectorEndpoint,
		Insecure: true,
	})
}

func NewFromConfig(ctx context.Context, serviceName string, cfg Config) (*sdktrace.TracerProvider, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	exporterOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.Endpoint),
	}

	if cfg.Insecure {
		exporterOpts = append(exporterOpts, otlptracegrpc.WithInsecure())
	}

	exporter, err := otlptracegrpc.New(ctx,
		exporterOpts...,
	)

	if err != nil {
		return nil, fmt.Errorf("tracing: create exporter: %w", err)
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		),
	)

	if err != nil {
		return nil, fmt.Errorf("tracing: create resource: %w", err)
	}

	tracerProviderOpts := []sdktrace.TracerProviderOption{
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))),
	}

	tp := sdktrace.NewTracerProvider(tracerProviderOpts...)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp, nil
}

func Shutdown(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp == nil {
		return nil
	}

	return tp.Shutdown(ctx)
}
