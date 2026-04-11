package boot

import (
	"os"
	"path/filepath"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/httphelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/httprequesthelper"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
)

// initHTTPClientRequestLog 初始化HTTP客户端请求日志记录器
// 注意：这个是当本服务的HTTP客户端发起请求其他服务的接口等时，记录的请求日志。不是记录接受到的请求的日志
func initHTTPClientRequestLog() {
	// 日志目录
	logDir := "log/dependence_http_requests"

	if cenvhelper.IsLocalDev() {
		logRootDir := os.Getenv("AGENT_FACTORY_LOCAL_DEV_LOG_ROOT_DIR")
		if logRootDir == "" {
			panic("AGENT_FACTORY_LOCAL_DEV_LOG_ROOT_DIR environment variable is required in local dev mode")
		}

		logDir = filepath.Join(logRootDir, "dependence_http_requests")
	}

	// 配置请求日志记录器
	isDebugMode := cenvhelper.IsDebugMode()
	config := &httprequesthelper.Config{
		Enabled:             isDebugMode,
		OutputMode:          httprequesthelper.OutputModeFile, // 输出到文件
		LogDir:              logDir,
		FileNamePattern:     "requests_2006-01-02.log",
		PrettyJSON:          false, // 生产环境不格式化JSON
		MaxBodySize:         10 * 1024,
		IncludeHeaders:      true,
		IncludeResponseBody: true,
	}

	// 本地开发环境配置（如需额外配置可在此处添加）
	// if cenvhelper.IsLocalDev() {
	// 	config.OutputMode = httprequesthelper.OutputModeBoth
	// 	config.PrettyJSON = true
	// }

	// 启用请求日志记录
	if err := httphelper.EnableHTTPClientRequestLogging(config); err != nil {
		logger.GetLogger().Errorf("Failed to enable request logging: %v", err)
	}
}
