package httprequesthelper

import (
	"os"
	"path/filepath"
	"time"
)

// Config 请求日志记录器配置
type Config struct {
	// Enabled 是否启用日志记录
	Enabled bool

	// OutputMode 输出模式: "file", "console", "both"
	OutputMode OutputMode

	// LogDir 日志文件根目录
	LogDir string

	// FileNamePattern 文件名模式，支持时间格式化
	// 例如: "requests_2006-01-02.log"
	FileNamePattern string

	// PrettyJSON 是否格式化JSON输出（换行缩进）
	PrettyJSON bool

	// MaxBodySize 最大记录的body大小（字节），超过则截断，0表示不限制
	MaxBodySize int

	// IncludeHeaders 是否记录请求头
	IncludeHeaders bool

	// IncludeResponseBody 是否记录响应体
	IncludeResponseBody bool

	// SingleFileMaxEntries single日志最大保留条目数，0表示禁用single日志
	// single日志会将所有请求记录到 LogDir/single/all_requests.log 文件中
	SingleFileMaxEntries int
}

// OutputMode 输出模式
type OutputMode string

const (
	OutputModeFile    OutputMode = "file"
	OutputModeConsole OutputMode = "console"
	OutputModeBoth    OutputMode = "both"
)

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Enabled:             true,
		OutputMode:          OutputModeFile,
		LogDir:              "log/requests",
		FileNamePattern:     "requests_2006-01-02.log",
		PrettyJSON:          false,
		MaxBodySize:         10 * 1024, // 10KB
		IncludeHeaders:      true,
		IncludeResponseBody: true,
	}
}

// GetLogFilePath 获取当前日志文件路径
// userID 可选，如果不为空则文件名中包含用户ID
func (c *Config) GetLogFilePath(userID string) string {
	fileName := time.Now().Format(c.FileNamePattern)
	if userID != "" {
		// 在文件名中追加用户ID，如: requests_2006-01-02.log -> requests_2006-01-02_user123.log
		ext := filepath.Ext(fileName)
		base := fileName[:len(fileName)-len(ext)]
		fileName = base + "_" + userID + ext
	}

	return filepath.Join(c.LogDir, fileName)
}

// EnsureLogDir 确保日志目录存在
func (c *Config) EnsureLogDir() error {
	if c.OutputMode == OutputModeConsole {
		return nil
	}

	return os.MkdirAll(c.LogDir, 0o755)
}

// GetSingleFilePath 获取single日志文件路径
func (c *Config) GetSingleFilePath() string {
	return filepath.Join(c.LogDir, "single", "all_requests.log")
}

// EnsureSingleFileDir 确保single日志目录存在
func (c *Config) EnsureSingleFileDir() error {
	singleDir := filepath.Join(c.LogDir, "single")
	return os.MkdirAll(singleDir, 0o755)
}
