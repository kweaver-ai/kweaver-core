package chelper

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"go.uber.org/zap"
)

var simpleStdoutLogger *zap.SugaredLogger

// GetStdoutLogger 不依赖配置文件，直接输出到stdout
func GetStdoutLogger() icmp.Logger {
	if simpleStdoutLogger != nil {
		return simpleStdoutLogger
	}

	simpleStdoutLogger = logger.GetLogger()

	return simpleStdoutLogger
}
