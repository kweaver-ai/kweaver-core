package otel

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Providers 保存所有 provider，便于服务退出时统一关闭。
type Providers struct {
	TracerProvider *sdktrace.TracerProvider
	LoggerProvider *sdklog.LoggerProvider
}

// Shutdown 按顺序优雅关闭所有 provider。
func (p *Providers) Shutdown(ctx context.Context) {
	if p.TracerProvider != nil {
		if err := p.TracerProvider.Shutdown(ctx); err != nil {
			log.Printf("[OTel] Error shutting down tracer provider: %v", err)
		}
	}

	if p.LoggerProvider != nil {
		if err := p.LoggerProvider.Shutdown(ctx); err != nil {
			log.Printf("[OTel] Error shutting down logger provider: %v", err)
		}
	}
}

func newTracerProvider(ctx context.Context, endpoint string, samplingRate float64, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(samplingRate))

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	), nil
}

func newLoggerProvider(ctx context.Context, endpoint string, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	logExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}

	return sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
	), nil
}
