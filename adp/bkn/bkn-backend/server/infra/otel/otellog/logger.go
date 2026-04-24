package otellog

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"go.opentelemetry.io/otel/codes"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/trace"
)

// globalServiceName 全局服务名称，由 InitOTel 设置。
var globalServiceName string

// SetServiceName 设置全局服务名称。
func SetServiceName(name string) {
	globalServiceName = name
}

// LogDebug 发送 Debug 级别的结构化日志，自动关联 trace 上下文；同时写入 zap stdout。
func LogDebug(ctx context.Context, message string, attrs ...otellog.KeyValue) {
	emitLog(ctx, otellog.SeverityDebug, message, attrs...)
	logger.Debug(formatForStdout(ctx, message, attrs))
}

// LogInfo 发送 Info 级别的结构化日志，自动关联 trace 上下文；同时写入 zap stdout。
func LogInfo(ctx context.Context, message string, attrs ...otellog.KeyValue) {
	emitLog(ctx, otellog.SeverityInfo, message, attrs...)
	logger.Info(formatForStdout(ctx, message, attrs))
}

// LogWarn 发送 Warn 级别的结构化日志，自动关联 trace 上下文；同时写入 zap stdout。
func LogWarn(ctx context.Context, message string, attrs ...otellog.KeyValue) {
	emitLog(ctx, otellog.SeverityWarn, message, attrs...)
	logger.Warn(formatForStdout(ctx, message, attrs))
}

// LogError 发送 Error 级别的结构化日志，在当前 span 记录错误；同时写入 zap stdout。
func LogError(ctx context.Context, message string, err error, attrs ...otellog.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, message)
	}

	allAttrs := baseLogAttributes(span)
	if err != nil {
		allAttrs = append(allAttrs, otellog.String("error.message", err.Error()))
	}
	allAttrs = append(allAttrs, attrs...)

	otelLogger := global.GetLoggerProvider().Logger(globalServiceName)

	record := otellog.Record{}
	record.SetTimestamp(time.Now())
	record.SetSeverity(otellog.SeverityError)
	record.SetSeverityText("ERROR")
	record.SetBody(otellog.StringValue(message))
	record.AddAttributes(allAttrs...)
	otelLogger.Emit(ctx, record)

	logger.Error(formatForStdout(ctx, message, allAttrs))
}

// emitLog 通用 OTel 日志发送函数。
func emitLog(ctx context.Context, severity otellog.Severity, message string, attrs ...otellog.KeyValue) {
	span := trace.SpanFromContext(ctx)

	allAttrs := baseLogAttributes(span)
	allAttrs = append(allAttrs, attrs...)

	otelLogger := global.GetLoggerProvider().Logger(globalServiceName)

	record := otellog.Record{}
	record.SetTimestamp(time.Now())
	record.SetSeverity(severity)
	record.SetSeverityText(severity.String())
	record.SetBody(otellog.StringValue(message))
	record.AddAttributes(allAttrs...)
	otelLogger.Emit(ctx, record)
}

// baseLogAttributes 构建基础日志属性，包含 trace 关联信息。
func baseLogAttributes(span trace.Span) []otellog.KeyValue {
	attrs := []otellog.KeyValue{
		otellog.String("service.name", globalServiceName),
	}

	spanCtx := span.SpanContext()
	if spanCtx.HasTraceID() {
		attrs = append(attrs, otellog.String("trace_id", spanCtx.TraceID().String()))
	}

	if spanCtx.HasSpanID() {
		attrs = append(attrs, otellog.String("span_id", spanCtx.SpanID().String()))
	}

	return attrs
}

// formatForStdout 拼装 stdout 日志行：[trace=... span=...] message k=v k=v
func formatForStdout(ctx context.Context, message string, attrs []otellog.KeyValue) string {
	var b strings.Builder
	sc := trace.SpanFromContext(ctx).SpanContext()
	if sc.HasTraceID() {
		fmt.Fprintf(&b, "[trace=%s span=%s] ", sc.TraceID(), sc.SpanID())
	}
	b.WriteString(message)
	for _, a := range attrs {
		fmt.Fprintf(&b, " %s=%s", a.Key, a.Value.String())
	}
	return b.String()
}
