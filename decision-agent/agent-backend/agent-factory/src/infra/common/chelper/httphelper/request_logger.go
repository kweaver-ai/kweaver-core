package httphelper

import (
	"context"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/net/gclient"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httprequesthelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

// requestLogger 请求日志记录器
var requestLogger *httprequesthelper.Logger

// EnableHTTPClientRequestLogging 启用HTTP客户端请求日志记录
func EnableHTTPClientRequestLogging(config *httprequesthelper.Config) error {
	_logger, err := httprequesthelper.NewLogger(config)
	if err != nil {
		return err
	}

	requestLogger = _logger

	return nil
}

// logGClientRequest 记录gclient请求和响应
// ctx 用于获取用户ID
// respBody 可选，传入nil则不记录响应体
func logGClientRequest(ctx context.Context, method, url string, data interface{}, resp *gclient.Response, respBody []byte, startTime time.Time) {
	if requestLogger == nil || !requestLogger.IsEnabled() {
		return
	}

	duration := time.Since(startTime)

	// 构建请求体字符串
	var reqBodyStr string

	if data != nil {
		bodyBytes, err := cutil.JSON().Marshal(data)
		if err == nil {
			reqBodyStr = string(bodyBytes)
		}
	}

	// 优先使用响应中的原始请求（包含完整的请求头）
	var req *http.Request
	if resp != nil && resp.Request != nil {
		req = resp.Request
	} else {
		// 如果没有原始请求，则构建模拟的http.Request
		req, _ = http.NewRequest(method, url, nil)
	}

	var statusCode int

	var respHeaders http.Header

	if resp != nil {
		statusCode = resp.StatusCode
		respHeaders = resp.Header
	}

	requestLogger.LogRequest(
		ctx,
		req,
		reqBodyStr,
		statusCode,
		respHeaders,
		string(respBody),
		duration,
	)
}
