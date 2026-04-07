package ginrequestlogger

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httprequesthelper"
)

// InitDefaultRequestLogger 初始化默认的请求日志记录器（单例）
// 如果已经初始化过，则忽略本次调用
func InitDefaultRequestLogger(config *httprequesthelper.Config) error {
	var initErr error

	defaultRequestLoggerOnce.Do(func() {
		var err error

		defaultRequestLogger, err = NewRequestLogger(config)
		if err != nil {
			initErr = err
		}
	})

	return initErr
}

// GetDefaultRequestLogger 获取默认的请求日志记录器
// 如果未初始化，返回nil
func GetDefaultRequestLogger() *RequestLogger {
	return defaultRequestLogger
}
