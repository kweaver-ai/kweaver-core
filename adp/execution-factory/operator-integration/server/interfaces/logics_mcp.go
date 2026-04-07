package interfaces

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
	"github.com/mark3labs/mcp-go/mcp"
)

//go:generate mockgen -source=logics_mcp.go -destination=../mocks/logics_mcp.go -package=mocks

// MCPMode MCP运行模式
type MCPMode string

func (b MCPMode) String() string {
	return string(b)
}

const (
	MCPModeStdioUv  MCPMode = "stdio_uv"  // 标准UV
	MCPModeStdioNpx MCPMode = "stdio_npx" // 标准NPX
	MCPModeSSE      MCPMode = "sse"       // SSE
	MCPModeStream   MCPMode = "stream"    // 流式
)

// MCPCreationType MCP创建类型
type MCPCreationType string

func (b MCPCreationType) String() string {
	return string(b)
}

const (
	MCPCreationTypeCustom       MCPCreationType = "custom"        // 自定义
	MCPCreationTypeToolImported MCPCreationType = "tool_imported" // 工具导入
)

// MCPParseSSERequest MCP解析SSE请求
type MCPParseSSERequest struct {
	Mode    MCPMode           `json:"mode" validate:"required,oneof=stdio_uv stdio_npx sse stream"` // 运行模式
	URL     string            `json:"url" validate:"required,url"`                                  // 请求URL
	Headers map[string]string `json:"headers"`                                                      // 请求头
}

type MCPParseSSEResponse struct {
	Tools          []mcp.Tool            `json:"tools"`     // 工具
	ServerInitInfo *mcp.InitializeResult `json:"init_info"` // 初始化信息
}

// MCPCoreConfigInfo MCP核心信息
type MCPCoreConfigInfo struct {
	Mode    MCPMode           `json:"mode,omitempty" default:"stream" validate:"required,oneof=sse stream"` // 运行模式
	Command string            `json:"command,omitempty"`                                                    // 运行命令
	Args    []string          `json:"args,omitempty"`                                                       // 运行参数
	URL     string            `json:"url,omitempty" validate:"omitempty,url"`                               // 服务URL
	Headers map[string]string `json:"headers,omitempty"`                                                    // 请求头
	Env     map[string]string `json:"env,omitempty"`                                                        // 环境变量
}

type MCPToolConfigInfo struct {
	BoxID           string `json:"box_id"`      // 工具箱ID
	ToolID          string `json:"tool_id"`     // 工具ID
	BoxName         string `json:"box_name"`    // 工具箱名称
	ToolName        string `json:"tool_name"`   // 工具名称
	ToolDescription string `json:"description"` // 工具描述
	UseRule         string `json:"use_rule"`    // 使用规则
}

type MCPAppEndpointRequest struct {
	MCPID string `uri:"mcp_id" validate:"required"` // MCP Server ID
}

// MCPServerAddRequest MCP Server注册请求
type MCPServerAddRequest struct {
	MCPCoreConfigInfo
	BusinessDomainID string               `header:"x-business-domain" validate:"required"`                                       // 业务域ID
	UserID           string               `header:"user_id"`                                                                     // 用户ID，内部使用
	IsPublic         bool                 `header:"is_public"`                                                                   // 是否为公共接口
	CreationType     MCPCreationType      `json:"creation_type" default:"custom" validate:"required,oneof=custom tool_imported"` // 创建类型
	Name             string               `json:"name" validate:"required"`                                                      // MCP Server名称
	Description      string               `json:"description"`                                                                   // 描述信息
	Source           string               `json:"source" default:"custom"`                                                       // 来源
	IsInternal       bool                 `json:"is_internal" default:"false"`                                                   // 是否为内置
	Category         string               `json:"category" default:"other_category"`                                             // 分类
	ToolConfigs      []*MCPToolConfigInfo `json:"tool_configs"`                                                                  // 工具配置
}

// MCPServerAddResponse MCP Server注册响应
type MCPServerAddResponse struct {
	MCPID  string `json:"mcp_id"` // MCP Server ID
	Status string `json:"status"` // 状态
}

// MCPServerDeleteRequest MCP Server删除请求
type MCPServerDeleteRequest struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"` // 业务域ID
	UserID           string `header:"user_id"`                               // 用户ID，内部使用
	IsPublic         bool   `header:"is_public"`                             // 是否为公共接口
	MCPID            string `uri:"mcp_id" validate:"required"`               // MCP Server ID
}

type MCPServerConfigInfo struct {
	MCPCoreConfigInfo `json:",inline"`
	BusinessDomainID  string               `json:"business_domain_id"`                          // 业务域ID
	MCPID             string               `json:"mcp_id"`                                      // MCP Server ID
	Version           int                  `json:"version,omitempty"`                           // MCP Server版本
	CreationType      MCPCreationType      `json:"creation_type,omitempty"`                     // 创建类型
	Name              string               `json:"name,omitempty"`                              // MCP Server名称
	Description       string               `json:"description,omitempty"`                       // 描述信息
	Status            string               `json:"status,omitempty"`                            // 状态
	Source            string               `json:"source,omitempty"`                            // 来源
	IsInternal        bool                 `json:"is_internal"`                                 // 是否为内置
	Category          string               `json:"category,omitempty" default:"other_category"` // 分类
	CreateUser        string               `json:"create_user,omitempty"`                       // 创建用户
	CreateTime        int64                `json:"create_time,omitempty"`                       // 创建时间
	UpdateUser        string               `json:"update_user,omitempty"`                       // 更新用户
	UpdateTime        int64                `json:"update_time,omitempty"`                       // 更新时间
	ReleaseTime       int64                `json:"release_time,omitempty"`                      // 发布时间
	ReleaseUser       string               `json:"release_user,omitempty"`                      // 发布用户
	ToolConfigs       []*MCPToolConfigInfo `json:"tool_configs,omitempty"`                      // 工具配置
}

func (m *MCPServerConfigInfo) ToMapByFields(fields []string) map[string]any {
	// 先序列化为JSON，再反序列化为map
	data, _ := json.Marshal(m)
	var fullMap map[string]interface{}
	_ = json.Unmarshal(data, &fullMap)

	// 如果没有指定fields，返回所有字段
	if len(fields) == 0 {
		return fullMap
	}

	// 根据fields按需返回
	result := make(map[string]interface{})
	for _, field := range fields {
		if value, exists := fullMap[field]; exists {
			result[field] = value
		}
	}

	return result
}

// MCPConnectionInfo MCP连接信息
type MCPConnectionInfo struct {
	SSEURL    string `json:"sse_url,omitempty"`    // SSE URL
	StreamURL string `json:"stream_url,omitempty"` // 流式URL，如果为空，则表示不支持流式
}

// MCPServerListRequest MCP Server列表请求
type MCPServerListRequest struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`                                     // 业务域ID
	UserID           string `header:"user_id"`                                                                   // 用户ID，内部使用
	IsPublic         bool   `header:"is_public"`                                                                 // 是否为公共接口
	Page             int    `form:"page" default:"1" validate:"min=1"`                                           // 页码
	PageSize         int    `form:"page_size" default:"10" validate:"min=1,max=100"`                             // 每页条数
	SortBy           string `form:"sort_by" default:"update_time" validate:"oneof=update_time create_time name"` // 排序字段
	SortOrder        string `form:"sort_order" default:"desc" validate:"oneof=asc desc"`                         // 排序顺序
	Name             string `form:"name"`                                                                        // MCP名称
	Source           string `form:"source"`                                                                      // 来源
	IsInternal       bool   `form:"is_internal"`                                                                 // 是否为内置
	Category         string `form:"category"`                                                                    // 分类
	Status           string `form:"status"`                                                                      // 状态
	CreateUser       string `form:"create_user"`                                                                 // 创建用户
	All              bool   `form:"all"`                                                                         // 是否返回所有信息
	Mode             string `form:"mode" validate:"omitempty,oneof=stdio_uv stdio_npx sse stream"`               // 运行模式
}

// MCPServerListResponse MCP Server列表响应
type MCPServerListResponse struct {
	*ormhelper.QueryResult `json:",inline"`
	Data                   []*MCPServerConfigInfo `json:"data"` // 数据列表
}

// MCPServerDetailRequest MCP Server详情请求
type MCPServerDetailRequest struct {
	UserID   string `header:"user_id"`                 // 用户ID，内部使用
	IsPublic bool   `header:"is_public"`               // 是否为公共接口
	ID       string `uri:"mcp_id" validate:"required"` // MCP Server ID
}

type MCPServerDetailResponse struct {
	BaseInfo       *MCPServerConfigInfo `json:"base_info"`       // MCP Server基本信息
	ConnectionInfo *MCPConnectionInfo   `json:"connection_info"` // MCP连接信息
}

// MCPServerReleaseListRequest MCP Server发布列表请求
type MCPServerReleaseListRequest struct {
	MCPServerListRequest `json:",inline"`
	ReleaseUser          string `form:"release_user"` // 发布者
}

// MCPServerReleaseListResponse MCP Server发布列表响应
type MCPServerReleaseListResponse struct {
	MCPServerListResponse `json:",inline"`
}

type MCPServerReleaseDetailRequest struct {
	MCPServerDetailRequest `json:",inline"`
}

type MCPServerReleaseDetailResponse struct {
	MCPServerDetailResponse `json:",inline"`
}

var MCPFields = []string{
	"mcp_id",
	"name",
	"description",
	"source",
	"category",
	"mode",
	"is_internal",
	"create_user",
	"create_time",
	"update_user",
	"update_time",
	"release_time",
	"release_user",
}

// MCPServerReleaseBatchRequest MCP Server发布批量详情请求
type MCPServerReleaseBatchRequest struct {
	UserID   string `header:"user_id"`                  // 用户ID，内部使用
	IsPublic bool   `header:"is_public"`                // 是否为公共接口
	MCPIDs   string `uri:"mcp_ids" validate:"required"` // MCP Server ID列表，多个ID用逗号分隔
	Fields   string `uri:"fields" validate:"required"`  // 获取MCP信息字段名：（可任意组合，若获取多个，用逗号分隔）
}

// MCPServerUpdateRequest MCP Server更新请求
type MCPServerUpdateRequest struct {
	UserID       string               `header:"user_id" validate:"required"`                                                           // 用户ID，内部使用
	IsPublic     bool                 `header:"is_public"`                                                                             // 是否为公开
	MCPID        string               `json:"mcp_id"`                                                                                  // MCP Server ID
	Name         string               `json:"name,omitempty"`                                                                          // MCP Server名称
	Description  string               `json:"description,omitempty"`                                                                   // 描述信息
	CreationType MCPCreationType      `json:"creation_type" default:"custom" validate:"required,oneof=custom tool_imported"`           // 创建类型
	Mode         MCPMode              `json:"mode,omitempty" default:"stream" validate:"required,oneof=stdio_uv stdio_npx sse stream"` // 运行模式
	URL          string               `json:"url,omitempty" validate:"omitempty,url"`                                                  // 服务URL
	Headers      map[string]string    `json:"headers,omitempty"`                                                                       // 请求头
	Command      string               `json:"command,omitempty"`                                                                       // 运行命令
	Args         []string             `json:"args,omitempty"`                                                                          // 运行参数
	Env          map[string]string    `json:"env,omitempty"`                                                                           // 环境变量
	Source       string               `json:"source,omitempty" default:"custom"`                                                       // 来源
	Category     string               `json:"category,omitempty" default:"other_category"`                                             // 分类
	ToolConfigs  []*MCPToolConfigInfo `json:"tool_configs"`                                                                            // 工具配置
}

// MCPServerUpdateResponse MCP Server更新响应
type MCPServerUpdateResponse struct {
	MCPID  string    `json:"mcp_id"` // MCP Server ID
	Status BizStatus `json:"status"` // 状态
}

// UpdateMCPStatusRequest MCP Server状态更新请求
type UpdateMCPStatusRequest struct {
	UserID   string    `header:"user_id" validate:"required"`                                        // 用户ID，内部使用
	IsPublic bool      `header:"is_public"`                                                          // 是否为公开
	MCPID    string    `uri:"mcp_id" validate:"required"`                                            // MCP Server ID
	Status   BizStatus `json:"status" validate:"required,oneof=unpublish editing published offline"` // 状态
}

// UpdateMCPStatusResponse MCP Server状态更新响应
type UpdateMCPStatusResponse struct {
	MCPID  string    `json:"mcp_id"` // MCP Server ID
	Status BizStatus `json:"status"` // 状态
}

// MCPToolDebugRequest MCP工具调试请求
type MCPToolDebugRequest struct {
	UserID     string         `header:"user_id" validate:"required"` // 用户ID，内部使用
	IsPublic   bool           `header:"is_public"`                   // 是否为公开
	MCPID      string         `uri:"mcp_id" validate:"required"`     // MCP Server ID
	ToolName   string         `uri:"tool_name" validate:"required"`  // 工具名称
	Parameters map[string]any `json:"parameters"`                    // 工具请求参数
}

// MCPToolDebugResponse MCP工具调试响应
type MCPToolDebugResponse struct {
	Content []mcp.Content `json:"content"`  // 工具调用结果内容
	IsError bool          `json:"is_error"` // 是否为错误
}

type MCPProxyToolListRequest struct {
	UserID string `header:"user_id"`                 // 用户ID，内部使用
	MCPID  string `uri:"mcp_id" validate:"required"` // MCP Server ID
}

type MCPProxyToolListResponse struct {
	Tools []mcp.Tool `json:"tools"` // 工具
}

// MCPProxyCallToolRequest MCP工具调用请求
type MCPProxyCallToolRequest struct {
	UserID     string         `header:"user_id" validate:"required"`  // 用户ID,内部使用
	MCPID      string         `uri:"mcp_id" validate:"required"`      // MCP Server ID
	ToolName   string         `json:"tool_name" validate:"required"`  // 工具名称
	Parameters map[string]any `json:"parameters" validate:"required"` // 工具请求参数
}

// MCPProxyCallToolResponse MCP工具调用响应
type MCPProxyCallToolResponse struct {
	Content []mcp.Content `json:"content"`  // 工具调用结果内容
	IsError bool          `json:"is_error"` // 是否为错误
}

// MCPBuiltinRegisterRequest MCP内置注册请求
type MCPBuiltinRegisterRequest struct {
	MCPCoreConfigInfo `json:",inline"`
	BusinessDomainID  string           `header:"x-business-domain" validate:"required"`             // 业务域ID
	UserID            string           `header:"user_id"`                                           // 用户ID,内部使用
	IsPublic          bool             `header:"is_public"`                                         // 是否为公开
	MCPID             string           `json:"mcp_id"`                                              // MCP Server ID
	Name              string           `json:"name,omitempty"`                                      // MCP Server名称
	Description       string           `json:"description,omitempty"`                               // 描述信息
	Status            string           `json:"status,omitempty"`                                    // 状态
	Source            string           `json:"source,omitempty"`                                    // 来源
	IsInternal        bool             `json:"is_internal"`                                         // 是否为内置
	CreateUser        string           `json:"create_user,omitempty"`                               // 创建用户
	CreateTime        int64            `json:"create_time,omitempty"`                               // 创建时间
	UpdateUser        string           `json:"update_user,omitempty"`                               // 更新用户
	UpdateTime        int64            `json:"update_time,omitempty"`                               // 更新时间
	ReleaseTime       int64            `json:"release_time,omitempty"`                              // 发布时间
	ReleaseUser       string           `json:"release_user,omitempty"`                              // 发布用户
	ProtectedFlag     bool             `json:"protected_flag"`                                      // 版本周期内保护标志，true表示保护，false表示不保护，这个字段主要作用于内置MCP Server
	ConfigVersion     string           `json:"config_version" validate:"required"`                  // 配置版本
	ConfigSource      ConfigSourceType `json:"config_source" validate:"required,oneof=auto manual"` // 配置来源(自动/手动)
}

type MCPBuiltinRegisterResponse struct {
	MCPID  string    `json:"mcp_id"` // MCP Server ID
	Status BizStatus `json:"status"` // 状态
}

// MCPBuiltinUnregisterRequest MCP内置注销请求
type MCPBuiltinUnregisterRequest struct {
	UserID   string `header:"user_id" validate:"required"` // 用户ID,内部使用
	IsPublic bool   `header:"is_public"`                   // 是否为公开
	MCPID    string `uri:"mcp_id" validate:"required"`     // MCP Server ID
}

// IMCPManageService MCP管理接口
type IMCPManageService interface {
	// ParseSSE 解析SSE MCPServer
	ParseSSE(ctx context.Context, req *MCPParseSSERequest) (*MCPParseSSEResponse, error)
	// AddMCPServer 注册MCPServer
	AddMCPServer(ctx context.Context, req *MCPServerAddRequest) (*MCPServerAddResponse, error)
	// DeleteMCPServer 删除MCPServer
	DeleteMCPServer(ctx context.Context, req *MCPServerDeleteRequest) error
	// List 获取MCPServer列表
	QueryPage(ctx context.Context, req *MCPServerListRequest) (*MCPServerListResponse, error)
	// Detail 获取MCPServer详情
	GetDetail(ctx context.Context, req *MCPServerDetailRequest) (*MCPServerDetailResponse, error)
	// UpdateMCPServer 编辑MCPServer
	UpdateMCPServer(ctx context.Context, req *MCPServerUpdateRequest) (*MCPServerUpdateResponse, error)
	// UpdateMCPStatus 更新MCP Server状态
	UpdateMCPStatus(ctx context.Context, req *UpdateMCPStatusRequest) (*UpdateMCPStatusResponse, error)
	// DebugTool 工具调试
	DebugTool(ctx context.Context, req *MCPToolDebugRequest) (*MCPToolDebugResponse, error)
}

// IMCPReleaseService MCP市场接口
type IMCPReleaseService interface {
	// getList 获取MCP Server发布列表
	QueryRelease(ctx context.Context, req *MCPServerReleaseListRequest) (*MCPServerReleaseListResponse, error)
	// getDetail 获取MCP Server发布详情
	GetReleaseDetail(ctx context.Context, req *MCPServerReleaseDetailRequest) (*MCPServerReleaseDetailResponse, error)
	// QueryReleaseBatch 批量获取MCP Server发布详情
	QueryReleaseBatch(ctx context.Context, req *MCPServerReleaseBatchRequest) ([]map[string]any, error)
}

// IMCPExecuteService MCP代理接口
type IMCPExecuteService interface {
	// GetMCPTools 获取MCP Server发布列表
	GetMCPTools(ctx context.Context, req *MCPProxyToolListRequest) (*MCPProxyToolListResponse, error)
	// CallMCPTool 调用MCP工具
	CallMCPTool(ctx context.Context, req *MCPProxyCallToolRequest) (*MCPProxyCallToolResponse, error)
}

// IMCPBuiltinService MCP内置接口
type IMCPBuiltinService interface {
	// RegisterBuiltinMCPServer 注册内置MCP Server
	RegisterBuiltinMCPServer(ctx context.Context, req *MCPBuiltinRegisterRequest) (*MCPBuiltinRegisterResponse, error)
	// UnregisterBuiltinMCPServer 注销内置MCP Server
	UnregisterBuiltinMCPServer(ctx context.Context, req *MCPBuiltinUnregisterRequest) error
}

// IMCPService MCP服务接口
type IMCPService interface {
	// MCPManageService MCP管理接口
	IMCPManageService
	// MCPReleaseService MCP市场接口
	IMCPReleaseService
	// MCPExecuteService MCP代理接口
	IMCPExecuteService
	// MCPBuiltinService MCP内置接口
	IMCPBuiltinService
	// IMCPImpexService MCP导入导出接口
	IMCPImpexService
	// UpgradeMCPInstance 升级MCP Server实例
	UpgradeMCPInstance(ctx context.Context, mcpID string) error
	// MCPInstancConfig MCP Server实例配置接口
	IMCPInstancConfig
	// IMCPToolExecutor MCP工具执行器
	IMCPToolExecutor
}

// IMCPImpexService MCP导入导出接口
type IMCPImpexService interface {
	// 导入导出
	// Impex[*MCPImpexData]
	Import(ctx context.Context, tx *sql.Tx, mode ImportType, data *ComponentImpexConfigModel, userID string) (err error)
	// Export 导出配置
	Export(ctx context.Context, req *ExportReq) (data *ComponentImpexConfigModel, err error)
}

// IMCPToolExecutor MCP工具执行器
type IMCPToolExecutor interface {
	ExecuteTool(ctx context.Context, mcpToolID string, params HTTPRequestParams) (*HTTPResponse, error)
}

// MCP 实例相关配置接口
type IMCPInstancConfig interface {
	// GetMCPInstanceConfig 获取MCP Server实例配置
	GetMCPInstanceConfig(ctx context.Context, mcpID string, mode MCPMode) (*MCPInstancConfigInfo, error)
}

// MCPInstancConfigInfo MCP Server实例配置信息
type MCPInstancConfigInfo struct {
	MCPID   string            // MCP Server ID
	URL     string            // MCP Server URL
	Headers map[string]string // MCP Server请求头
	Mode    MCPMode           // MCP Server运行模式
	Version int               // MCP Server版本（tool_imported场景使用）
}
