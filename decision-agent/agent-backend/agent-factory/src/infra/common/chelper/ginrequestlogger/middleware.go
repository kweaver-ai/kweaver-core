package ginrequestlogger

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware 返回gin中间件
func (r *RequestLogger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !r.logger.IsEnabled() {
			c.Next()
			return
		}

		start := time.Now()

		// 1. 读取并复制请求体
		var reqBodyBytes []byte
		if c.Request.Body != nil {
			reqBodyBytes, _ = io.ReadAll(c.Request.Body)
			// 恢复请求体供后续使用
			c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBodyBytes))
		}

		// 2. 包装ResponseWriter以捕获响应体
		respBodyWriter := &ResponseBodyWriter{
			ResponseWriter: c.Writer,
			Body:           bytes.NewBufferString(""),
		}
		c.Writer = respBodyWriter

		// 3. 继续处理请求
		c.Next()

		// 4. 计算耗时
		duration := time.Since(start)

		// 5. 记录日志
		r.logger.LogRequest(
			c.Request.Context(),
			c.Request,
			string(reqBodyBytes),
			c.Writer.Status(),
			c.Writer.Header(),
			respBodyWriter.Body.String(),
			duration,
		)
	}
}

// DefaultMiddleware 返回默认请求日志记录器的中间件
// 如果未初始化，则panic
func DefaultMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if defaultRequestLogger == nil {
			panic("ginrequestlogger: defaultRequestLogger is not initialized, please call InitDefaultRequestLogger first")
		}

		defaultRequestLogger.Middleware()(c)
	}
}
