package httprequesthelper

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
)

// Logger HTTP请求日志记录器
type Logger struct {
	config       *Config
	formatter    *Formatter
	writer       Writer
	singleWriter *SingleFileWriter
	mu           sync.RWMutex
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// NewLogger 创建新的日志记录器
func NewLogger(config *Config) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	formatter := NewFormatter(config.PrettyJSON, config.MaxBodySize)

	var writer Writer

	var err error

	switch config.OutputMode {
	case OutputModeFile:
		writer, err = NewFileWriter(config)
		if err != nil {
			return nil, err
		}
	case OutputModeConsole:
		writer = NewConsoleWriter()
	case OutputModeBoth:
		fileWriter, err := NewFileWriter(config)
		if err != nil {
			return nil, err
		}

		writer = NewMultiWriter(NewConsoleWriter(), fileWriter)
	default:
		writer = NewConsoleWriter()
	}

	logger := &Logger{
		config:    config,
		formatter: formatter,
		writer:    writer,
	}

	// 如果配置了 SingleFileMaxEntries，则初始化 singleWriter
	if config.SingleFileMaxEntries > 0 {
		singleWriter, err := NewSingleFileWriter(config)
		if err != nil {
			// single writer 创建失败不影响主日志记录
			// 只是不记录到 single 文件
		} else {
			logger.singleWriter = singleWriter
		}
	}

	return logger, nil
}

// GetDefaultLogger 获取默认日志记录器（单例）
func GetDefaultLogger() *Logger {
	once.Do(func() {
		var err error

		defaultLogger, err = NewLogger(DefaultConfig())
		if err != nil {
			// 如果创建失败，使用控制台输出
			defaultLogger = &Logger{
				config:    DefaultConfig(),
				formatter: NewFormatter(false, 10*1024),
				writer:    NewConsoleWriter(),
			}
		}
	})

	return defaultLogger
}

// IsEnabled 检查是否启用
func (l *Logger) IsEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.config.Enabled
}

// SetEnabled 设置是否启用
func (l *Logger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.Enabled = enabled
}

// SetPrettyJSON 设置是否格式化JSON
func (l *Logger) SetPrettyJSON(pretty bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.config.PrettyJSON = pretty
	l.formatter = NewFormatter(pretty, l.config.MaxBodySize)
}

// LogRequest 记录请求和响应
// ctx 用于获取用户ID，如果用户ID不为空，日志文件名中会包含用户ID
func (l *Logger) LogRequest(ctx context.Context, req *http.Request, reqBody string, statusCode int, respHeaders http.Header, respBody string, duration time.Duration) {
	if !l.IsEnabled() {
		return
	}

	// 从 ctx 中获取用户ID
	userID := chelper.GetUserIDFromCtx(ctx)

	reqRecord := NewRequestRecord(req, reqBody, l.config.IncludeHeaders)
	respRecord := NewResponseRecord(statusCode, respHeaders, respBody, duration, l.config.IncludeHeaders)

	logRecord := &LogRecord{
		Request:  reqRecord,
		Response: respRecord,
	}

	formatted := l.formatter.Format(logRecord)
	_ = l.writer.Write(formatted, userID)

	// 同时写入到 single 日志
	if l.singleWriter != nil {
		_ = l.singleWriter.Write(formatted)
	}
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	if l.singleWriter != nil {
		_ = l.singleWriter.Close()
	}

	return l.writer.Close()
}
