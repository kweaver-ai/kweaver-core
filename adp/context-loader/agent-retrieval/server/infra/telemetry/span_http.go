// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var clearUserPassRe = regexp.MustCompile(`(://)[^/]*@`)

const (
	// maxHeaderLogSize 请求头日志记录最大字节数
	maxHeaderLogSize = 4096
	// maxBodyLogSize 请求体日志记录最大字节数
	maxBodyLogSize = 2048
)

// HTTPRequest 发起HTTP请求
func HTTPRequest(ctx context.Context, req *http.Request, fn func(req *http.Request) (*http.Response, error)) (rsp *http.Response, err error) {
	tracer := otel.GetTracerProvider()
	if tracer != nil {
		var span trace.Span
		ctx, span = o11y.GlobalTracer().Start(ctx, "http.request", trace.WithSpanKind(trace.SpanKindClient))
		arr := strings.Split(req.Proto, "/")
		span.SetAttributes(attribute.Key("net.protocol.name").String(arr[0]))
		if len(arr) > 1 {
			span.SetAttributes(attribute.Key("net.protocol.version").String(arr[1]))
		}
		span.SetAttributes(attribute.Key("http.method").String(req.Method))
		span.SetAttributes(attribute.Key("http.request_content_length").Int64(req.ContentLength))
		span.SetAttributes(attribute.Key("http.url").String(clearUserPassRe.ReplaceAllString(req.URL.String(), "$1")))
		span.SetAttributes(attribute.Key("net.peer.name").String(req.URL.Hostname()))
		span.SetAttributes(attribute.Key("net.peer.port").String(req.URL.Port()))

		// 记录查询参数
		if req.URL.RawQuery != "" {
			span.SetAttributes(attribute.Key("http.query_params").String(req.URL.RawQuery))
		}

		// 记录请求头
		if req.Header != nil {
			headerBytes, _ := json.Marshal(req.Header)
			headerStr := string(headerBytes)
			if len(headerStr) > maxHeaderLogSize {
				headerStr = headerStr[:maxHeaderLogSize] + "...[truncated]"
			}
			span.SetAttributes(attribute.Key("http.request_headers").String(headerStr))
		}

		// 记录请求体
		if req.Body != nil && req.ContentLength > 0 && req.ContentLength < 1024*1024 {
			bodyBytes, _ := io.ReadAll(req.Body)
			if len(bodyBytes) > 0 {
				bodyStr := string(bodyBytes)
				if len(bodyStr) > maxBodyLogSize {
					bodyStr = bodyStr[:maxBodyLogSize] + "...[truncated]"
				}
				span.SetAttributes(attribute.Key("http.request_body").String(bodyStr))
				// 重置Body以便后续读取
				req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		otel.GetTextMapPropagator().Inject(trace.ContextWithSpan(ctx, span), propagation.HeaderCarrier(req.Header))
		req = req.WithContext(ctx)
		defer func() {
			if rsp != nil {
				span.SetAttributes(attribute.Key("http.status_code").Int(rsp.StatusCode))
				span.SetAttributes(attribute.Key("http.response_content_length").Int64(rsp.ContentLength))
			}
			// 400以上的错误记录到trace中
			e := err
			if e == nil {
				e = recordHTTPErrorBody(rsp)
			}
			o11y.TelemetrySpanEnd(span, e)
		}()
	}
	rsp, err = fn(req)
	return
}

func recordHTTPErrorBody(rsp *http.Response) (err error) {
	// 只记录 400以上错误
	if rsp.StatusCode < http.StatusBadRequest {
		return
	}
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return
	}
	rsp.Body = io.NopCloser(bytes.NewBuffer(body)) // 将body重新赋值给rsp.Body，后续可以读取
	limitBody := body
	err = errors.New(string(limitBody))
	return
}

// BuildUpOperateName 获取TraceOperate名称
func BuildUpOperateName(ops ...string) string {
	return strings.Join(ops, ".")
}
