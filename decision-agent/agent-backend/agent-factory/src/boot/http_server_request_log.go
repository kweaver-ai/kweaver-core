package boot

import (
	"os"
	"path/filepath"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capimiddleware"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/httprequesthelper"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

// initHTTPServerRequestLog 初始化HTTP服务端请求日志记录器
// 注意：这个是记录本服务接收到的HTTP请求的日志，不是记录本服务发起的请求的日志
func initHTTPServerRequestLog() {
	// 日志目录
	logDir := "log/received_http_requests"

	if cenvhelper.IsLocalDev() {
		logRootDir := os.Getenv("AGENT_FACTORY_LOCAL_DEV_LOG_ROOT_DIR")
		if logRootDir == "" {
			panic("AGENT_FACTORY_LOCAL_DEV_LOG_ROOT_DIR environment variable is required in local dev mode")
		}

		logDir = filepath.Join(logRootDir, "received_http_requests")
	}

	// 配置请求日志记录器
	// 本地开发环境或debug模式下启用日志
	isEnabled := cenvhelper.IsLocalDev() || cenvhelper.IsDebugMode()
	config := &httprequesthelper.Config{
		Enabled:              isEnabled,
		OutputMode:           httprequesthelper.OutputModeFile, // 输出到文件
		LogDir:               logDir,
		FileNamePattern:      "requests_2006-01-02.log",
		PrettyJSON:           false, // 生产环境不格式化JSON
		MaxBodySize:          10 * 1024,
		IncludeHeaders:       true,
		IncludeResponseBody:  true,
		SingleFileMaxEntries: 500, // 同时记录到 single/all_requests.log，保留最近500条
	}

	// 本地开发环境配置（如需额外配置可在此处添加）
	// if cenvhelper.IsLocalDev() {
	// 	config.OutputMode = httprequesthelper.OutputModeBoth
	// 	config.PrettyJSON = true
	// }

	// 启用请求日志记录
	if err := capimiddleware.InitDefaultRequestLoggerV2(config); err != nil {
		logger.GetLogger().Errorf("Failed to enable server request logging: %v", err)
	}
}
