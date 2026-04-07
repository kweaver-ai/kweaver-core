package model

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

// MCPServerReleaseDB MCP Server发布表对应的结构体
//
//go:generate mockgen -source=mcp_server_release.go -destination=../../mocks/model_mcp_server_release.go -package=mocks
type MCPServerReleaseDB struct {
	ID           int64  `json:"f_id" db:"f_id"`                       // id
	MCPID        string `json:"f_mcp_id" db:"f_mcp_id"`               // mcp_id
	CreateUser   string `json:"f_create_user" db:"f_create_user"`     // 创建者
	CreateTime   int64  `json:"f_create_time" db:"f_create_time"`     // 创建时间
	UpdateUser   string `json:"f_update_user" db:"f_update_user"`     // 编辑者
	UpdateTime   int64  `json:"f_update_time" db:"f_update_time"`     // 编辑时间
	CreationType string `json:"f_creation_type" db:"f_creation_type"` // 创建类型
	Name         string `json:"f_name" db:"f_name"`                   // MCP Server名称，全局唯一
	Description  string `json:"f_description" db:"f_description"`     // 描述信息
	Mode         string `json:"f_mode" db:"f_mode"`                   // 通信模式（sse、streamable、stdio_npx、stdio_uvx）
	URL          string `json:"f_url" db:"f_url"`                     // 通信地址,SSE/Streamable模式下的服务URL
	Headers      string `json:"f_headers" db:"f_headers"`             // http请求头,JSON字符串
	Command      string `json:"f_command" db:"f_command"`             // stdio模式下的命令
	Env          string `json:"f_env" db:"f_env"`                     // 环境变量
	Args         string `json:"f_args" db:"f_args"`                   // 命令参数
	Category     string `json:"f_category" db:"f_category"`           // 分类
	Source       string `json:"f_source" db:"f_source"`               // 服务来源
	IsInternal   bool   `json:"f_is_internal" db:"f_is_internal"`     // 是否为内置

	Version     int    `json:"f_version" db:"f_version"`           // 发布版本
	ReleaseDesc string `json:"f_release_desc" db:"f_release_desc"` // 发布描述
	ReleaseUser string `json:"f_release_user" db:"f_release_user"` // 发布者
	ReleaseTime int64  `json:"f_release_time" db:"f_release_time"` // 发布时间
}

// GetBizID 获取业务ID
func (m *MCPServerReleaseDB) GetBizID() string {
	return m.MCPID
}

// DBMCPServerRelease MCP Server发布表数据库操作
type DBMCPServerRelease interface {
	// Insert 插入MCP Server发布
	Insert(ctx context.Context, tx *sql.Tx, release *MCPServerReleaseDB) (err error)
	// UpdateByMCPID 根据mcp_id更新MCP Server发布
	UpdateByMCPID(ctx context.Context, tx *sql.Tx, release *MCPServerReleaseDB) error
	// SelectListPage 分页查询mcp server发布列表
	SelectListPage(ctx context.Context, tx *sql.Tx, filter map[string]interface{},
		sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) (releaseList []*MCPServerReleaseDB, err error)
	// SelectByMCPID 根据mcp_id查询MCP Server发布
	SelectByMCPID(ctx context.Context, tx *sql.Tx, mcpID string) (release *MCPServerReleaseDB, err error)
	// SelectByMCPIDs 根据mcp_id列表查询MCP Server发布
	SelectByMCPIDs(ctx context.Context, tx *sql.Tx, mcpIDs []string, fields []string) (releaseList []*MCPServerReleaseDB, err error)
	// CountByWhereClause 根据条件统计数量
	CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]interface{}) (count int64, err error)
	// DeleteByMCPID 根据mcp_id删除MCP Server发布
	DeleteByMCPID(ctx context.Context, tx *sql.Tx, mcpID string) error
}
