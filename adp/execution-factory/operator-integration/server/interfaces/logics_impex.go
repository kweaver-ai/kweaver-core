package interfaces

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
)

//go:generate mockgen -source=logics_impex.go -destination=../mocks/logics_impex.go -package=mocks

// ImportType 导入类型
type ImportType string

const (
	ImportTypeUpsert ImportType = "upsert" // 更新或创建
	ImportTypeCreate ImportType = "create" // 仅创建
)

// IComponentImpexConfig 组件导入导出配置接口
type IComponentImpexConfig interface {
	ExportConfig(ctx context.Context, exportReq *ExportConfigReq) (config *ComponentImpexConfigModel, err error)
	ImportConfig(ctx context.Context, importReq *ImportConfigReq) (err error)
}

// ExportConfigReq 导出配置请求
type ExportConfigReq struct {
	UserID string        `header:"user_id" validate:"required"`                      // 用户ID
	Type   ComponentType `uri:"type" validate:"required,oneof=operator toolbox mcp"` // 组件类型
	ID     string        `uri:"id" validate:"required"`                              // 组件ID
}

// ImportConfigReq 导入配置请求
type ImportConfigReq struct {
	BusinessDomainID string          `header:"x-business-domain" validate:"required"`              // 业务域ID
	UserID           string          `header:"user_id" validate:"required"`                        // 用户ID
	Type             ComponentType   `uri:"type" validate:"required,oneof=operator toolbox mcp"`   // 组件类型
	Mode             ImportType      `form:"mode" default:"create" validate:"oneof=create upsert"` // 配置导入类型
	Data             json.RawMessage `form:"data" validate:"required"`
}

// ComponentImpexConfigModel 组件导入导出配置模型
type ComponentImpexConfigModel struct {
	Operator *OperatorImpexConfig `json:"operator,omitempty"`
	Toolbox  *ToolBoxImpexConfig  `json:"toolbox,omitempty"`
	MCP      *MCPImpexConfig      `json:"mcp,omitempty"`
}

// OperatorImpexConfig 算子导入导出配置
type OperatorImpexConfig struct {
	Configs          []*OperatorImpexItem `json:"configs"`           // 算子配置
	CompositeConfigs []any                `json:"composite_configs"` // 组合算子依赖配置
}

// ToolBoxImpexConfig 工具箱导入导出配置
type ToolBoxImpexConfig struct {
	Configs []*ToolBoxImpexItem `json:"configs"`
}

// MCPImpexConfig MCP导入导出配置
type MCPImpexConfig struct {
	Configs []*MCPServersImpexItem `json:"configs"`
}

// Impex 导入导出接口
type Impex[T any] interface {
	Import(context.Context, *sql.Tx, *ImportReq[T]) error
	Export(context.Context, *ExportReq) (T, error)
}

// ExportReq 导出请求
type ExportReq struct {
	UserID string   `header:"user_id" validate:"required"`
	IDs    []string `json:"ids" validate:"required,min=1"` // 校验长度
}

// ImportReq 导入请求
type ImportReq[T any] struct {
	UserID string     `header:"user_id" validate:"required"`
	Mode   ImportType `json:"mode" validate:"required,oneof=upsert create"`
	Data   T          `json:"data" validate:"required"`
}

type ImportResp struct {
	// 成功数
	SuccessCount int `json:"success_count"`
	// 失败数
	FailedCount int `json:"failed_count"`
	// 失败详情
	FailedDetails []*ImportFailedDetail `json:"failed_details,omitempty"`
}

type ImportFailedDetail struct {
	Type ComponentType `json:"type"`
	// 失败对象信息
	ID   string `json:"id"`
	Name string `json:"name"`
	// 错误信息
	Error error `json:"error,omitempty"`
}

// // ParseData 解析表单中的JSON数据到泛型类型
// func (r *ImportReq[T]) ParseData(data T) error {
// 	// var data T
// 	if r.Data == nil {
// 		return fmt.Errorf("import data is nil")
// 	}
// 	if err := json.Unmarshal([]byte(r.Data), data); err != nil {
// 		err = fmt.Errorf("data unmarshal err: %w", err)
// 		return err
// 	}
// 	return nil
// }

type ToolBoxImpexData struct {
	ToolBoxes []*ToolBoxImpexItem `json:"tool_boxes"`
}

// ToolBoxImpexItem 工具箱导入导出数据模型
type ToolBoxImpexItem struct {
	BoxID        string           `json:"box_id" validate:"required"`                                                 // 工具箱ID
	BoxName      string           `json:"box_name" validate:"required"`                                               // 工具箱名称
	BoxDesc      string           `json:"box_desc"`                                                                   // 工具箱描述
	BoxSvcURL    string           `json:"box_svc_url" validate:"required,url"`                                        // 工具箱服务地址
	Status       BizStatus        `json:"status" validate:"oneof=unpublish published offline"`                        // 工具箱状态
	CategoryType string           `json:"category_type" validate:"required"`                                          // 分类
	CategoryName string           `json:"category_name"`                                                              // 分类名称
	IsInternal   bool             `json:"is_internal"`                                                                // 是否为内部工具箱
	Source       string           `json:"source" default:"custom" validate:"oneof=custom internal"`                   // 工具箱来源
	Tools        []*ToolImpexItem `json:"tools"`                                                                      // 工具箱下的工具列表
	CreateTime   int64            `json:"create_time"`                                                                // 创建时间
	UpdateTime   int64            `json:"update_time"`                                                                // 更新时间
	CreateUser   string           `json:"create_user"`                                                                // 创建用户
	UpdateUser   string           `json:"update_user"`                                                                // 更新用户
	MetadataType MetadataType     `json:"metadata_type" default:"openapi" validate:"required,oneof=openapi function"` // 元数据类型
}

// ToolImpexItem 工具导入导出数据模型
type ToolImpexItem struct {
	ToolInfo        `json:",inline"`
	SourceID        string           `json:"source_id"`
	SourceType      model.SourceType `json:"source_type"`
	FunctionContent `json:",inline"` // 当metadata_type=="function"时函数内容定义
}
type MCPImpexData struct {
	MCPServers      []*MCPServersImpexItem `json:"mcp_servers"`
	DeployToolBoxes []*ToolBoxImpexItem    `json:"deploy_tool_boxes,omitempty"` // 依赖工具
}

// MCPServersImpexItem MCP Server 导出数据
type MCPServersImpexItem struct {
	MCPCoreConfigInfo `json:",inline"`
	MCPID             string          `json:"mcp_id" validate:"required"`                                           // MCP Server ID
	Version           int             `json:"version,omitempty"`                                                    // MCP Server版本
	CreationType      MCPCreationType `json:"creation_type" validate:"required"`                                    // 创建类型
	Name              string          `json:"name" validate:"required"`                                             // MCP Server名称
	Description       string          `json:"description" validate:"required"`                                      // 描述信息
	Status            BizStatus       `json:"status" validate:"required,oneof=unpublish editing published offline"` // 状态
	Source            string          `json:"source" validate:"required"`                                           // 来源
	IsInternal        bool            `json:"is_internal"`                                                          // 是否为内置
	Category          string          `json:"category,omitempty" default:"other_category"`                          // 分类
	CreateUser        string          `json:"create_user,omitempty"`                                                // 创建用户
	CreateTime        int64           `json:"create_time,omitempty"`                                                // 创建时间
	UpdateUser        string          `json:"update_user,omitempty"`                                                // 更新用户
	UpdateTime        int64           `json:"update_time,omitempty"`                                                // 更新时间
	MCPTools          []*MCPToolItem  `json:"mcp_tools,omitempty"`                                                  // mcp tool 配置
}

// MCPToolItem MCP工具配置
type MCPToolItem struct {
	MCPToolID   string `json:"mcp_tool_id" validate:"required"` // mcp tool id
	MCPID       string `json:"mcp_id" validate:"required"`      // mcp id
	MCPVersion  int    `json:"mcp_version"`
	BoxID       string `json:"box_id" validate:"required"` // box id
	BoxName     string `json:"box_name" validate:"required"`
	ToolID      string `json:"tool_id" validate:"required"` // tool id
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	UseRule     string `json:"use_rule"`
}

type OperatorImpexData struct {
	Operators []*OperatorImpexItem `json:"operators"`
}

// OperatorImpexItem 算子导入导出数据模型
type OperatorImpexItem struct {
	OperatorID             string                  `json:"operator_id" validate:"uuid4"`                                          // 算子ID
	OperatorName           string                  `json:"operator_name" validate:"required"`                                     // 算子名称
	Version                string                  `json:"version" validate:"uuid4"`                                              // 算子版本
	Status                 BizStatus               `json:"status" validate:"omitempty,oneof=unpublish published offline editing"` // 状态
	MetadataType           MetadataType            `json:"metadata_type" default:"openapi" validate:"oneof=openapi function"`     // 算子元数据类型(强制参数)
	Metadata               *MetadataInfo           `json:"metadata" validate:"required"`                                          // 算子元数据
	ExtendInfo             map[string]interface{}  `json:"extend_info,omitempty"`                                                 // 扩展信息
	OperatorInfo           *OperatorInfo           `json:"operator_info"`                                                         // 算子信息
	OperatorExecuteControl *OperatorExecuteControl `json:"operator_execute_control"`                                              // 算子执行控制
	CreateUser             string                  `json:"create_user"`                                                           // 创建用户
	CreateTime             int64                   `json:"create_time"`                                                           // 创建时间
	UpdateUser             string                  `json:"update_user"`                                                           // 更新用户
	UpdateTime             int64                   `json:"update_time"`                                                           // 更新时间
	IsInternal             bool                    `json:"is_internal"`                                                           // 是否内部算子
}
