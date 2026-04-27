package model

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source=mcp_tool.go -destination=../../mocks/model_mcp_tool.go -package=mocks

// MCPToolDB MCP工具表
type MCPToolDB struct {
	ID          int64  `json:"id" db:"f_id"`                   // 主键
	MCPToolID   string `json:"mcp_tool_id" db:"f_mcp_tool_id"` // mcp_tool_id
	MCPID       string `json:"mcp_id" db:"f_mcp_id"`           // mcp_id
	MCPVersion  int    `json:"mcp_version" db:"f_mcp_version"` // mcp_version
	BoxID       string `json:"box_id" db:"f_box_id"`           // box_id
	BoxName     string `json:"box_name" db:"f_box_name"`       // box_name
	ToolID      string `json:"tool_id" db:"f_tool_id"`         // tool_id
	Name        string `json:"name" db:"f_name"`               // 工具名称
	Description string `json:"description" db:"f_description"` // 描述信息
	UseRule     string `json:"use_rule" db:"f_use_rule"`       // 使用规则
	CreateUser  string `json:"create_user" db:"f_create_user"` // 创建者
	CreateTime  int64  `json:"create_time" db:"f_create_time"` // 创建时间
	UpdateUser  string `json:"update_user" db:"f_update_user"` // 编辑者
	UpdateTime  int64  `json:"update_time" db:"f_update_time"` // 编辑时间
}

type DBMCPTool interface {
	// 批量插入MCP工具配置信息
	BatchInsert(ctx context.Context, tx *sql.Tx, tools []*MCPToolDB) (err error)
	// 批量删除MCP工具配置信息
	DeleteByMCPIDAndVersion(ctx context.Context, tx *sql.Tx, mcpID string, mcpVersion int) (err error)
	// 根据MCPID和版本号查询MCP工具配置信息
	SelectListByMCPIDAndVersion(ctx context.Context, tx *sql.Tx, mcpID string, mcpVersion int) (tools []*MCPToolDB, err error)
	// 根据MCPID列表查询MCP工具配置信息
	SelectListByMCPIDS(ctx context.Context, tx *sql.Tx, mcpIDs []string) (tools []*MCPToolDB, err error)
	// 根据MCPToolID查询MCP工具配置信息
	SelectByMCPToolID(ctx context.Context, tx *sql.Tx, mcpToolID string) (tool *MCPToolDB, err error)
}
