package httpserver

import "github.com/gin-gonic/gin"

// registerHealthRoutes 注册健康检查路由
func (s *httpServer) registerHealthRoutes(engine *gin.Engine) {
	routerHealth := engine.Group("/health")
	s.httpHealthHandler.RegHealthRouter(routerHealth)
}
