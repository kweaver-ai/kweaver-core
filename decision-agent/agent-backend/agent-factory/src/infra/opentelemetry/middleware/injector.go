package middleware

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/opentelemetry/logs"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/opentelemetry/metrics"

	"github.com/gin-gonic/gin"
)

// DependencyInjector 依赖注入中间件
type DependencyInjector struct {
	logger  *logs.Logger
	metrics *metrics.Metrics
}

// NewDependencyInjector 创建新的依赖注入中间件
func NewDependencyInjector(logger *logs.Logger, metrics *metrics.Metrics) *DependencyInjector {
	return &DependencyInjector{
		logger:  logger,
		metrics: metrics,
	}
}

// Middleware 返回Gin中间件函数，将logger和metrics注入到context中
func (di *DependencyInjector) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前上下文（应该已经包含OpenTelemetry的span信息）
		ctx := c.Request.Context()

		// 注入logger和metrics到context中
		ctx = logs.WithLogger(ctx, di.logger)
		ctx = metrics.WithMetrics(ctx, di.metrics)

		// 更新请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 继续处理下一个中间件或处理器
		c.Next()
	}
}
