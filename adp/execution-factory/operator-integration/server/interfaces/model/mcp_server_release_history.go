package model

import (
	"context"
	"database/sql"
)

// MCPServerReleaseHistoryDB MCP Server发布历史表对应的结构体
//
//go:generate mockgen -source=mcp_server_release_history.go -destination=../../mocks/model_mcp_server_release_history.go -package=mocks
type MCPServerReleaseHistoryDB struct {
	ID         int64  `json:"f_id" db:"f_id"`                   // id
	CreateUser string `json:"f_create_user" db:"f_create_user"` // 创建者
	CreateTime int64  `json:"f_create_time" db:"f_create_time"` // 创建时间
	UpdateUser string `json:"f_update_user" db:"f_update_user"` // 编辑者
	UpdateTime int64  `json:"f_update_time" db:"f_update_time"` // 编辑时间

	MCPID       string `json:"f_mcp_id" db:"f_mcp_id"`             // mcp_id
	MCPRelease  string `json:"f_mcp_release" db:"f_mcp_release"`   // mcp server 发布信息
	Version     int    `json:"f_version" db:"f_version"`           // 发布版本
	ReleaseDesc string `json:"f_release_desc" db:"f_release_desc"` // 发布描述
	// FromVersion    string `json:"from_version" db:"from_version"`       // 回滚源版本
	// RollbackReason string `json:"rollback_reason" db:"rollback_reason"` // 回滚原因
}

// DBMCPServerReleaseHistory MCP Server发布历史表数据库操作
type DBMCPServerReleaseHistory interface {
	// InsertMCPServerReleaseHistory 插入MCP Server发布历史
	Insert(ctx context.Context, tx *sql.Tx, history *MCPServerReleaseHistoryDB) (id string, err error)
	// SelectByMCPID 根据mcp_id查询MCP Server发布历史
	SelectByMCPID(ctx context.Context, tx *sql.Tx, mcpID string) (historys []*MCPServerReleaseHistoryDB, err error)
	// DeleteByID 根据id删除MCP Server发布历史
	DeleteByID(ctx context.Context, tx *sql.Tx, id int64) error
	// DeleteByMCPID 根据mcp_id删除MCP Server发布历史
	DeleteByMCPID(ctx context.Context, tx *sql.Tx, mcpID string) error
	// DeleteByMCPIDAndVersion 根据mcp_id和version删除MCP Server发布历史
	DeleteByMCPIDAndVersion(ctx context.Context, tx *sql.Tx, mcpID string, version int) error
}
