package utils

import (
	"io/ioutil"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"
	"gopkg.in/yaml.v2"
)

//go:generate mockgen -package mock -source ../utils/config.go -destination ../mock/mock_config.go

var (
	MQConfigPath      = "/sysvol/conf/service_conf/mq_config.yaml"
	CronsvrPath       = "/sysvol/conf/service_conf/cronsvr.yaml"
	CronsvrSecretPath = "/sysvol/conf/secret_conf/cronsvr-secret.yaml"
)

// Configer 配置服务
type Configer interface {
	Config() common.ConfigLoader
	Lang() common.LangLoader
}

// NewConfiger 加载服务
func NewConfiger() Configer {
	return newResourceLoader()
}

type resourceloader struct {
	common.ConfigLoader
	common.LangLoader
}

func (l resourceloader) Lang() common.LangLoader {
	return l.LangLoader
}

func (l resourceloader) Config() common.ConfigLoader {
	return l.ConfigLoader
}

var (
	resourceHandle *resourceloader = nil
	resourceMutex  sync.Mutex

	configLog = NewLogger()
)

func newResourceLoader() *resourceloader {
	resourceMutex.Lock()
	defer resourceMutex.Unlock()

	if nil != resourceHandle {
		return resourceHandle
	}

	cLoader := common.ConfigLoader{}
	getConf(&cLoader, CronsvrPath)
	getConf(&cLoader, CronsvrSecretPath)
	// 获取 MQConnectorType，不同连接方式subsribe不同的channel
	getConf(&cLoader, MQConfigPath)
	cLoader.MQConfigPath = MQConfigPath
	configLog.Infof("Cron Addr:%v", cLoader.CronPort)
	l := newLang(cLoader)
	lLoader := common.LangLoader{Lang: l}
	resourceHandle = &resourceloader{ConfigLoader: cLoader, LangLoader: lLoader}
	return resourceHandle
}

func newLang(c common.ConfigLoader) map[string]string {
	l := make(map[string]string)

	switch c.Lang {
	case "":
		fallthrough
	case "zh_CN":
		l["IDS_BAD_REQUEST"] = "请求错误。"
		l["IDS_UNAUTHORIZED"] = "授权无效。"
		l["IDS_NOT_FOUND"] = "数据不存在。"
		l["IDS_TOO_MANY_REQUESTS"] = "请求太多。"
		l["IDS_INTERNAL_ERROR"] = "内部错误。"

		l[common.ErrDataBaseUnavailable] = "数据库服务不可用"
		l[common.ErrHTTPClientUnavailable] = "HTTP客户端服务不可用"
		l[common.ErrMSMQClientUnavailable] = "NSQ消息队列服务不可用"
		l[common.ErrCronClientUnavailable] = "定时服务不可用"
		l[common.ErrUnsupportedExecutionMode] = "任务执行方式不支持"

		l[common.ErrExecuteIDAndJobIDConfused] = "执行流水号和任务ID混乱"
		l[common.ErrJobExecutedTooManyTimes] = "任务操过了执行次数"
		l[common.ErrExecutorUnavailable] = "执行服务不可用"
		l[common.ErrGinContextUnavailable] = "HTTP服务上下文无效"
		l[common.ErrDataBaseDisconnected] = "数据库连接失败"

		l[common.ErrJobNameExists] = "任务名称已存在"
		l[common.ErrJobNameEmpty] = "任务名称为空"
		l[common.ErrDataBaseExecError] = "数据库执行错误"
		l[common.InfoDataBaseConnected] = "数据库连接成功"

	case "zh_TW":
		l["IDS_BAD_REQUEST"] = "请求错误。"
		l["IDS_UNAUTHORIZED"] = "授权无效。"
		l["IDS_NOT_FOUND"] = "数据不存在。"
		l["IDS_TOO_MANY_REQUESTS"] = "请求太多。"
		l["IDS_INTERNAL_ERROR"] = "內部錯誤。"

		l[common.ErrDataBaseUnavailable] = "資料服務不可用"
		l[common.ErrHTTPClientUnavailable] = "HTTP用戶端服務不可用"
		l[common.ErrMSMQClientUnavailable] = "NSQ訊息佇列服務不可用"
		l[common.ErrCronClientUnavailable] = "定時服務不可用"
		l[common.ErrUnsupportedExecutionMode] = "任務執行方式不支援"

		l[common.ErrExecuteIDAndJobIDConfused] = "執行流水號和任務ID混亂"
		l[common.ErrJobExecutedTooManyTimes] = "任務操過了執行次數"
		l[common.ErrExecutorUnavailable] = "執行服務不可用"
		l[common.ErrGinContextUnavailable] = "HTTP服務上下文無效"
		l[common.ErrDataBaseDisconnected] = "資料庫連接失敗"

		l[common.ErrJobNameExists] = "任務名稱已存在"
		l[common.ErrJobNameEmpty] = "任務名稱為空"
		l[common.ErrDataBaseExecError] = "資料庫執行錯誤"
		l[common.InfoDataBaseConnected] = "資料庫連接成功"

	case "en_US":
		l["IDS_BAD_REQUEST"] = "Bad request."
		l["IDS_UNAUTHORIZED"] = "Unauthorized."
		l["IDS_NOT_FOUND"] = "Not found."
		l["IDS_TOO_MANY_REQUESTS"] = "Too many requests."
		l["IDS_INTERNAL_ERROR"] = "Internal error."

		l[common.ErrDataBaseUnavailable] = "Database is unavailable"
		l[common.ErrHTTPClientUnavailable] = "HTTPP client is unavailable"
		l[common.ErrMSMQClientUnavailable] = "Msmq client is unavailable"
		l[common.ErrCronClientUnavailable] = "Cron client is unavailable"
		l[common.ErrUnsupportedExecutionMode] = "Unsupported execution mode"

		l[common.ErrExecuteIDAndJobIDConfused] = "Executed ID and job ID are in chaos"
		l[common.ErrJobExecutedTooManyTimes] = "Too many job executions"
		l[common.ErrExecutorUnavailable] = "Executor is unavailable"
		l[common.ErrGinContextUnavailable] = "Gin context is unavailable"
		l[common.ErrDataBaseDisconnected] = "Database is disconnected"

		l[common.ErrJobNameExists] = "Job name already exists"
		l[common.ErrJobNameEmpty] = "Job name is empty"
		l[common.ErrDataBaseExecError] = "Database execute error"
		l[common.InfoDataBaseConnected] = "Database is connected"
	}

	if 0 == len(l) {
		configLog.Warnln("load laguange resource failed")
	}

	return l
}

func getConf(c *common.ConfigLoader, filePath string) {
	file, err := ioutil.ReadFile(filePath)
	if nil != err {
		configLog.Errorf("load %s failed: %v\n", filePath, err)
	}

	err = yaml.Unmarshal(file, &c)
	if nil != err {
		configLog.Errorln("unmarshal yaml file failed", err)
	}
}
