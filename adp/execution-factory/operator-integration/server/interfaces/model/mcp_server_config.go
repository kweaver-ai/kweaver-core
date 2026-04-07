package model

import (
	"context"
	"database/sql"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common/ormhelper"
)

//go:generate mockgen -source=mcp_server_config.go -destination=../../mocks/model_mcp_server_config.go -package=mocks

// MCPServerConfigDB MCP Server配置表对应的结构体
type MCPServerConfigDB struct {
	ID           int64  `json:"f_id" db:"f_id"`                       // id
	MCPID        string `json:"f_mcp_id" db:"f_mcp_id"`               // mcp_id
	CreateUser   string `json:"f_create_user" db:"f_create_user"`     // 创建者
	CreateTime   int64  `json:"f_create_time" db:"f_create_time"`     // 创建时间
	UpdateUser   string `json:"f_update_user" db:"f_update_user"`     // 编辑者
	UpdateTime   int64  `json:"f_update_time" db:"f_update_time"`     // 编辑时间
	CreationType string `json:"f_creation_type" db:"f_creation_type"` // 创建类型
	Version      int    `json:"f_version" db:"f_version"`             // 版本号
	Name         string `json:"f_name" db:"f_name"`                   // MCP Server名称，全局唯一
	Description  string `json:"f_description" db:"f_description"`     // 描述信息
	Mode         string `json:"f_mode" db:"f_mode"`                   // 通信模式（sse、streamable、stdio_npx、stdio_uvx）
	URL          string `json:"f_url" db:"f_url"`                     // 通信地址,SSE/Streamable模式下的服务URL
	Headers      string `json:"f_headers" db:"f_headers"`             // http请求头,JSON字符串
	Command      string `json:"f_command" db:"f_command"`             // stdio模式下的命令
	Env          string `json:"f_env" db:"f_env"`                     // 环境变量
	Args         string `json:"f_args" db:"f_args"`                   // 命令参数
	Status       string `json:"f_status" db:"f_status"`               // 状态
	Category     string `json:"f_category" db:"f_category"`           // 分类
	Source       string `json:"f_source" db:"f_source"`               // 服务来源
	IsInternal   bool   `json:"f_is_internal" db:"f_is_internal"`     // 是否为内置
}

// GetBizID 获取业务ID
func (m *MCPServerConfigDB) GetBizID() string {
	return m.MCPID
}

// DBMCPServerConfig MCP Server配置表数据库操作
type DBMCPServerConfig interface {
	// Insert 插入MCP Server配置
	Insert(ctx context.Context, tx *sql.Tx, config *MCPServerConfigDB) (ID string, err error)
	// UpdateByID 更新MCP Server配置
	UpdateByID(ctx context.Context, tx *sql.Tx, config *MCPServerConfigDB) error
	// UpdateStatus 更新MCP Server配置状态
	UpdateStatus(ctx context.Context, tx *sql.Tx, ID string, status string, updateUser string, version int) error
	// DeleteByID 删除MCP Server配置
	DeleteByID(ctx context.Context, tx *sql.Tx, ID string) error
	// BatchDelete 批量删除MCP Server配置
	BatchDelete(ctx context.Context, tx *sql.Tx, IDs []string) error
	// SelectListPage 分页查询mcp server配置列表
	SelectListPage(ctx context.Context, tx *sql.Tx, filter map[string]interface{},
		sort *ormhelper.SortParams, cursor *ormhelper.CursorParams) (configList []*MCPServerConfigDB, err error)
	// SelectByID 查询MCP Server配置
	SelectByID(ctx context.Context, tx *sql.Tx, ID string) (config *MCPServerConfigDB, err error)
	// SelectByName 查询MCP Server配置
	SelectByName(ctx context.Context, tx *sql.Tx, name string, status []string) (config *MCPServerConfigDB, err error)
	// CountByWhereClause 根据条件统计数量
	CountByWhereClause(ctx context.Context, tx *sql.Tx, filter map[string]interface{}) (count int64, err error)
	// SelectByMCPIDs 查询MCP Server配置列表
	SelectByMCPIDs(ctx context.Context, mcpIDs []string) (configList []*MCPServerConfigDB, err error)
	// SelectListByNamesAndStatus 根据名字及状态批量获取列表
	SelectListByNamesAndStatus(ctx context.Context, names []string, status ...string) (configList []*MCPServerConfigDB, err error)
}
