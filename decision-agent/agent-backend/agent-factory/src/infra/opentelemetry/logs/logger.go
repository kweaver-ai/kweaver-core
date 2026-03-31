package logs

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/embedded"
	"go.opentelemetry.io/otel/trace"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
)

// Logger OpenTelemetry 日志记录器
type Logger struct {
	config   *conf.OtelConfig
	logger   otellog.Logger
	logLevel otellog.Severity
}

// NewLogger 创建新的日志记录器
func NewLogger(cfg *conf.OtelConfig, logger otellog.Logger) (*Logger, error) {
	// 解析日志级别
	logLevel, err := parseLogLevel(cfg.Log.Level)
	if err != nil {
		return nil, fmt.Errorf("failed to parse log level: %w", err)
	}

	// 如果logger为nil，使用no-op logger
	if logger == nil {
		logger = newNoopLogger()
	}

	return &Logger{
		config:   cfg,
		logger:   logger,
		logLevel: logLevel,
	}, nil
}

// parseLogLevel 解析日志级别字符串为 Severity
func parseLogLevel(level string) (otellog.Severity, error) {
	switch level {
	case "debug":
		return otellog.SeverityDebug, nil
	case "info":
		return otellog.SeverityInfo, nil
	case "warn":
		return otellog.SeverityWarn, nil
	case "error":
		return otellog.SeverityError, nil
	default:
		return otellog.SeverityInfo, fmt.Errorf("invalid log level: %s", level)
	}
}

// Debug 记录调试日志
func (l *Logger) Debug(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	if l.logLevel <= otellog.SeverityDebug {
		l.emit(ctx, otellog.SeverityDebug, msg, attrs...)
	}
}

// Info 记录信息日志
func (l *Logger) Info(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	if l.logLevel <= otellog.SeverityInfo {
		l.emit(ctx, otellog.SeverityInfo, msg, attrs...)
	}
}

// Warn 记录警告日志
func (l *Logger) Warn(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	if l.logLevel <= otellog.SeverityWarn {
		l.emit(ctx, otellog.SeverityWarn, msg, attrs...)
	}
}

// Error 记录错误日志
func (l *Logger) Error(ctx context.Context, msg string, attrs ...attribute.KeyValue) {
	if l.logLevel <= otellog.SeverityError {
		l.emit(ctx, otellog.SeverityError, msg, attrs...)
	}
}

// emit 发送日志记录
func (l *Logger) emit(ctx context.Context, severity otellog.Severity, msg string, attrs ...attribute.KeyValue) {
	// 获取当前 span context
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	// 创建日志记录
	record := otellog.Record{}
	record.SetTimestamp(time.Now())
	record.SetObservedTimestamp(time.Now())
	record.SetSeverity(severity)
	record.SetSeverityText(severity.String())
	record.SetBody(otellog.StringValue(msg))

	// 转换属性为 log.KeyValue
	var logAttrs []otellog.KeyValue
	for _, attr := range attrs {
		logAttrs = append(logAttrs, otellog.KeyValueFromAttribute(attr))
	}

	// 设置 trace 和 span ID
	if spanContext.HasTraceID() {
		logAttrs = append(logAttrs, otellog.String("trace_id", spanContext.TraceID().String()))
	}

	if spanContext.HasSpanID() {
		logAttrs = append(logAttrs, otellog.String("span_id", spanContext.SpanID().String()))
	}

	// 添加服务信息
	logAttrs = append(logAttrs,
		otellog.String("service.name", l.config.ServiceName),
		otellog.String("service.version", l.config.ServiceVersion),
	)

	// 添加所有属性到记录
	for _, attr := range logAttrs {
		record.AddAttributes(attr)
	}

	// 发送日志记录
	l.logger.Emit(ctx, record)
}

// WithContext 从上下文创建带有 trace 信息的日志记录器
func (l *Logger) WithContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger: l,
		ctx:    ctx,
	}
}

// ContextLogger 带有上下文的日志记录器
type ContextLogger struct {
	logger *Logger
	ctx    context.Context
}

// defaultNoopLogger 默认的no-op logger实例
var defaultNoopLogger *Logger

func init() {
	// 创建默认配置
	cfg := &conf.OtelConfig{
		ServiceName:    "noop",
		ServiceVersion: "1.0.0",
		Log: conf.LogConfig{
			Level: "info",
		},
	}
	// 忽略错误，因为配置是有效的
	defaultNoopLogger, _ = NewLogger(cfg, newNoopLogger())
}

// Debug 记录调试日志（带上下文）
func (cl *ContextLogger) Debug(msg string, attrs ...attribute.KeyValue) {
	cl.logger.Debug(cl.ctx, msg, attrs...)
}

// Info 记录信息日志（带上下文）
func (cl *ContextLogger) Info(msg string, attrs ...attribute.KeyValue) {
	cl.logger.Info(cl.ctx, msg, attrs...)
}

// Warn 记录警告日志（带上下文）
func (cl *ContextLogger) Warn(msg string, attrs ...attribute.KeyValue) {
	cl.logger.Warn(cl.ctx, msg, attrs...)
}

// Error 记录错误日志（带上下文）
func (cl *ContextLogger) Error(msg string, attrs ...attribute.KeyValue) {
	cl.logger.Error(cl.ctx, msg, attrs...)
}

// GetGlobalLogger 获取全局日志记录器（简化使用）
func GetGlobalLogger(cfg *conf.OtelConfig, logger otellog.Logger) (*Logger, error) {
	return NewLogger(cfg, logger)
}

// LogHTTPRequest 记录 HTTP 请求日志
func (l *Logger) LogHTTPRequest(ctx context.Context, method, path, clientIP string, statusCode int, duration time.Duration) {
	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.path", path),
		attribute.String("http.client_ip", clientIP),
		attribute.Int("http.status_code", statusCode),
		attribute.Int64("http.duration_ms", duration.Milliseconds()),
	}

	// 根据状态码选择日志级别
	if statusCode >= 400 && statusCode < 500 {
		l.Warn(ctx, "HTTP client error", attrs...)
	} else if statusCode >= 500 {
		l.Error(ctx, "HTTP server error", attrs...)
	} else {
		l.Info(ctx, "HTTP request completed", attrs...)
	}
}

// LogBusinessEvent 记录业务事件日志
func (l *Logger) LogBusinessEvent(ctx context.Context, eventType, eventName, description string, additionalAttrs ...attribute.KeyValue) {
	attrs := []attribute.KeyValue{
		attribute.String("event.type", eventType),
		attribute.String("event.name", eventName),
		attribute.String("event.description", description),
	}

	// 添加额外属性
	attrs = append(attrs, additionalAttrs...)

	l.Info(ctx, fmt.Sprintf("Business event: %s", eventName), attrs...)
}

// WithLogger 将logger注入context
func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, constant.LoggerKey, logger)
}

// LoggerFromContext 从context获取logger
func LoggerFromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(constant.LoggerKey).(*Logger); ok {
		return logger
	}
	// 如果没有找到logger，返回默认的no-op logger
	return defaultNoopLogger
}

// noopLogger 是一个no-op实现，用于当OpenTelemetry logger未启用时
type noopLogger struct {
	embedded.Logger // 嵌入以实现Logs API Logger接口
}

// Emit 实现otellog.Logger接口，不执行任何操作
func (n *noopLogger) Emit(ctx context.Context, record otellog.Record) {
	// No-op
}

// Enabled 实现otellog.Logger接口，总是返回false
func (n *noopLogger) Enabled(ctx context.Context, params otellog.EnabledParameters) bool {
	return false
}

// newNoopLogger 创建新的no-op logger
func newNoopLogger() otellog.Logger {
	return &noopLogger{}
}
