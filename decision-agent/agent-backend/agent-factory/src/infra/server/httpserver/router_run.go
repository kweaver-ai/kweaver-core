package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// registerRunRoutes 注册Run侧路由 (V1)
// 根据配置决定是否保留老的APP路径
func (s *httpServer) registerRunRoutes(engine *gin.Engine) {
	// 始终注册新路径
	s.runPubRouter(engine, "/api/agent-factory/v1")
	s.runPriRouter(engine, "/api/agent-factory/internal/v1")

	// 根据配置决定是否保留老路径
	if global.GConfig.SwitchFields.KeepLegacyAppPath {
		s.runPubRouter(engine, "/api/agent-app/v1")
		s.runPriRouter(engine, "/api/agent-app/internal/v1")
	}
}

// runPubRouter 注册Run侧公开路由
func (s *httpServer) runPubRouter(engine *gin.Engine, basePath string) {
	router := engine.Group(basePath)

	if cenvhelper.IsLocalDev() {
		router.Use(capimiddleware.Cors())

		// 添加通用OPTIONS路由处理CORS预检请求
		router.OPTIONS("/*path", func(c *gin.Context) {})
	}

	router.Use(
		capimiddleware.Recovery(),
		capimiddleware.RequestLoggerV2Middleware(),
		// 获取访问语言
		capimiddleware.Language(),
		// 新增 Hydra 接口鉴权，开发环境可以临时屏蔽
		capimiddleware.VerifyOAuthMiddleWare(),
		// 业务域：外部接口要求必须携带业务域ID
		capimiddleware.HandleBizDomain(false),

		// 注入OpenTelemetry中间件
		otelgin.Middleware(global.GConfig.OtelConfig.ServiceName),
		// 注入logs和metrics
		global.GDependencyInjector.Middleware(),
	)

	s.agentHandler.RegPubRouter(router)
	s.conversationHandler.RegPubRouter(router)
	s.observabilityHandler.RegPubRouter(router)
	s.sessionHandler.RegPubRouter(router)
}

// runPriRouter 注册Run侧私有路由
func (s *httpServer) runPriRouter(engine *gin.Engine, basePath string) {
	internalRouterG := engine.Group(basePath)

	internalRouterG.Use(
		capimiddleware.Recovery(),
		capimiddleware.RequestLoggerV2Middleware(),
		capimiddleware.Language(),
		// 业务域：内部接口自动使用默认业务域
		capimiddleware.HandleBizDomain(true),

		// 注入OpenTelemetry中间件
		otelgin.Middleware(global.GConfig.OtelConfig.ServiceName),
		// 注入logs和metrics
		global.GDependencyInjector.Middleware(),
	)

	s.agentHandler.RegPriRouter(internalRouterG)
	s.conversationHandler.RegPriRouter(internalRouterG)
	s.observabilityHandler.RegPriRouter(internalRouterG)
}
