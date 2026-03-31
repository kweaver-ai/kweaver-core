package capimiddleware

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RequestInfo struct {
	Timestamp time.Time   `json:"timestamp"`
	Method    string      `json:"method"`
	URL       string      `json:"url"`
	Headers   http.Header `json:"headers"`
	Body      string      `json:"body,omitempty"`
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 读取并复制请求体
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}

		// 恢复请求体供后续使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		//jsonData, err := json.MarshalIndent(bodyBytes, "", "  ")
		//if err != nil {
		//	return
		//}
		// 收集请求信息
		reqInfo := RequestInfo{
			URL:       c.Request.URL.String(),
			Method:    c.Request.Method,
			Timestamp: start,
			// Headers:   c.Request.Header.Clone(),
			Body: string(bodyBytes),
		}

		// 打印日志（可替换为其他日志输出方式）
		log.Printf("[INFO]Request: %+v", reqInfo)

		// 继续处理请求
		c.Next()

		// 可选：记录处理时间
		log.Printf("[INFO]Request processed in %v", time.Since(start))
	}
}
