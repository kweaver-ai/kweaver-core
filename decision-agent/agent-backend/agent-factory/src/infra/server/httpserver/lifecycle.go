package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
)

// Start 启动HTTP服务器
func (s *httpServer) Start() {
	go func() {
		gin.SetMode(gin.ReleaseMode)

		if cenvhelper.IsLocalDev() {
			gin.SetMode(gin.DebugMode)
		}

		engine := gin.New()

		// 开启 ContextWithFallback
		engine.ContextWithFallback = true

		engine.Use(gin.Logger())

		// 注册路由 - 健康检查
		s.registerHealthRoutes(engine)

		// 注册路由 - Swagger UI
		s.registerSwaggerRoutes(engine)

		// 注册路由 - Management侧 (V3)
		s.registerManagementPubRoutes(engine)
		s.registerManagementPriRoutes(engine)

		// 注册路由 - Run侧 (V1)
		s.registerRunRoutes(engine)

		url := fmt.Sprintf("%s:%d", global.GConfig.Project.Host, global.GConfig.Project.Port)

		// 创建 HTTP 服务器
		s.httpSrv = &http.Server{
			Addr:    url,
			Handler: engine,
		}

		// 启动服务器
		err := s.httpSrv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			err = fmt.Errorf("http server start failed, err: %w", err)
			panic(err)
		}
	}()
}

// Shutdown 优雅关闭服务器
func (s *httpServer) Shutdown(ctx context.Context) error {
	if s.httpSrv == nil {
		return nil
	}

	// 直接使用传入的上下文，由调用方控制超时
	return s.httpSrv.Shutdown(ctx)
}
