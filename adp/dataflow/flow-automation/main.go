// Package main 模块
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/admin"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/alarm"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/anydata"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/auth"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/callback"
	cognitiveassistant "github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/cognitive_assistant"
	cconfig "github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/config"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/database_con"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/dataflow"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/dataflow_doc"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/executor"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/health"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/master"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/mgnt"
	modellib "github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/model_lib"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/observability"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/operators"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/policy"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/sandbox"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/security_policy"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/trigger"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/driveradapters/versions"
	wlHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	threadPool "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/pools"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/module/initial"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/download_pool"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/analysis"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/management"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var analysisServer analysis.AnalysisService
var managementServer management.ManagementService
var tracerProvider *sdktrace.TracerProvider

type app struct {
	hRESTHandler        health.RESTHandler
	mRESTHandler        mgnt.RESTHandler
	aRESTHandler        auth.RESTHandler
	cbRESTHandler       callback.RESTHandler
	pRESTHandler        policy.RESTHandler
	tRESTHandler        trigger.RESTHandler
	tMQHandler          trigger.MQHandler
	spRESTHandler       security_policy.RESTHandler
	mLRESTHandler       modellib.RESTHandler
	tMaster             master.Master
	cRESTHandler        cognitiveassistant.RESTHandler
	executorRESTHandler executor.RESTHandler
	adminRESTHandler    admin.RESTHandler
	cfRESTHandler       cconfig.RESTHandler
	adRESTHandler       anydata.RESTHandler
	alarmRESTHandler    alarm.RESTHandler
	dfRESTHandler       dataflow.RESTHandler
	dfDocRESTHandler    dataflow_doc.RESTHandler
	coRESTHandler       operators.RESTHandler
	obsRESTHandler      observability.RESTHandler
	dvRESTHandler       versions.RESTHandler
	dbRESTHandler       database_con.RESTHandler
	sandboxRESTHandler  sandbox.RESTHandler
	downloadPool        download_pool.Pool
}

func CacheControl() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "no-cache")
	}
}

func (a *app) Start() {
	log := commonLog.NewLogger()
	log.Infoln("start server.....")

	debugMode := os.Getenv("DEBUG")
	if debugMode == "false" {
		gin.SetMode(gin.ReleaseMode)
	}

	go func() {
		engine := gin.New()
		engine.Use(gin.Recovery())
		engine.Use(CacheControl())
		engine.UseRawPath = true
		port := os.Getenv("API_SERVER_PORT")
		prefix := os.Getenv("API_PREFIX")

		group := engine.Group(prefix)
		group.Use(wlHttp.MiddlewareTrace(), wlHttp.LanguageMiddleware())

		// 注册API
		a.hRESTHandler.RegisterAPI(group)
		a.mRESTHandler.RegisterAPI(group)
		a.aRESTHandler.RegisterAPI(group)
		a.pRESTHandler.RegisterAPI(group)
		a.mLRESTHandler.RegisterAPI(group)
		a.executorRESTHandler.RegisterAPI(group)
		a.adminRESTHandler.RegisterAPI(group)
		a.cfRESTHandler.RegisterAPI(group)
		a.alarmRESTHandler.RegisterAPI(group)
		a.tRESTHandler.RegisterAPI(group)
		a.adRESTHandler.RegisterAPI(group)
		a.coRESTHandler.RegisterAPI(group)
		a.obsRESTHandler.RegisterAPI(group)
		a.dvRESTHandler.RegisterAPI(group)
		a.dbRESTHandler.RegisterAPI(group)
		a.sandboxRESTHandler.RegisterAPI(group)
		spGroup := group.Group("security-policy")
		a.spRESTHandler.RegisterAPI(spGroup)

		cGroup := group.Group("cognitive-assistant")
		a.cRESTHandler.RegisterAPI(cGroup)

		dfGroup := group.Group("data-flow")
		a.dfRESTHandler.RegisterAPI(dfGroup)

		groupV2 := engine.Group(common.APIPREFIXV2)
		groupV2.Use(wlHttp.MiddlewareTrace())
		a.mRESTHandler.RegisterAPIv2(groupV2)
		a.dfDocRESTHandler.RegisterAPIv2(groupV2)

		if err := engine.Run(":" + port); err != nil {
			log.Errorln(err)
		}
	}()

	go func() {
		engine := gin.New()
		engine.Use(gin.Recovery())
		engine.Use(CacheControl())
		engine.UseRawPath = true
		port := os.Getenv("API_SERVER_PRIVATE_PORT")
		prefix := os.Getenv("API_PREFIX")

		group := engine.Group(prefix)

		// 注册API
		a.cbRESTHandler.RegisterPrivateAPI(group)
		a.tRESTHandler.RegisterPrivateAPI(group)
		a.cfRESTHandler.RegisterPrivateAPI(group)
		a.mRESTHandler.RegisterPrivateAPI(group)
		a.coRESTHandler.RegisterPrivateAPI(group)

		spGroup := group.Group("security-policy")
		a.spRESTHandler.RegisterPrivateAPI(spGroup)

		if err := engine.Run(":" + port); err != nil {
			log.Errorln(err)
		}
	}()

	// 订阅nsq 处理订阅消息
	go func() {
		a.tMQHandler.Subscribe()
	}()

	go func() {
		a.tMaster.Run()
	}()

	// 启动文件下载线程池
	if a.downloadPool != nil {
		go func() {
			a.downloadPool.Start(context.Background())
		}()
	}
}

func StartDataFlow() {
	// 加载环境变量文件
	// 先加载 .env，再加载 .env.local（如果存在）来覆盖
	if err := godotenv.Load(".env"); err != nil {
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}
	// .env.local 用于本地开发，会覆盖 .env 中的配置
	if err := godotenv.Overload(".env.local"); err != nil {
		// .env.local 是可选的，不存在时不报错
		fmt.Printf("Info: .env.local file not found (optional): %v\n", err)
	}

	config, err := common.InitConfig()
	if err != nil {
		panic(err.Error())
	}

	if err := initial.Init(&initial.InitialOption{
		ParserWorkersCnt:         config.Server.ParserrCount,
		LowestExecutorWorkerCnt:  config.Server.LowestExecutorCount,
		LowExecutorWorkerCnt:     config.Server.LowExecutorCount,
		MediumExecutorWorkerCnt:  config.Server.MediumExecutorCount,
		HighExecutorWorkerCnt:    config.Server.HighExecutorCount,
		HighestExecutorWorkerCnt: config.Server.HighestExecutorCount,
		ListInsCount:             config.Server.ListInsCount,
		ExecutorTimeout:          time.Duration(config.Server.ExecutorTimeout) * time.Second,
		DagScheduleTimeout:       time.Duration(config.Server.ScheduleTimeout) * time.Second,
	}); err != nil {
		panic(err.Error())
	}

	tracerProvider = trace.SetTraceExporter(&config.Telemetry)
	server := &app{
		hRESTHandler:        health.NewRESTHandler(),
		mRESTHandler:        mgnt.NewRESTHandler(),
		aRESTHandler:        auth.NewRESTHandler(),
		cbRESTHandler:       callback.NewRESTHandler(),
		pRESTHandler:        policy.NewRESTHandler(),
		tRESTHandler:        trigger.NewRESTHandler(),
		tMQHandler:          trigger.NewMQHandler(),
		spRESTHandler:       security_policy.NewRESTHandler(),
		mLRESTHandler:       modellib.NewRESTHandler(),
		tMaster:             master.NewOnMaster(),
		cRESTHandler:        cognitiveassistant.NewRESTHandler(),
		executorRESTHandler: executor.NewRESTHandler(),
		adminRESTHandler:    admin.NewRESTHandler(),
		cfRESTHandler:       cconfig.NewRESTHandler(),
		alarmRESTHandler:    alarm.NewRESTHandler(),
		adRESTHandler:       anydata.NewRestHandler(),
		dfRESTHandler:       dataflow.NewRESTHandler(),
		dfDocRESTHandler:    dataflow_doc.NewRESTHandler(),
		coRESTHandler:       operators.NewRESTHandler(),
		obsRESTHandler:      observability.NewRESTHandler(),
		dvRESTHandler:       versions.NewRESTHandler(),
		dbRESTHandler:       database_con.NewRestHandler(),
		sandboxRESTHandler:  sandbox.NewRESTHandler(),
		downloadPool:        download_pool.NewPool(),
	}
	server.Start()
}

func StartEcronAnalysis() {
	analysisServer = analysis.NewAnalysisService()
	analysisServer.Start()
}

func StartEcronManagement() {
	managementServer = management.NewManagementService()
	managementServer.Start()
}

func Release() {
	traceLog.Close()
	traceLog.CloseFlowO11yLogger()
	initial.Close()
	trace.ExitTraceExporter(context.Background(), tracerProvider)
	threadPool.Pools.ShutdownAll()

	if analysisServer != nil {
		analysisServer.Stop()
	}

	if managementServer != nil {
		managementServer.Stop()
	}
}

func main() {
	go StartEcronManagement()
	go StartEcronAnalysis()
	go StartDataFlow()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) //nolint
	<-c

	// 服务后置操作
	Release()
}
