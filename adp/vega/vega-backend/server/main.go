// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package main

import (
	"context"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	_ "unicode/utf8"

	"github.com/gin-gonic/gin"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	_ "go.uber.org/automaxprocs"

	"vega-backend/common"
	"vega-backend/drivenadapters/auth"
	"vega-backend/drivenadapters/permission"
	"vega-backend/drivenadapters/user_mgmt"
	"vega-backend/driveradapters"
	"vega-backend/logics"
	"vega-backend/logics/connectors/factory"
	"vega-backend/logics/discover_schedule"
	"vega-backend/logics/discover_task"
	"vega-backend/worker"
)

type mgrService struct {
	appSetting    *common.AppSetting
	otelProviders *otel.Providers
	restHandler   driveradapters.RestHandler
}

func (server *mgrService) start() {
	logger.Info("Server Starting")

	// 创建 gin.engine 并注册 API
	engine := gin.New()

	server.restHandler.RegisterPublic(engine)
	logger.Info("Server Register API Success")

	// 监听中断信号（SIGINT、SIGTERM）
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// 在收到信号的时候，会自动触发 ctx 的 Done ，这个 stop 是不再捕获注册的信号的意思，算是一种释放资源。
	defer stop()

	// 初始化 http 服务
	s := &http.Server{
		Addr:           ":" + strconv.Itoa(server.appSetting.ServerSetting.HttpPort),
		Handler:        engine,
		ReadTimeout:    server.appSetting.ServerSetting.ReadTimeOut * time.Second,
		WriteTimeout:   server.appSetting.ServerSetting.WriteTimeout * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// 启动 http 服务
	go func() {
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Fatalf("s.ListenAndServe err:%v", err)
		}
	}()

	logger.Infof("Server Started on Port:%d", server.appSetting.ServerSetting.HttpPort)

	<-ctx.Done()

	// 设置系统最后处理时间
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 停止 http 服务
	logger.Info("Server Start Shutdown")
	if err := s.Shutdown(ctx); err != nil {
		logger.Fatalf("Server Shutdown:%v", err)
	}

	server.otelProviders.Shutdown(ctx)

	logger.Info("Server Exited")
}

func main() {
	logger.Info("Server Initializing")

	// 初始化服务配置
	appSetting := common.NewSetting()
	logger.Info("Server Init Setting Success")

	// 设置错误码语言
	rest.SetLang(appSetting.ServerSetting.Language)
	logger.Info("Server Set Language Success")

	// 设置 gin 运行模式
	gin.SetMode(appSetting.ServerSetting.RunMode)
	logger.Infof("Server RunMode: %s", appSetting.ServerSetting.RunMode)

	logger.Infof("Server Start By Port:%d", appSetting.ServerSetting.HttpPort)

	otelProviders, err := otel.InitOTel(context.Background(), &appSetting.OtelSetting)
	if err != nil {
		logger.Fatalf("Failed to initialize OpenTelemetry provider: %v", err)
	}

	// 初始化数据库连接
	db := libdb.NewDB(&appSetting.DBSetting)
	logics.SetDB(db)

	// Set顺序按字母升序排序
	if common.GetAuthEnabled() {
		logics.SetAuthAccess(auth.NewHydraAuthAccess(appSetting))
		logics.SetPermissionAccess(permission.NewPermissionAccess(appSetting))
		logics.SetUserMgmtAccess(user_mgmt.NewUserMgmtAccess(appSetting))
	}

	// 初始化 Connector Factory 并注册内置的 Local Connector Builder
	factory.Init(appSetting)
	logger.Info("VEGA Manager Init Connector Factory Success")

	// 初始化并启动统一的 TaskWorkerManger，处理所有类型的任务
	taskWorkerMgr := worker.NewTaskWorkerManager(appSetting)
	taskWorkerMgr.Start()
	logger.Info("VEGA Manager Init Task Worker Success")

	// 初始化并启动调度器
	dts := discover_task.NewDiscoverTaskService(appSetting)
	dss := discover_schedule.NewDiscoverScheduleService(appSetting, dts)
	sw := worker.NewScheduleWorker(appSetting, dss)
	if err := sw.Start(); err != nil {
		logger.Fatalf("Failed to start scheduler: %v", err)
	}
	logger.Info("VEGA Manager Init Scheduler Success")
	defer sw.Stop()

	// 创建并启动服务
	server := &mgrService{
		appSetting:    appSetting,
		otelProviders: otelProviders,
		restHandler:   driveradapters.NewRestHandler(appSetting, sw),
	}
	server.start()
}
