package capimiddleware

import (
	"github.com/gin-gonic/gin"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/ginrequestlogger"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httprequesthelper"
)

// InitDefaultRequestLoggerV2 初始化默认的请求日志记录器V2（单例）
// 如果已经初始化过，则忽略本次调用
func InitDefaultRequestLoggerV2(config *httprequesthelper.Config) error {
	return ginrequestlogger.InitDefaultRequestLogger(config)
}

// RequestLoggerV2Middleware 返回默认请求日志记录器V2的中间件
// 如果未初始化，则panic
func RequestLoggerV2Middleware() gin.HandlerFunc {
	return ginrequestlogger.DefaultMiddleware()
}
