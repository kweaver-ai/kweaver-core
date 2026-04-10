package oteltrace

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/otelconst"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	// InstrumentationName 用于创建 tracer 的 instrumentation name
	InstrumentationName = "agent-factory/otel"
)

// StartInternalSpan 服务内函数调用创建 span，自动从 runtime.Caller 获取 span name。
func StartInternalSpan(ctx context.Context) (context.Context, trace.Span) {
	pc, file, lineNo, ok := runtime.Caller(1)
	if !ok {
		newCtx, span := otel.Tracer(InstrumentationName).Start(ctx, "unknown", trace.WithSpanKind(trace.SpanKindInternal))

		return newCtx, span
	}

	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := otel.Tracer(InstrumentationName).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.String("code.filepath", fmt.Sprintf("%s:%v", file, lineNo)))

	return newCtx, span
}

// StartInvokeAgentSpan 创建 invoke_agent span，按照 OTel Gen AI Agent Spans 规范。
func StartInvokeAgentSpan(ctx context.Context, agentName string) (context.Context, trace.Span) {
	spanName := "invoke_agent"
	if agentName != "" {
		spanName = fmt.Sprintf("invoke_agent %s", agentName)
	}

	newCtx, span := otel.Tracer(InstrumentationName).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))

	return newCtx, span
}

// StartServerSpan 跨服务（HTTP 接口）创建 span，从请求头提取 trace 上下文。
func StartServerSpan(c *gin.Context) (context.Context, trace.Span) {
	newCtx := ExtractTraceHeader(c.Request.Context(), c.Request.Header)
	spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
	newCtx, span := otel.Tracer(InstrumentationName).Start(newCtx, spanName, trace.WithSpanKind(trace.SpanKindServer))
	span.SetAttributes(
		attribute.String("http.request.method", c.Request.Method),
		attribute.String("http.route", c.FullPath()),
		attribute.String("client.address", c.ClientIP()),
	)

	return newCtx, span
}

// ExtractTraceHeader 从 HTTP Header 中提取 Trace 上下文。
func ExtractTraceHeader(ctx context.Context, header http.Header) context.Context {
	if header == nil {
		return ctx
	}

	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
}

// SetAttributes 在当前 span 上设置属性。
func SetAttributes(ctx context.Context, kv ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(kv...)
}

// SetConversationID 在当前 span 上设置标准会话属性。
// 只有拿到真实 conversationID 时才写入，避免空值污染链路。
func SetConversationID(ctx context.Context, conversationID string) {
	if conversationID == "" {
		return
	}

	SetAttributes(ctx, attribute.String(otelconst.AttrGenAIConversationID, conversationID))
}

// EndSpan 结束当前 span，如有错误则记录。
func EndSpan(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "OK")
	}

	span.End()
}
