package trace

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/public"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/common"
	"github.com/labstack/gommon/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
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

// StartInternalSpan 内部方法调用
func StartInternalSpan(ctx context.Context) (context.Context, trace.Span) {
	pc, file, linkNo, ok := runtime.Caller(1)
	if !ok {
		log.Error("start span error")
		newCtx, span := ar_trace.Tracer.Start(ctx, "unknow", trace.WithSpanKind(trace.SpanKindInternal))
		return newCtx, span
	}
	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := ar_trace.Tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.String(FUNC_PATH, fmt.Sprintf("%s:%v", file, linkNo)))
	return newCtx, span
}

// StartServerSpan 接口层调用时使用
func StartServerSpan(ctx *gin.Context) (context.Context, trace.Span) {
	newCtx, span := ar_trace.Tracer.Start(ctx.Request.Context(), ctx.FullPath(), trace.WithSpanKind(trace.SpanKindServer))
	span.SetAttributes(attribute.String(HTTP_METHOD, ctx.Request.Method))
	span.SetAttributes(attribute.String(HTTP_ROUTE, ctx.FullPath()))
	span.SetAttributes(attribute.String(HTTP_CLIENT_IP, ctx.ClientIP()))
	return newCtx, span
}

// StartConsumerSpan 消费者消费消息时记录使用
func StartConsumerSpan(ctx context.Context) (context.Context, trace.Span) {
	pc, file, linkNo, ok := runtime.Caller(1)
	if !ok {
		log.Error("start span error")
		newCtx, span := ar_trace.Tracer.Start(ctx, "unknow", trace.WithSpanKind(trace.SpanKindConsumer))
		return newCtx, span
	}

	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := ar_trace.Tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindConsumer))
	span.SetAttributes(attribute.String(FUNC_PATH, fmt.Sprintf("%s:%v", file, linkNo)))

	return newCtx, span
}

// StartProducerSpan 生产者生产消息时记录使用
func StartProducerSpan(ctx context.Context) (context.Context, trace.Span) {
	pc, file, linkNo, ok := runtime.Caller(1)
	if !ok {
		log.Error("start span error")
		newCtx, span := ar_trace.Tracer.Start(ctx, "unknow", trace.WithSpanKind(trace.SpanKindProducer))
		return newCtx, span
	}

	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := ar_trace.Tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindProducer))
	span.SetAttributes(attribute.String(FUNC_PATH, fmt.Sprintf("%s:%v", file, linkNo)))

	return newCtx, span
}

// EndSpan 关闭span
func EndSpan(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	TelemetrySpanEnd(span, err)
}

// TelemetrySpanEnd 关闭span
func TelemetrySpanEnd(span trace.Span, err error) {
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

// SetAttributes 设置attribute
func SetAttributes(ctx context.Context, kv ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(kv...)
}

func SetTraceExporter(conf *common.TelemetryConf) *sdktrace.TracerProvider {
	var tracerProvider *sdktrace.TracerProvider
	if conf.TraceEnabled {
		var traceClient public.Client
		if conf.TraceURL == "" { // 没上报地址就 输出到标准输出, 一般用于无环境调试
			traceClient = public.NewConsoleClient()
		} else {
			traceClient = public.NewHTTPClient(public.WithAnyRobotURL(conf.TraceURL))
		}

		traceExporter := ar_trace.NewExporter(traceClient)
		hostName, _ := os.Hostname()
		public.SetServiceInfo(conf.ServerName, conf.ServerVersion, hostName)
		tracerProvider = sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExporter), sdktrace.WithResource(ar_trace.TraceResource()))
		otel.SetTracerProvider(tracerProvider)
	}
	return tracerProvider
}

func ExitTraceExporter(ctx context.Context, tracerProvider *sdktrace.TracerProvider) {
	if tracerProvider != nil {
		_ = tracerProvider.ForceFlush(ctx)
		if err := tracerProvider.Shutdown(ctx); err != nil {
			commonLog.NewLogger().Errorf("Server forced to ExitTraceExporter: %s", err.Error())
		}
		commonLog.NewLogger().Infof("Trace exporter normal exited")
	}
}
