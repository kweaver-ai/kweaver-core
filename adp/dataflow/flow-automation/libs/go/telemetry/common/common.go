package common

import (
	"strings"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/config"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/log"
)

type TelemetryConf struct {
	LogLevel      string `json:"log_level" yaml:"log_level" mapstructure:"log_level"` //日志级别，从0~7，0代表全部输出，7代表关闭输出。all,trace,debug,info,warn,error,fatal,off
	TraceURL      string `json:"trace_url" yaml:"trace_url" mapstructure:"trace_url"`
	LogURL        string `json:"log_url" yaml:"log_url" mapstructure:"log_url"`
	ServerName    string `json:"server_name" yaml:"server_name" mapstructure:"server_name"`
	ServerVersion string `json:"server_version" yaml:"server_version" mapstructure:"server_version"`
	TraceEnabled  bool   `json:"trace_enabled" yaml:"trace_enabled" mapstructure:"trace_enabled"`
	// 新exporter配置注入方式
	Exporters *config.ExportersTypConfig `json:"exporters" yaml:"exporters" mapstructure:"exporters"`
}

const (
	ALL   = log.AllLevel
	TRACE = log.TraceLevel
	DEBUG = log.DebugLevel
	INFO  = log.InfoLevel
	WARN  = log.WarnLevel
	ERROR = log.ErrorLevel
	FATAL = log.FatalLevel
	OFF   = log.OffLevel
)

func (tc *TelemetryConf) GetLogLevel() int {
	logType := strings.ToUpper(tc.LogLevel)
	switch logType {
	case "ALL":
		return ALL
	case "TRACE":
		return TRACE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	case "OFF":
		return OFF
	default:
		return ERROR
	}
}
