// Package interfaces 定义接口
// @file infra.go
// @description: 定义基础设施接口
package interfaces

import (
	"context"
	"net/url"
)

// ContextKey Context Value Key
//
//go:generate mockgen -source=infra.go -destination=../mocks/infra.go -package=mocks
type ContextKey string

const (
	// KeyToken context中token key
	KeyToken ContextKey = "token"
	// KeyIP context中ip key
	KeyIP ContextKey = "ip"
	// OperationID API操作唯一标识
	OperationID ContextKey = "operationID"
	// XLanguageKey 语言类型
	XLanguageKey ContextKey = "X-Language"
	// FileNameKey 文件名
	FileNameKey ContextKey = "FileName"
	// Headers header 请求参数
	Headers ContextKey = "headers"
	// UserAgent user agent
	UserAgent ContextKey = "User-Agent"
	// IsPublic 是否公开
	IsPublic ContextKey = "is_public"
	// KeyRequestID 请求ID
	KeyRequestID ContextKey = "request_id"
	// 上下文键
	KeyResponseWriter ContextKey = "response_writer" // 响应写入器
	KeyExecutionMode  ContextKey = "execution_mode"  // 执行模式
	KeyStreamingMode  ContextKey = "streaming_mode"  // 流式模式
	// KeyAccountAuthContext 账户认证上下文
	KeyAccountAuthContext ContextKey = "account_auth_context"
	// XBusinessDomain 业务域id
	XBusinessDomain ContextKey = "x-business-domain"
)

// HeaderKey 上下文键
type HeaderKey string

const (
	// HeaderXBusinessDomain 业务域id头参数
	HeaderXBusinessDomain HeaderKey = "x-business-domain"
	// HeaderXAccountID 账户ID头参数
	HeaderXAccountID HeaderKey = "x-account-id"
	// HeaderXAccountType 账户类型头参数
	HeaderXAccountType HeaderKey = "x-account-type"
	// HeaderUserID 用户ID
	HeaderUserID HeaderKey = "user_id"
)

const (
	// DefaultBusinessDomain 默认业务域
	DefaultBusinessDomain = "bd_public"
)

const (
	HTTP  = "http"  // http协议
	HTTPS = "https" // https协议
)

// Logger 日志接口
type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	WithContext(ctx context.Context) Logger
}

// HTTPClient HTTP客户端服务接口
type HTTPClient interface {
	Get(ctx context.Context, url string, queryValues url.Values, headers map[string]string) (respCode int, respData interface{}, err error)
	GetNoUnmarshal(ctx context.Context, url string, queryValues url.Values, headers map[string]string) (respCode int, respBody []byte, err error)
	Delete(ctx context.Context, url string, headers map[string]string) (respCode int, respData interface{}, err error)
	DeleteNoUnmarshal(ctx context.Context, url string, headers map[string]string) (respCode int, respBody []byte, err error)
	Post(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error)
	PostNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error)
	Put(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error)
	PutNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error)
	Patch(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error)
	PatchNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error)
	PostStream(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (chan string, chan error, error)
}

// Cache 缓存接口
type Cache interface {
	Set(key string, value interface{})
	Get(key string) (interface{}, bool)
	Delete(key string)
	Size() int
}

// MetricLogger 指标日志记录器
type MetricLogger interface {
	Log(ctx context.Context, logType string, params interface{}) (err error)
}

// Validator 验证接口:用于验证算子名称、描述、单次导入个数、导入数据大小等
type Validator interface {
	ValidateOperatorName(ctx context.Context, name string) (err error)
	ValidateOperatorDesc(ctx context.Context, desc string) (err error)
	ValidateOperatorImportCount(ctx context.Context, count int64) (err error)
	ValidateOperatorImportSize(ctx context.Context, size int64) (err error)
	ValidatorToolBoxName(ctx context.Context, name string) (err error)
	ValidatorToolBoxDesc(ctx context.Context, desc string) (err error)
	ValidatorToolName(ctx context.Context, name string) (err error)
	ValidatorToolDesc(ctx context.Context, desc string) (err error)
	ValidatorIntCompVersion(ctx context.Context, version string) (err error)
	ValidatorMCPName(ctx context.Context, name string) (err error)
	ValidatorMCPDesc(ctx context.Context, desc string) (err error)
	ValidatorCategoryName(ctx context.Context, name string) (err error)
	ValidatorStruct(ctx context.Context, obj interface{}) (err error)
	ValidatorURL(ctx context.Context, url string) (err error)
	VisitorParameterDef(ctx context.Context, paramDef *ParameterDef) (err error)
}
