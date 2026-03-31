package opentelemetry

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	otelsdklog "go.opentelemetry.io/otel/sdk/log"
	otelsdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Provider 管理 OpenTelemetry 的所有提供者
type Provider struct {
	config         *conf.OtelConfig
	tracer         trace.Tracer
	tracerProvider *sdktrace.TracerProvider
	loggerProvider *otelsdklog.LoggerProvider
	meterProvider  *otelsdkmetric.MeterProvider
	shutdownFuncs  []func(context.Context) error
}

// NewProvider 创建新的 OpenTelemetry 提供者
func NewProvider(cfg *conf.OtelConfig) (*Provider, error) {
	// 创建资源
	res, err := createResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	provider := &Provider{
		config: cfg,
	}

	// 初始化 Trace 提供者
	if cfg.Trace.Enabled {
		if err := provider.initTracer(res); err != nil {
			return nil, fmt.Errorf("failed to init tracer: %w", err)
		}
	}

	// 初始化 Log 提供者
	if cfg.Log.Enabled {
		if err := provider.initLogger(res); err != nil {
			return nil, fmt.Errorf("failed to init logger: %w", err)
		}
	}

	// 初始化 Metric 提供者
	if cfg.Metric.Enabled {
		if err := provider.initMeter(res); err != nil {
			return nil, fmt.Errorf("failed to init meter: %w", err)
		}
	}

	// 设置全局传播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return provider, nil
}

// createResource 创建 OpenTelemetry 资源
func createResource(cfg *conf.OtelConfig) (*resource.Resource, error) {
	// 基础属性
	attrs := []attribute.KeyValue{
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("service.version", cfg.ServiceVersion),
	}

	// 创建资源
	res := resource.NewWithAttributes(
		"https://opentelemetry.io/schemas/1.39.0",
		attrs...,
	)

	return res, nil
}

// initTracer 初始化 Trace 提供者
func (p *Provider) initTracer(res *resource.Resource) error {
	var exporter sdktrace.SpanExporter

	var err error

	// 根据配置选择 exporter
	switch p.config.Trace.Exporter {
	case "console":
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	case "http":
		exporter, err = otlptracehttp.New(context.Background(),
			otlptracehttp.WithEndpointURL(p.config.Trace.HTTPEndpoint),
			otlptracehttp.WithInsecure(),
		)
	default:
		return fmt.Errorf("unsupported trace exporter: %s", p.config.Trace.Exporter)
	}

	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// 创建采样器
	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(p.config.Trace.SamplingRate))

	// 创建 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// 设置全局 TracerProvider
	otel.SetTracerProvider(tp)

	// 创建 tracer
	p.tracer = tp.Tracer(
		p.config.ServiceName,
		trace.WithInstrumentationVersion(p.config.ServiceVersion),
		trace.WithSchemaURL("https://opentelemetry.io/schemas/1.39.0"),
	)

	p.tracerProvider = tp

	// 添加关闭函数
	p.shutdownFuncs = append(p.shutdownFuncs, func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	})

	log.Printf("Trace provider initialized with %s exporter", p.config.Trace.Exporter)

	return nil
}

// initLogger 初始化 Log 提供者
func (p *Provider) initLogger(res *resource.Resource) error {
	var exporter otelsdklog.Exporter

	var err error

	// 根据配置选择 exporter
	switch p.config.Log.Exporter {
	case "console":
		exporter, err = stdoutlog.New(
			stdoutlog.WithPrettyPrint(),
		)
	case "http":
		exporter, err = otlploghttp.New(context.Background(),
			otlploghttp.WithEndpointURL(p.config.Log.HTTPEndpoint),
			otlploghttp.WithInsecure(),
		)
	default:
		return fmt.Errorf("unsupported log exporter: %s", p.config.Log.Exporter)
	}

	if err != nil {
		return fmt.Errorf("failed to create log exporter: %w", err)
	}

	// 创建 LoggerProvider
	lp := otelsdklog.NewLoggerProvider(
		otelsdklog.WithProcessor(otelsdklog.NewBatchProcessor(exporter)),
		otelsdklog.WithResource(res),
	)

	// 设置全局 LoggerProvider
	p.loggerProvider = lp

	// 添加关闭函数
	p.shutdownFuncs = append(p.shutdownFuncs, func(ctx context.Context) error {
		return lp.Shutdown(ctx)
	})

	return nil
}

// initMeter 初始化 Metric 提供者
func (p *Provider) initMeter(res *resource.Resource) error {
	var exporter otelsdkmetric.Exporter

	var err error

	// 根据配置选择 exporter
	switch p.config.Metric.Exporter {
	case "console":
		exporter, err = stdoutmetric.New(
			stdoutmetric.WithPrettyPrint(),
		)
	case "http":
		exporter, err = otlpmetrichttp.New(context.Background(),
			otlpmetrichttp.WithEndpointURL(p.config.Metric.HTTPEndpoint),
			otlpmetrichttp.WithInsecure(),
		)
	default:
		return fmt.Errorf("unsupported metric exporter: %s", p.config.Metric.Exporter)
	}

	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// 创建 MeterProvider
	mp := otelsdkmetric.NewMeterProvider(
		otelsdkmetric.WithReader(otelsdkmetric.NewPeriodicReader(
			exporter,
			otelsdkmetric.WithInterval(time.Duration(p.config.Metric.ExportInterval)*time.Second),
		)),
		otelsdkmetric.WithResource(res),
	)

	// 设置全局 MeterProvider
	p.meterProvider = mp

	// 添加关闭函数
	p.shutdownFuncs = append(p.shutdownFuncs, func(ctx context.Context) error {
		return mp.Shutdown(ctx)
	})

	log.Printf("Metric provider initialized with %s exporter", p.config.Metric.Exporter)

	return nil
}

// Tracer 获取 tracer
func (p *Provider) Tracer() trace.Tracer {
	return p.tracer
}

// LoggerProvider 获取 logger provider
func (p *Provider) LoggerProvider() *otelsdklog.LoggerProvider {
	return p.loggerProvider
}

// MeterProvider 获取 meter provider
func (p *Provider) MeterProvider() *otelsdkmetric.MeterProvider {
	return p.meterProvider
}

// Shutdown 关闭所有提供者
func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error

	// 逆序关闭，确保依赖关系正确
	for i := len(p.shutdownFuncs) - 1; i >= 0; i-- {
		if err := p.shutdownFuncs[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to shutdown providers: %v", errs)
	}

	log.Println("All OpenTelemetry providers shutdown successfully")

	return nil
}
