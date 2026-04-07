package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/driveradapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/telemetry"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	logicscommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/mcpinstance"
)

// Server 服务
type Server struct {
	// 健康检查
	httpHealthHandler  interfaces.HTTPRouterInterface
	restPublicHandler  interfaces.HTTPRouterInterface
	restPrivateHandler interfaces.HTTPRouterInterface
	MQHandler          interfaces.MQHandler
	outboxMessageEvent interfaces.App
	config             *config.Config
}

// Start 开启服务
func (s *Server) Start() {
	gin.SetMode(gin.ReleaseMode)
	err := s.outboxMessageEvent.Start()
	if err != nil {
		s.config.Logger.Errorf("start outbox message event failed, error: %v", err)
		panic(err)
	}

	// 注册路由 - 健康检查
	go func() {
		engine := gin.New()
		engine.Use(gin.Recovery())
		engine.UseRawPath = true
		routerHealth := engine.Group("/health")
		s.httpHealthHandler.RegisterRouter(routerHealth)

		// 注册内部接口路由 - 算子相关接口
		routerInternalGroup := engine.Group("/api/agent-operator-integration/internal-v1")
		routerInternalGroup.Use(gin.Recovery())
		s.restPrivateHandler.RegisterRouter(routerInternalGroup)

		// 注册外部路由 - 算子相关接口
		routerGroup := engine.Group("/api/agent-operator-integration/v1")
		routerGroup.Use(gin.Recovery())
		s.restPublicHandler.RegisterRouter(routerGroup)

		url := fmt.Sprintf("%s:%d", s.config.Project.Host, s.config.Project.Port)
		err := engine.Run(url)
		if err != nil {
			s.config.Logger.Errorf("start server failed, error: %v", err)
		}
	}()
	// 启动MQ处理
	go s.MQHandler.Subscribe()
}

// Stop 停止服务
func (s *Server) Stop(ctx context.Context) {
	s.config.Logger.Info("stop agent-operator-integration server")
	// sandbox.Close()      // 关闭并销毁沙箱会话池
	s.outboxMessageEvent.Stop(ctx)
	mcpinstance.Close() // 关闭实例池
}

func main() {
	// 初始化全局配置
	config := config.NewConfigLoader()
	// 设置错误码语言
	common.SetLang(config.Project.Language)
	s := &Server{
		config:             config,
		httpHealthHandler:  driveradapters.NewHTTPHealthHandler(),
		restPublicHandler:  driveradapters.NewRestPublicHandler(),
		restPrivateHandler: driveradapters.NewRestPrivateHandler(),
		outboxMessageEvent: logicscommon.NewOutboxMessageEvent(),
		MQHandler:          driveradapters.NewMQHandler(),
	}
	s.config.Logger.Info("start agent-operator-integration server")
	// 检查是否开启了可观测性
	if config.Observability.TraceEnabled {
		defer telemetry.StopTrace()
	}
	s.Start()
	defer s.Stop(context.Background())
	// 等待信号量
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) //nolint
	<-c
}
