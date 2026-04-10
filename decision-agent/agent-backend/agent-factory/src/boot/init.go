package boot

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common"
	_ "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/redishelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/mq"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/pkg/errors"
)

// 初始化
func init() {
	// 1. 初始化配置
	global.GConfig = conf.NewConfig()
	cglobal.GConfig = global.GConfig.Config

	// 设置默认语言
	rest.SetLang(cglobal.GConfig.GetDefaultLanguage())

	// 2. 初始化数据库
	global.GDB = common.NewDBPool() // 初始化全局DB
	cglobal.GDB = global.GDB

	// 3. 初始化redis
	redishelper.ConnectRedis(&global.GConfig.Redis)

	// 4. 初始化日志
	logFile := global.GConfig.Project.LogFile
	if logFile == "" {
		logFile = "/app/agent-factory/logs/agent-factory.log"
	}

	// new 2025年04月16日14:42:00
	// logger 初始化
	lggerSetting := logger.LogSetting{
		LogServiceName: "agent-factory",
		LogLevel:       global.GConfig.GetLogLevelString(),
		LogFileName:    logFile,
		MaxAge:         30,
		MaxBackups:     10,
		MaxSize:        100,
	}
	logger.InitGlobalLogger(lggerSetting)

	// 5. 设置http request log
	// 5.1 设置http client request log (记录发起的请求)
	initHTTPClientRequestLog()

	// 5.2 设置http server request log (记录接收到的请求)
	initHTTPServerRequestLog()

	// 6. 初始化权限
	err := initPermission()
	if err != nil {
		err = errors.Wrap(err, "init permission failed")
		logger.GetLogger().Panic(err)

		return
	}

	// 7. 初始化业务域关联关系
	err = initBizDomainRel()
	if err != nil {
		err = errors.Wrap(err, "init biz domain rel failed")
		logger.GetLogger().Panic(err)

		return
	}

	// 8. 初始化审计日志
	if !global.GConfig.SwitchFields.DisableAuditInit {
		mqSetting := &mq.MQSetting{
			MQType: global.GConfig.MQ.MQType,
			MQHost: global.GConfig.MQ.MQHost,
			MQPort: global.GConfig.MQ.MQPort,
			Tenant: global.GConfig.MQ.Tenant,
			Auth: mq.MQAuthSetting{
				Mechanism: global.GConfig.MQ.Auth.Mechanism,
				Username:  global.GConfig.MQ.Auth.Username,
				Password:  global.GConfig.MQ.Auth.Password,
			},
		}
		audit.Init(mqSetting)
	}

	// 9. 初始化 OTel (新版可观测性)
	if global.GConfig.OtelV2Config != nil &&
		(global.GConfig.OtelV2Config.Trace.Enabled || global.GConfig.OtelV2Config.Log.Enabled) {
		_, err := otel.InitOTel(context.Background(), global.GConfig.OtelV2Config)
		if err != nil {
			logger.GetLogger().Errorf("init otel failed: %v", err)
		}
	}
}
