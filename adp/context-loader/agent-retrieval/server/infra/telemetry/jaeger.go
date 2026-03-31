// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package telemetry

import (
	"context"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// newJaegerOTLPGRPCExporter 使用 OTLP gRPC 连接 Jaeger
func newJaegerOTLPGRPCExporter(endpoint string) (*otlptrace.Exporter, error) {
	// Jaeger OTLP gRPC 默认端口: 4317
	exporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint(endpoint), // 例如: "localhost:4317"
		otlptracegrpc.WithInsecure(),
	)
	return exporter, err
}

// newJaegerOTLPHTTPExporter 使用 OTLP HTTP 连接 Jaeger
func newJaegerOTLPHTTPExporter(endpoint string) (*otlptrace.Exporter, error) {
	// Jaeger OTLP HTTP 默认端口: 4318
	exporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(endpoint), // 例如: "localhost:4318"
		otlptracehttp.WithInsecure(),
	)
	return exporter, err
}

// InitJaegerExporter 初始化 Jaeger 导出器
// serviceName 服务名称
// endpoint Jaeger 采集器地址
func InitJaegerExporter(serviceName, exporterType, endpoint string) (tp *sdktrace.TracerProvider, err error) {
	ctx := context.Background()
	var exporter *otlptrace.Exporter
	// 创建 OTLP exporter
	switch exporterType {
	case "grpc":
		exporter, err = newJaegerOTLPGRPCExporter(endpoint)
	case "http":
		exporter, err = newJaegerOTLPHTTPExporter(endpoint)
	default:
		exporter, err = otlptracehttp.New(
			ctx,
			otlptracehttp.WithEndpoint(endpoint),
			otlptracehttp.WithInsecure(),
		)
	}
	if err != nil {
		return nil, err
	}
	// 创建资源
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
			attribute.String("environment", "development"),
			attribute.String("language", "Go"),
			attribute.String("go_version", runtime.Version()),
			attribute.String("go_os", runtime.GOOS),
			attribute.String("go_arch", runtime.GOARCH),
		),
	)
	if err != nil {
		return nil, err
	}
	// 创建 TracerProvider
	tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	// 设置全局 TracerProvider
	otel.SetTracerProvider(tp)
	return tp, nil
}
