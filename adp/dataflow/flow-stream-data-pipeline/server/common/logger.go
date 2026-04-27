package common

import (
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	"flow-stream-data-pipeline/version"
)

const (
	// 日志保存位置
	logFileName = "/opt/pipeline/logs/pipeline.log"
)

// 获取日志句柄
func init() {
	setting := logger.LogSetting{
		LogServiceName: version.ServerName,
		LogFileName:    logFileName,
		LogLevel:       "info",
		DevelopMode:    false,
		MaxAge:         100,
		MaxBackups:     20,
		MaxSize:        100,
	}
	logger.InitGlobalLogger(setting)
}

// SetLogSetting 设置日志配置
func SetLogSetting(setting logger.LogSetting) {
	setting.LogServiceName = version.ServerName
	setting.LogFileName = logFileName
	logger.InitGlobalLogger(setting)
}
