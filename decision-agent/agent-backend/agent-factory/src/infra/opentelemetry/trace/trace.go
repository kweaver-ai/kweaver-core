package trace

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/appruntime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	HTTP_METHOD    = "http.method"
	HTTP_ROUTE     = "http.route"
	HTTP_CLIENT_IP = "http.client_ip"
	FUNC_PATH      = "func.path"
	DB_QUERY       = "db.query"
	TABLE_NAME     = "table.name"
	DB_SQL         = "db.sql"
	DB_Values      = "db.values"
)

// 从 HTTP Header 中提取 Trace 上下文
func ExtractTraceHeader(ctx context.Context, header http.Header) context.Context {
	if header == nil {
		return ctx
	}

	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(header))
}

// 服务内函数调用创建 span
func StartInternalSpan(ctx context.Context) (context.Context, trace.Span) {
	pc, file, linkNo, ok := runtime.Caller(1)
	if !ok {
		newCtx, span := otel.Tracer(appruntime.TraceInstrumentationName).Start(ctx, "unknow", trace.WithSpanKind(trace.SpanKindInternal))
		return newCtx, span
	}

	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := otel.Tracer(appruntime.TraceInstrumentationName).Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.String(FUNC_PATH, fmt.Sprintf("%s:%v", file, linkNo)))

	return newCtx, span
}

// 跨服务（接口）创建 span
func StartServerSpan(ctx *gin.Context) (context.Context, trace.Span) {
	newCtx := ExtractTraceHeader(ctx.Request.Context(), ctx.Request.Header)
	newCtx, span := otel.Tracer(appruntime.TraceInstrumentationName).Start(newCtx, ctx.FullPath(), trace.WithSpanKind(trace.SpanKindServer))
	span.SetAttributes(attribute.String(HTTP_METHOD, ctx.Request.Method))
	span.SetAttributes(attribute.String(HTTP_ROUTE, ctx.FullPath()))
	span.SetAttributes(attribute.String(HTTP_CLIENT_IP, ctx.ClientIP()))

	return newCtx, span
}

// 设置attribute
func SetAttributes(ctx context.Context, kv ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(kv...)
}

// 关闭span
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
