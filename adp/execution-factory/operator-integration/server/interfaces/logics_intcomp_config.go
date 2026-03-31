package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source=logics_intcomp_config.go -destination=../mocks/intcomp_config.go -package=mocks

// ComponentType 组件类型
type ComponentType string

const (
	// ComponentTypeToolBox 工具箱组件
	ComponentTypeToolBox ComponentType = "toolbox"
	// ComponentTypeMCP MCP组件
	ComponentTypeMCP ComponentType = "mcp"
	// ComponentTypeOperator 算子组件
	ComponentTypeOperator ComponentType = "operator"
)

func (c ComponentType) String() string {
	return string(c)
}

type ConfigSourceType string

const (
	// ConfigSourceTypeAuto 自动配置
	ConfigSourceTypeAuto ConfigSourceType = "auto"
	// ConfigSourceTypeManual 手动配置
	ConfigSourceTypeManual ConfigSourceType = "manual"
)

func (c ConfigSourceType) String() string {
	return string(c)
}

// IntCompConfig 内置组件配置
type IntCompConfig struct {
	ComponentID   string           `json:"component_id" validate:"required"`
	ComponentType ComponentType    `json:"component_type" validate:"required"`
	ConfigVersion string           `json:"config_version" validate:"required"`                  // 配置版本
	ConfigSource  ConfigSourceType `json:"config_source" validate:"required,oneof=auto manual"` // 配置来源(自动/手动)
	ProtectedFlag bool             `json:"protected_flag"`                                      // 手动配置保护锁(内部)
}

// IntCompConfigAction 内置组件配置操作
type IntCompConfigAction string

const (
	// IntCompConfigActionTypeCreate 创建
	IntCompConfigActionTypeCreate IntCompConfigAction = "create"
	// IntCompConfigActionTypeUpdate 修改
	IntCompConfigActionTypeUpdate IntCompConfigAction = "update"
	// IntCompConfigActionTypeSkip 跳过
	IntCompConfigActionTypeSkip IntCompConfigAction = "skip"
)

// IIntCompConfigService 内置组件配置服务
type IIntCompConfigService interface {
	// CompareConfig 比较当前配置和待检查的配置，返回结果
	CompareConfig(ctx context.Context, check *IntCompConfig) (action IntCompConfigAction, err error)
	UpdateConfig(ctx context.Context, tx *sql.Tx, config *IntCompConfig) (err error)
	DeleteConfig(ctx context.Context, tx *sql.Tx, configType, configID string) error
}
