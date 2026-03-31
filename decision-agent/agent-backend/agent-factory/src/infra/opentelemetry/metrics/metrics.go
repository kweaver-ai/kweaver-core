package metrics

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
)

// Metrics 指标记录器
type Metrics struct {
	config *conf.OtelConfig
	meter  metric.Meter

	// HTTP 指标
	httpRequestDuration metric.Float64Histogram
	httpRequestCount    metric.Int64Counter
	httpRequestSize     metric.Int64Histogram
	httpResponseSize    metric.Int64Histogram

	// 性能指标
	processingTime metric.Float64Histogram
}

// NewMetrics 创建新的指标记录器
func NewMetrics(cfg *conf.OtelConfig, meter metric.Meter) (*Metrics, error) {
	m := &Metrics{
		config: cfg,
		meter:  meter,
	}

	// 初始化指标
	if err := m.initMetrics(); err != nil {
		return nil, fmt.Errorf("failed to init metrics: %w", err)
	}

	return m, nil
}

// initMetrics 初始化所有指标
func (m *Metrics) initMetrics() error {
	// 如果meter为nil，跳过指标初始化（no-op模式）
	if m.meter == nil {
		return nil
	}

	var err error

	// HTTP 指标
	m.httpRequestDuration, err = m.meter.Float64Histogram(
		"http.request.duration",
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return fmt.Errorf("failed to create http.request.duration metric: %w", err)
	}

	m.httpRequestCount, err = m.meter.Int64Counter(
		"http.request.count",
		metric.WithDescription("Count of HTTP requests"),
	)
	if err != nil {
		return fmt.Errorf("failed to create http.request.count metric: %w", err)
	}

	m.httpRequestSize, err = m.meter.Int64Histogram(
		"http.request.size",
		metric.WithDescription("Size of HTTP requests"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return fmt.Errorf("failed to create http.request.size metric: %w", err)
	}

	m.httpResponseSize, err = m.meter.Int64Histogram(
		"http.response.size",
		metric.WithDescription("Size of HTTP responses"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return fmt.Errorf("failed to create http.response.size metric: %w", err)
	}

	// 性能指标
	m.processingTime, err = m.meter.Float64Histogram(
		"processing.time",
		metric.WithDescription("Processing time of operations"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		return fmt.Errorf("failed to create processing.time metric: %w", err)
	}

	return nil
}

// RecordHTTPRequest 记录 HTTP 请求指标
func (m *Metrics) RecordHTTPRequest(ctx context.Context, method, route string, statusCode int, durationMs float64, requestSize, responseSize int64) {
	// 如果指标未初始化（meter为nil），直接返回
	if m.httpRequestDuration == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.route", route),
		attribute.String("http.status_code", fmt.Sprintf("%d", statusCode)),
	}

	m.httpRequestDuration.Record(ctx, durationMs, metric.WithAttributes(attrs...))
	m.httpRequestCount.Add(ctx, 1, metric.WithAttributes(attrs...))
	m.httpRequestSize.Record(ctx, requestSize, metric.WithAttributes(attrs...))
	m.httpResponseSize.Record(ctx, responseSize, metric.WithAttributes(attrs...))
}

// RecordProcessingTime 记录处理时间指标
func (m *Metrics) RecordProcessingTime(ctx context.Context, operation string, durationMs float64) {
	// 如果指标未初始化（meter为nil），直接返回
	if m.processingTime == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("operation", operation),
	}
	m.processingTime.Record(ctx, durationMs, metric.WithAttributes(attrs...))
}

// GetGlobalMetrics 获取全局指标记录器（简化使用）
func GetGlobalMetrics(cfg *conf.OtelConfig, meter metric.Meter) (*Metrics, error) {
	return NewMetrics(cfg, meter)
}

// WithMetrics 将metrics注入context
func WithMetrics(ctx context.Context, metrics *Metrics) context.Context {
	return context.WithValue(ctx, constant.MetricsKey, metrics)
}

// MetricsFromContext 从context获取metrics
func MetricsFromContext(ctx context.Context) *Metrics {
	if metrics, ok := ctx.Value(constant.MetricsKey).(*Metrics); ok {
		return metrics
	}
	// 如果没有找到metrics，返回默认的no-op metrics
	return defaultNoopMetrics
}

// defaultNoopMetrics 默认的no-op metrics实例
var defaultNoopMetrics *Metrics

func init() {
	// 创建默认配置
	cfg := &conf.OtelConfig{
		ServiceName:    "noop",
		ServiceVersion: "1.0.0",
		Metric: conf.MetricConfig{
			Enabled:        false,
			Exporter:       "console",
			HTTPEndpoint:   "",
			ExportInterval: 30,
		},
	}
	// 使用nil meter创建metrics，这将创建一个no-op实例
	defaultNoopMetrics, _ = NewMetrics(cfg, nil)
}
