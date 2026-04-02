package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/apimiddleware"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// registerManagementPubRoutes 注册Management侧公开路由 (V3)
func (s *httpServer) registerManagementPubRoutes(engine *gin.Engine) {
	router := engine.Group("/api/agent-factory/v3")

	if cenvhelper.IsLocalDev() {
		router.Use(capimiddleware.Cors())

		// 添加通用OPTIONS路由处理CORS预检请求
		router.OPTIONS("/*path", func(c *gin.Context) {})
	}

	router.Use(
		capimiddleware.Recovery(),
		capimiddleware.RequestLoggerV2Middleware(),
		capimiddleware.ErrorHandler(),
		// 获取访问语言
		capimiddleware.Language(),
		// 新增 Hydra 接口鉴权，开发环境可以临时屏蔽
		capimiddleware.VerifyOAuthMiddleWare(),
		// 业务域：外部接口要求必须携带业务域ID
		capimiddleware.HandleBizDomain(false),
		apimiddleware.VisitorTypeCheck(),

		// 注入OpenTelemetry中间件
		otelgin.Middleware(global.GConfig.OtelV2Config.ServiceName),
	)

	s.v3AgentConfigHandler.RegPubRouter(router)
	s.v3AgentTplHandler.RegPubRouter(router)
	s.productHandler.RegPubRouter(router)
	s.categoryHandler.RegPubRouter(router)
	s.releaseHandler.RegPubRouter(router)
	s.squareHandler.RegPubRouter(router)
	s.permissionHandler.RegPubRouter(router)
	s.publishedHandler.RegPubRouter(router)

	s.personalSpaceHandler.RegPubRouter(router)
	s.otherHandler.RegPubRouter(router)
	s.testHandler.RegPubRouter(router)
	s.anysharedsHandler.RegPubRouter(router)
}

// registerManagementPriRoutes 注册Management侧私有路由 (V3)
func (s *httpServer) registerManagementPriRoutes(engine *gin.Engine) {
	internalRouterG := engine.Group("/api/agent-factory/internal/v3")

	internalRouterG.Use(
		capimiddleware.Recovery(),
		capimiddleware.ErrorHandler(),
		capimiddleware.RequestLoggerV2Middleware(),
		capimiddleware.Language(),
		// 业务域：内部接口自动使用默认业务域
		capimiddleware.HandleBizDomain(true),

		// 注入OpenTelemetry中间件
		otelgin.Middleware(global.GConfig.OtelV2Config.ServiceName),
	)

	s.releaseHandler.RegPriRouter(internalRouterG)
	s.v3AgentConfigHandler.RegPriRouter(internalRouterG)
	s.v3AgentTplHandler.RegPriRouter(internalRouterG)
	s.squareHandler.RegPriRouter(internalRouterG)
	s.publishedHandler.RegPriRouter(internalRouterG)
	s.permissionHandler.RegPriRouter(internalRouterG)
	s.otherHandler.RegPriRouter(internalRouterG)
	s.testHandler.RegPriRouter(internalRouterG)
}
