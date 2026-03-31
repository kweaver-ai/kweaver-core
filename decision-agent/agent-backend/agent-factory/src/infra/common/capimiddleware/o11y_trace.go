package capimiddleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

/*
	我们在跨服务调用时一般使用otelgin中的Middleware去获取上游调用者的SpanContext, 避免链路断开.
	但它会产生一条Span, 且不可显示设置Span的Status.
	为克服这一缺陷, 所以我们基于otelgin, 实现了TracingMiddleware, 以获取上游调用者的SpanContext.
	otelgin: https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/instrumentation/github.com/gin-gonic/gin/otelgin/gintrace.go
*/

func O11yTraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		savedCtx := c.Request.Context()
		defer func() {
			c.Request = c.Request.WithContext(savedCtx)
		}()

		newCtx, _ := o11y.StartServerSpan(c)

		// pass the span through the request context
		c.Request = c.Request.WithContext(newCtx)

		// serve the request to the next middleware
		c.Next()

		status := c.Writer.Status()
		o11y.SetAttributes(newCtx, semconv.HTTPStatusCode(status))

		if status/100 >= 4 {
			o11y.EndSpan(newCtx, fmt.Errorf("request failed"))
		} else {
			o11y.EndSpan(newCtx, nil)
		}
	}
}
