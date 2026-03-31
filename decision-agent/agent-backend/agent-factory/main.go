package main

// @title           Agent Factory API
// @version         1.0
// @description     Agent Factory 智能体工厂 API 文档
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath  /api/agent-factory
// @host      localhost:30777

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer {token} - 需要有效的 OAuth 访问令牌

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/kweaver-ai/decision-agent/agent-factory/src/boot"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/opentelemetry"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/server/httpserver"
)

func main() {
	// 初始化OpenTelemetry provider
	otelProvider, err := opentelemetry.NewProvider(global.GConfig.OtelConfig)
	if err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry provider: %v", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := otelProvider.Shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown OpenTelemetry provider: %v", err)
		}
	}()

	// 初始化OpenTelemetry相关全局变量
	if err := global.InitOpenTelemetry(otelProvider); err != nil {
		log.Fatalf("Failed to initialize OpenTelemetry globals: %v", err)
	}

	s := httpserver.NewHTTPServer()
	s.Start()

	// 创建一个通道来接收操作系统信号
	quit := make(chan os.Signal, 1)
	// 注册通道接收特定的信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞，直到接收到信号
	<-quit
	log.Println("正在关闭服务器...")

	// 创建一个超时上下文，给服务器10秒时间优雅关闭
	timeout := 10 * time.Second
	if cenvhelper.IsLocalDev() {
		timeout = 1 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 尝试优雅关闭服务器
	if err := s.Shutdown(ctx); err != nil {
		log.Printf("服务器强制关闭: %v", err)
		os.Exit(1)
	}

	log.Println("服务器已退出")
}
