package oteltrace

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	// InstrumentationName 用于创建 tracer 的 instrumentation name
	InstrumentationName = "bkn-backend/otel"

	KEY_HTTP_URL                    = "http.url"
	KEY_HTTP_METHOD                 = "http.method"
	KEY_HTTP_HEADER_METHOD_OVERRIDE = "http.header.X-Http-Method-Override"
	KEY_HTTP_HEADER_X_LANGUAGE      = "http.header.X-Language"
	KEY_HTTP_HEADER_CONTENT_TYPE    = "http.header.Content-Type"
	KEY_HTTP_HEADER_USER_AGENT      = "http.header.User-Agent"
	KEY_HTTP_HEADER_USERID          = "http.header.Userid"
	KEY_HTTP_STATUS                 = "http.status"
	KEY_HTTP_ERROR_CODE             = "http.error_code"
	KEY_HTTP_ROUTE                  = "http.route"
	KEY_HTTP_CLIENT_IP              = "http.client_ip"

	CONTENT_TYPE_NAME = "Content-Type"
	CONTENT_TYPE_JSON = "application/json"

	HTTP_HEADER_FORWARDED_FOR   = "X-Forwarded-For"
	HTTP_HEADER_METHOD_OVERRIDE = "X-Http-Method-Override"
	HTTP_HEADER_X_LANGUAGE      = "X-Language"
	HTTP_HEADER_USER_AGENT      = "User-Agent"
	HTTP_HEADER_USERID          = "Userid"
)

// TraceAttrs HTTP 请求相关属性的汇总，用于埋点。
type TraceAttrs struct {
	HttpUrl            string
	HttpMethod         string
	HttpMethodOverride string
	HttpXLanguage      string
	HttpContentType    string
	HttpUserAgent      string
	HttpUserID         string
	HttpRoute          string
	HttpClientIP       string
}

// GetAttrsByGinCtx 从 *gin.Context 抽取 TraceAttrs。
func GetAttrsByGinCtx(c *gin.Context) TraceAttrs {
	return TraceAttrs{
		HttpUrl:            fmt.Sprintf("http://%s%s", c.Request.Host, c.Request.RequestURI),
		HttpMethod:         c.Request.Method,
		HttpContentType:    c.GetHeader(CONTENT_TYPE_NAME),
		HttpMethodOverride: c.GetHeader(HTTP_HEADER_METHOD_OVERRIDE),
		HttpXLanguage:      c.GetHeader(HTTP_HEADER_X_LANGUAGE),
		HttpUserAgent:      c.GetHeader(HTTP_HEADER_USER_AGENT),
		HttpUserID:         c.GetHeader(HTTP_HEADER_USERID),
		HttpRoute:          c.FullPath(),
		HttpClientIP:       serverClientIP(c.GetHeader(HTTP_HEADER_FORWARDED_FOR)),
	}
}

// AddHttpAttrs4API 设置 API 入口 span 的 HTTP 属性。
func AddHttpAttrs4API(span trace.Span, attrs TraceAttrs) {
	span.SetAttributes(
		attr.Key(KEY_HTTP_URL).String(attrs.HttpUrl),
		attr.Key(KEY_HTTP_METHOD).String(attrs.HttpMethod),
		attr.Key(KEY_HTTP_HEADER_CONTENT_TYPE).String(attrs.HttpContentType),
		attr.Key(KEY_HTTP_HEADER_X_LANGUAGE).String(attrs.HttpXLanguage),
		attr.Key(KEY_HTTP_HEADER_USER_AGENT).String(attrs.HttpUserAgent),
		attr.Key(KEY_HTTP_HEADER_USERID).String(attrs.HttpUserID),
		attr.Key(KEY_HTTP_ROUTE).String(attrs.HttpRoute),
		attr.Key(KEY_HTTP_CLIENT_IP).String(attrs.HttpClientIP),
	)
	if attrs.HttpMethodOverride != "" {
		span.SetAttributes(
			attr.Key(KEY_HTTP_HEADER_METHOD_OVERRIDE).String(attrs.HttpMethodOverride),
		)
	}
}

// AddHttpAttrs4Error 设置错误状态到 span。
func AddHttpAttrs4Error(span trace.Span, status int, errorCode string, statusDescription string) {
	span.SetAttributes(
		attr.Key(KEY_HTTP_STATUS).Int(status),
		attr.Key(KEY_HTTP_ERROR_CODE).String(errorCode),
	)
	span.SetStatus(codes.Error, statusDescription)
}

// AddHttpAttrs4HttpError 从 *rest.HTTPError 设置错误到 span。
func AddHttpAttrs4HttpError(span trace.Span, err *rest.HTTPError) {
	span.SetAttributes(
		attr.Key(KEY_HTTP_STATUS).Int(err.HTTPCode),
		attr.Key(KEY_HTTP_ERROR_CODE).String(err.BaseError.ErrorCode),
	)
	span.SetStatus(codes.Error, fmt.Sprintf("%v", err.BaseError.ErrorDetails))
}

// AddHttpAttrs4Ok 设置成功状态到 span。
func AddHttpAttrs4Ok(span trace.Span, status int) {
	span.SetAttributes(attr.Key(KEY_HTTP_STATUS).Int(status))
	span.SetStatus(codes.Ok, "")
}

// AddAttrs4InternalHttp 设置内部 http 调用 span 的属性。
func AddAttrs4InternalHttp(span trace.Span, attrs TraceAttrs) {
	span.SetAttributes(
		attr.Key(KEY_HTTP_URL).String(attrs.HttpUrl),
		attr.Key(KEY_HTTP_METHOD).String(attrs.HttpMethod),
		attr.Key(KEY_HTTP_HEADER_CONTENT_TYPE).String(attrs.HttpContentType),
	)
	if attrs.HttpMethodOverride != "" {
		span.SetAttributes(attr.Key(KEY_HTTP_HEADER_METHOD_OVERRIDE).String(attrs.HttpMethodOverride))
	}
}

// serverClientIP 从 X-Forwarded-For 头取第一段。
func serverClientIP(xForwardedFor string) string {
	if idx := strings.Index(xForwardedFor, ","); idx >= 0 {
		xForwardedFor = xForwardedFor[:idx]
	}
	return xForwardedFor
}

// StartInternalSpan 服务内函数调用创建 span，自动从 runtime.Caller 获取 span name。
func StartInternalSpan(ctx context.Context) (context.Context, trace.Span) {
	name, filepath := callerFuncName(2)
	newCtx, span := StartNamedInternalSpan(ctx, name)
	if filepath != "" {
		span.SetAttributes(attr.String("code.filepath", filepath))
	}
	return newCtx, span
}

// StartClientSpan 外部依赖调用创建 span，自动从 runtime.Caller 获取 span name。
func StartClientSpan(ctx context.Context) (context.Context, trace.Span) {
	name, filepath := callerFuncName(2)
	newCtx, span := StartNamedClientSpan(ctx, name)
	if filepath != "" {
		span.SetAttributes(attr.String("code.filepath", filepath))
	}
	return newCtx, span
}

func callerFuncName(skip int) (string, string) {
	pc, file, lineNo, ok := runtime.Caller(skip)
	if !ok {
		return "unknown", ""
	}
	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	return funcPaths[len(funcPaths)-1], fmt.Sprintf("%s:%v", file, lineNo)
}

// StartNamedClientSpan用自定义业务名创建 SpanKindClient 类型 span。
func StartNamedClientSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer(InstrumentationName).Start(ctx, name, trace.WithSpanKind(trace.SpanKindClient))
}

// StartNamedInternalSpan 用自定义业务名创建 SpanKindInternal 类型 span。
func StartNamedInternalSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer(InstrumentationName).Start(ctx, name, trace.WithSpanKind(trace.SpanKindInternal))
}

// StartServerSpan 跨服务（HTTP 接口）创建 span。
// 前置依赖：TracingMiddleware 已把 trace header 提取到 c.Request.Context()，
//          LanguageMiddleware 已把 language 叠加到 c.Request.Context()。
func StartServerSpan(c *gin.Context) (context.Context, trace.Span) {
	spanName := fmt.Sprintf("%s %s", c.Request.Method, c.FullPath())
	newCtx, span := otel.Tracer(InstrumentationName).Start(c.Request.Context(), spanName, trace.WithSpanKind(trace.SpanKindServer))
	span.SetAttributes(
		attr.String("http.request.method", c.Request.Method),
		attr.String("http.route", c.FullPath()),
		attr.String("client.address", c.ClientIP()),
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
func SetAttributes(ctx context.Context, kv ...attr.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(kv...)
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
