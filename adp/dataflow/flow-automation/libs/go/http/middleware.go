package http

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/i18n"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// MiddlewareTrace 接口层链路追踪，设置span
func MiddlewareTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.Contains(c.Request.URL.Path, "/api/automation/v1/health") {
			c.Next()
			return
		}

		newCtx, span := trace.StartServerSpan(c)
		defer span.End()
		c.Request = c.Request.WithContext(newCtx)

		c.Next()

		status := c.Writer.Status()
		if status/100 >= 4 {
			span.SetStatus(codes.Error, "REQUEST FAILED")
		} else {
			span.SetStatus(codes.Ok, "OK")
		}
		if status > 0 {
			span.SetAttributes(semconv.HTTPStatusCode(status))
		}
	}
}

// LanguageMiddleware 全局Set language到ctx
func LanguageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从header获取x-language，默认为zh-CN
		language := i18n.Language(c.GetHeader(i18n.XLangKey))
		if language == "" {
			language = i18n.DefaultLanguage
		}

		newCtx := context.WithValue(c.Request.Context(), i18n.XLangKey, language)
		// 设置到gin的context中
		c.Request = c.Request.WithContext(newCtx)

		// 继续处理请求
		c.Next()
	}
}
