// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package main

import (
	"context"
	"net/http"

	// _ "net/http/pprof"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	_ "unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	libdb "github.com/kweaver-ai/kweaver-go-lib/db"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	_ "go.uber.org/automaxprocs"

	"bkn-backend/common"
	"bkn-backend/drivenadapters/action_schedule"
	"bkn-backend/drivenadapters/action_type"
	"bkn-backend/drivenadapters/auth"
	"bkn-backend/drivenadapters/business_system"
	"bkn-backend/drivenadapters/concept_group"
	"bkn-backend/drivenadapters/data_model"
	"bkn-backend/drivenadapters/data_view"
	"bkn-backend/drivenadapters/job"
	"bkn-backend/drivenadapters/knowledge_network"
	"bkn-backend/drivenadapters/model_factory"
	"bkn-backend/drivenadapters/object_type"
	"bkn-backend/drivenadapters/opensearch"
	"bkn-backend/drivenadapters/permission"
	"bkn-backend/drivenadapters/relation_type"
	"bkn-backend/drivenadapters/risk_type"
	"bkn-backend/drivenadapters/user_mgmt"
	"bkn-backend/drivenadapters/vega_backend"
	"bkn-backend/driveradapters"
	"bkn-backend/interfaces"
	"bkn-backend/logics"
	"bkn-backend/worker"
)

type mgrService struct {
	appSetting     *common.AppSetting
	restHandler    driveradapters.RestHandler
	conceptSyncer  *worker.ConceptSyncer
	jobExecutor    interfaces.JobExecutor
	scheduleWorker *worker.ScheduleWorker
}

func (server *mgrService) start() {
	logger.Info("Server Starting")

	err := logics.Init(context.Background(), server.appSetting)
	if err != nil {
		panic(err)
	}

	// 创建gin.engine 并注册Public API
	engine := gin.New()

	server.restHandler.RegisterPublic(engine)
	logger.Info("Server Register API Success")

	go server.conceptSyncer.Start()
	go server.jobExecutor.Start()
	go server.scheduleWorker.Start()

	// 监听中断信号（SIGINT、SIGTERM）
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// 在收到信号的时候，会自动触发 ctx 的 Done ，这个 stop 是不再捕获注册的信号的意思，算是一种释放资源。
	defer stop()

	// 初始化http服务
	s := &http.Server{
		Addr:           ":" + strconv.Itoa(server.appSetting.ServerSetting.HttpPort),
		Handler:        engine,
		ReadTimeout:    server.appSetting.ServerSetting.ReadTimeOut * time.Second,
		WriteTimeout:   server.appSetting.ServerSetting.WriteTimeout * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// 启动http服务
	go func() {
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Fatalf("s.ListenAndServe err:%v", err)
		}
	}()

	<-ctx.Done()

	// 设置系统最后处理时间
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 停止http服务
	logger.Info("Server Start Shutdown")
	if err := s.Shutdown(ctx); err != nil {
		logger.Fatalf("Server Shutdown:%v", err)
	}
	logger.Info("Server Exiting")
}

func main() {
	// 开启 pprof
	// go func() {
	// 	http.ListenAndServe("0.0.0.0:6060", nil)
	// }()

	logger.Info("Server Starting")

	// 初始化服务配置
	appSetting := common.NewSetting()

	logger.Info("Server Init Setting Success")

	// 设置错误码语言
	rest.SetLang(appSetting.ServerSetting.Language)
	logger.Info("Server Set Language Success")

	// 设置gin运行模式
	gin.SetMode(appSetting.ServerSetting.RunMode)
	logger.Info("Server Set RunMode Success")

	logger.Infof("Server Start By Port:%d", appSetting.ServerSetting.HttpPort)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	db := libdb.NewDB(&appSetting.DBSetting)
	logics.SetDB(db)

	audit.Init(&appSetting.MQSetting)

	// Set顺序按字母升序排序
	if common.GetAuthEnabled() {
		logics.SetAuthAccess(auth.NewHydraAuthAccess(appSetting))
		logics.SetPermissionAccess(permission.NewPermissionAccess(appSetting))
		logics.SetUserMgmtAccess(user_mgmt.NewUserMgmtAccess(appSetting))
	}
	logics.SetActionScheduleAccess(action_schedule.NewActionScheduleAccess(appSetting))
	logics.SetActionTypeAccess(action_type.NewActionTypeAccess(appSetting))
	if common.GetBusinessDomainEnabled() {
		logics.SetBusinessSystemAccess(business_system.NewBusinessSystemAccess(appSetting))
	}
	logics.SetConceptGroupAccess(concept_group.NewConceptGroupAccess(appSetting))
	logics.SetDataModelAccess(data_model.NewDataModelAccess(appSetting))
	logics.SetDataViewAccess(data_view.NewDataViewAccess(appSetting))
	logics.SetJobAccess(job.NewJobAccess(appSetting))
	logics.SetKNAccess(knowledge_network.NewKNAccess(appSetting))
	logics.SetModelFactoryAccess(model_factory.NewModelFactoryAccess(appSetting))
	logics.SetOpenSearchAccess(opensearch.NewOpenSearchAccess(appSetting))
	logics.SetObjectTypeAccess(object_type.NewObjectTypeAccess(appSetting))
	logics.SetRelationTypeAccess(relation_type.NewRelationTypeAccess(appSetting))
	logics.SetRiskTypeAccess(risk_type.NewRiskTypeAccess(appSetting))
	logics.SetVegaBackendAccess(vega_backend.NewVegaBackendAccess(appSetting))

	server := &mgrService{
		appSetting:     appSetting,
		restHandler:    driveradapters.NewRestHandler(appSetting),
		conceptSyncer:  worker.NewConceptSyncer(appSetting),
		jobExecutor:    worker.NewJobExecutor(appSetting),
		scheduleWorker: worker.NewScheduleWorker(appSetting),
	}
	server.start()
}
