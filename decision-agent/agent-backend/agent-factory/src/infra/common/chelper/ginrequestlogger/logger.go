package ginrequestlogger

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httprequesthelper"
)

// NewRequestLogger 创建请求日志记录器
// config 为nil时使用默认配置
func NewRequestLogger(config *httprequesthelper.Config) (*RequestLogger, error) {
	logger, err := httprequesthelper.NewLogger(config)
	if err != nil {
		return nil, err
	}

	return &RequestLogger{
		logger: logger,
	}, nil
}

// Close 关闭日志记录器
func (r *RequestLogger) Close() error {
	return r.logger.Close()
}
