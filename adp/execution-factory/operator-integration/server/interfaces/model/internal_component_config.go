package model

import (
	"context"
	"database/sql"
)

// InternalComponentConfigDB 内置组件组件配置表
//
//go:generate mockgen -source=internal_component_config.go -destination=../../mocks/model_internal_component_config.go -package=mocks
type InternalComponentConfigDB struct {
	ID            string `json:"f_id" db:"f_id"`                         // 主键
	ComponentType string `json:"f_component_type" db:"f_component_type"` // 组件类型
	ComponentID   string `json:"f_component_id" db:"f_component_id"`     // 组件ID
	ConfigVersion string `json:"f_config_version" db:"f_config_version"` // 配置版本
	ConfigSource  string `json:"f_config_source" db:"f_config_source"`   // 配置来源(自动/手动)
	ProtectedFlag bool   `json:"f_protected_flag" db:"f_protected_flag"` // 手动配置保护锁(内部)
}

// IInternalComponentConfigDB 内置组件配置接口
type IInternalComponentConfigDB interface {
	InsertConfig(ctx context.Context, tx *sql.Tx, config *InternalComponentConfigDB) error                   // 添加配置
	UpdateConfig(ctx context.Context, tx *sql.Tx, config *InternalComponentConfigDB) error                   // 更新配置
	DeleteConfig(ctx context.Context, tx *sql.Tx, configType, configID string) error                         // 删除配置
	SelectConfig(ctx context.Context, configType, configID string) (bool, *InternalComponentConfigDB, error) // 查询配置
}
