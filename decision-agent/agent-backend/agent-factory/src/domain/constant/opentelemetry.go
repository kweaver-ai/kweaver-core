package constant

// 定义context key类型
type contextKey string

const (
	// LoggerKey 上下文中logger的key
	LoggerKey contextKey = "logger"
	// MetricsKey 上下文中metrics的key
	MetricsKey contextKey = "metrics"
)
