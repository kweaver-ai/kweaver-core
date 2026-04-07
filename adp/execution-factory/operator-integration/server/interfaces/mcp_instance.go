package interfaces

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mark3labs/mcp-go/server"
)

// MCPToolArguments MCP工具参数
type MCPToolArguments struct {
	Headers     map[string]any `json:"header"`
	Body        any            `json:"body"`
	QueryParams map[string]any `json:"query"`
	PathParams  map[string]any `json:"path"`
}

func (args *MCPToolArguments) ToHTTPRequestParams() map[string]any {
	return map[string]any{
		"header": args.Headers,
		"body":   args.Body,
		"query":  args.QueryParams,
		"path":   args.PathParams,
	}
}

// MCPRuntimeConfig MCP运行时配置
type MCPRuntimeConfig struct {
	MCPID        string                 `json:"mcp_id"`
	Version      int                    `json:"version"`
	Name         string                 `json:"name"`
	Instructions string                 `json:"instructions"`
	CreationType string                 `json:"creation_type"`
	Tools        []*MCPToolDeployConfig `json:"tools"`
}

type MCPToolDeployConfig struct {
	ToolID      string          `json:"tool_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// MCPServerInstance MCP Server运行时实例
type MCPServerInstance struct {
	Config           *MCPRuntimeConfig
	MCPServer        *server.MCPServer
	StreamServer     *server.StreamableHTTPServer
	SSEServer        *server.SSEServer
	StreamRoutePath  string
	SSERoutePath     string
	MessageRoutePath string
	IsDisabled       bool
	CreatedAt        *time.Time
	// ActiveStreamConn / ActiveSSEConn 用于实例池淘汰保护：
	// 当存在活跃连接时（>0），实例不会被 LRU/TTL 淘汰。
	// 计数在 driveradapters/mcp 的 HTTP handler 中通过 atomic 维护。
	ActiveStreamConn int64
	ActiveSSEConn    int64
}

// InstanceService MCP 实例服务接口
type InstanceService interface {
	CreateMCPInstance(ctx context.Context, req *MCPInstanceCreateRequest) (*MCPInstanceCreateResponse, error)
	UpdateMCPInstance(ctx context.Context, mcpID string, version int, req *MCPInstanceUpdateRequest) (*MCPInstanceUpdateResponse, error)
	DeleteMCPInstance(ctx context.Context, mcpID string, version int) error
	DeleteAllMCPInstances(ctx context.Context, mcpID string) error
	UpgradeMCPInstance(ctx context.Context, req *MCPInstanceCreateRequest) (*MCPInstanceCreateResponse, error)
	GetMCPInstance(ctx context.Context, mcpID string, version int) (*MCPServerInstance, error)
}
