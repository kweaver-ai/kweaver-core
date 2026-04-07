package interfaces

import (
	"fmt"
	"strings"
	"time"
)

const (
	// DefaultPageSize 默认每页大小
	DefaultPageSize = 10
	// DefaultPage  默认页码
	DefaultPage = 1
	// MaxPageSize 最大每页大小
	MaxPageSize = 1000

	// OSSGatewayPrefix OSS 网关技能资产前缀
	OSSGatewayPrefix = "execution-factory"
)

var (
	// 当前服务配置
	AOIServerURL = "http://agent-operator-integration:9000"
	// AOPInternalV1Prefix AOP 内部 V1 前缀
	AOPInternalV1Prefix = "/api/agent-operator-integration/internal-v1"
)

// SetAOIFuncExecPath 获取函数执行路径
func SetAOIFuncExecPath(version string) string {
	return strings.ReplaceAll(GetAOIFuncExecPath(), ":version", version)
} // GetAOIFuncExecPath 获取函数执行路径
func GetAOIFuncExecPath() string {
	return fmt.Sprintf("%s/function/exec/:version", AOPInternalV1Prefix)
}

// CommonPageResult 通用分页返回结果
type CommonPageResult struct {
	TotalCount int  `json:"total"`       // 总记录数
	Page       int  `json:"page"`        // 当前页码
	PageSize   int  `json:"page_size"`   // 每页大小
	TotalPage  int  `json:"total_pages"` // 总页数
	HasNext    bool `json:"has_next"`    // 是否有下一页
	HasPrev    bool `json:"has_prev"`    // 是否有上一页
}

// PtrBizIdentifiable 业务ID可识别接口指针
type PtrBizIdentifiable[T any] interface {
	*T
	GetBizID() string // 获取业务ID
}

// QueryResponse 通用查询响应结构
type QueryResponse[T any] struct {
	CommonPageResult `json:",inline"`
	Data             []*T `json:"data"` // 数据列表
}

type ResultStatus string

const (
	ResultStatusFailed  ResultStatus = "failed"
	ResultStatusSuccess ResultStatus = "success"
)

// MetadataType 元数据类型
type MetadataType string

const (
	// MetadataTypeAPI API 源数据类型
	MetadataTypeAPI MetadataType = "openapi"
	// MetadataTypeFunc 函数源数据类型
	MetadataTypeFunc MetadataType = "function"
)

// ExecutionMode 执行模式
type ExecutionMode string

func (e ExecutionMode) String() string {
	return string(e)
}

const (
	ExecutionModeSync   ExecutionMode = "sync"   // 同步执行
	ExecutionModeAsync  ExecutionMode = "async"  // 异步执行
	ExecutionModeStream ExecutionMode = "stream" // 流式执行
)

// StreamingMode 定义流式传输类型
type StreamingMode string

const (
	StreamingModeSSE  StreamingMode = "sse"
	StreamingModeHTTP StreamingMode = "http"
)

// HTTPRequest API请求
type HTTPRequest struct {
	ClientID          string        `json:"client_id"` // 客户端ID
	Timeout           time.Duration `json:"timeout" validate:"gte=0"`
	ExecutionMode     ExecutionMode `json:"execution_mode" validate:"required,oneof=sync async stream"`
	HTTPRouter        `json:",inline"`
	HTTPRequestParams `json:",inline"`
}

// HTTPRouter HTTP路由
type HTTPRouter struct {
	URL    string `json:"url" validate:"required"`
	Method string `json:"method" validate:"required"`
}

// APIRouter API路由
// @description: API路由
type APIRouter struct {
	ServerURL  string `json:"server_url" validate:"required"` // 服务器URL
	HTTPRouter `json:",inline"`
}

// HTTPRequestParams HTTP请求参数
type HTTPRequestParams struct {
	Headers     map[string]any    `json:"header"`
	Body        interface{}       `json:"body"`
	QueryParams map[string]any    `json:"query"`
	PathParams  map[string]string `json:"path"`
}

// FuncRequestParams 函数请求参数
type FuncRequestParams struct {
	InputParams map[string]any `json:"inputs,omitempty" form:"inputs"` // 输入参数列表
	Code        string         `json:"code"`                           // 函数代码
}

// HTTPResponse API响应
type HTTPResponse struct {
	StatusCode int            `json:"status_code"` // 状态码
	Headers    map[string]any `json:"headers"`     // 响应头
	Body       interface{}    `json:"body"`        // 响应体
	Error      string         `json:"error"`       // 错误信息
	Duration   int64          `json:"duration_ms"` // 响应时间
}

// BizStatus 状态
type BizStatus string

func (b BizStatus) String() string {
	return string(b)
}

const (
	BizStatusUnpublish BizStatus = "unpublish" // 未发布
	BizStatusPublished BizStatus = "published" // 已发布
	BizStatusOffline   BizStatus = "offline"   // 已下架
	BizStatusEditing   BizStatus = "editing"   // 已发布编辑中
)

// OutboxMessageReq 消息事件请求
type OutboxMessageReq struct {
	EventID   string                 `json:"event_id"`
	EventType OutboxMessageEventType `json:"event_type" validate:"required"`
	Topic     string                 `json:"topic" validate:"required"`
	Payload   string                 `json:"payload" validate:"required"`
}

// OutboxMessageEventType 消息事件类型
type OutboxMessageEventType string

// String 返回字符串
func (eventType OutboxMessageEventType) String() string {
	return string(eventType)
}

const (
	OutboxMessageEventTypeAuditLog OutboxMessageEventType = "audit_log" // 审计日志
)
