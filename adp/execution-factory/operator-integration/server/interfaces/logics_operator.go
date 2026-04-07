// Package interfaces 定义接口
// @file operator.go
// @description: 定义算子操作接口
package interfaces

//go:generate mockgen -source=logics_operator.go -destination=../mocks/operator.go -package=mocks
import (
	"context"
	"database/sql"
)

// OperatorStatusItem 单个状态更新项的结构体
type OperatorStatusItem struct {
	OperatorID string    `json:"operator_id" validate:"required,uuid4"`
	Status     BizStatus `json:"status" validate:"required,oneof=unpublish published offline editing"`
}

// OperatorStatusUpdateReq 状态更新请求
type OperatorStatusUpdateReq struct {
	UserID      string                `header:"user_id" validate:"required"`
	StatusItems []*OperatorStatusItem `json:",inline"`
}

// OperatorDeleteItem 单个删除请求
type OperatorDeleteItem struct {
	OperatorID string `json:"operator_id" validate:"required,uuid4"`
}

// OperatorDeleteReq 删除请求
type OperatorDeleteReq []OperatorDeleteItem

// OperatorUpdateReq 更新请求
type OperatorUpdateReq struct {
	*OperatorRegisterReq `json:",inline"`
	OperatorID           string `json:"operator_id" form:"operator_id" validate:"required,uuid4"`
}

// OperatorRegisterReq 注册请求
type OperatorRegisterReq struct {
	MetadataType           MetadataType            `json:"operator_metadata_type" form:"operator_metadata_type" validate:"required" oneof:"openapi function"` // 算子元数据类型(强制参数)
	OperatorInfo           *OperatorInfo           `json:"operator_info" form:"operator_info"`                                                                // 算子信息
	OperatorExecuteControl *OperatorExecuteControl `json:"operator_execute_control" form:"operator_execute_control"`                                          // 控制参数
	ExtendInfo             map[string]interface{}  `json:"extend_info,omitempty" form:"extend_info,omitempty"`                                                // 拓展信息
	UserToken              string                  `json:"user_token" form:"user_token"`                                                                      // 内部接口传参
	DirectPublish          bool                    `json:"direct_publish,omitempty" form:"direct_publish,omitempty"`                                          // 直接发布
	FunctionInput          *FunctionInput          `json:"function_input,omitempty" form:"function_input,omitempty"`                                          // 函数输入参数
	Data                   string                  `json:"data" form:"data"`                                                                                  // 算子元数据，当算子元数据类型为openapi时必填
}

// OperatorRegisterResp 单个算子注册结果
type OperatorRegisterResp struct {
	Status     ResultStatus `json:"status"`          // 算子注册状态 (failed/success)
	OperatorID string       `json:"operator_id"`     // 算子ID
	Version    string       `json:"version"`         // 算子版本
	Error      error        `json:"error,omitempty"` // 错误信息(需要支持国际化)
}

// OperatorEditReq 编辑请求
type OperatorEditReq struct {
	UserID                 string                  `header:"user_id" validate:"required"` // 用户ID
	Name                   string                  `json:"name" form:"name"`
	Description            string                  `json:"description" form:"description"`                                       // 算子描述
	OperatorID             string                  `json:"operator_id" form:"operator_id" validate:"required,uuid4"`             // 算子ID
	OperatorInfoEdit       *OperatorInfoEdit       `json:"operator_info" form:"operator_info"`                                   // 算子信息
	OperatorExecuteControl *OperatorExecuteControl `json:"operator_execute_control" form:"operator_execute_control"`             // 执行控制
	ExtendInfo             map[string]interface{}  `json:"extend_info,omitempty" form:"extend_info,omitempty"`                   //	 扩展信息
	MetadataType           MetadataType            `json:"metadata_type" form:"metadata_type" validate:"oneof=openapi function"` // 元数据类型(可选参数)
	FunctionInputEdit      *FunctionInputEdit      `json:"function_input,omitempty" form:"function_input,omitempty"`             // 函数输入参数
	*OpenAPIInput          `json:",inline"`
}

// OperatorInfoEdit 算子信息编辑
type OperatorInfoEdit struct {
	Type          OperatorType  `json:"operator_type" default:"basic" validate:"oneof=basic composite"` // 算子类型(basic/composite)
	ExecutionMode ExecutionMode `json:"execution_mode" default:"sync"  validate:"oneof=sync async"`     // 执行模式(async/sync)
	Category      BizCategory   `json:"category" default:"other_category"`                              // 算子分类(data_process/control)
	Source        string        `json:"source" default:"unknown"`                                       // 算子来源(system/unknown)
	IsDataSource  *bool         `json:"is_data_source" form:"is_data_source" default:"false"`           // 是否为数据源算子
}

// OperatorEditResp 编辑响应
type OperatorEditResp struct {
	OperatorID string    `json:"operator_id" validate:"required,uuid4"`
	Version    string    `json:"version" validate:"required,uuid4"`
	Status     BizStatus `json:"status" validate:"required,oneof=unpublish published offline editing"` // validate:"oneof=asc desc"
}

// OperatorDataInfo 算子数据
type OperatorDataInfo struct {
	BusinessDomainID       string                  `json:"business_domain_id"` // 业务域ID
	Name                   string                  `json:"name"`               // 算子名称
	OperatorID             string                  `json:"operator_id" validate:"uuid4"`
	Version                string                  `json:"version" validate:"uuid4"`
	Status                 BizStatus               `json:"status" validate:"omitempty,oneof=unpublish published offline editing"` // 状态
	MetadataType           MetadataType            `json:"metadata_type" default:"openapi" validate:"oneof=openapi function"`     // 算子元数据类型(强制参数)
	Metadata               *MetadataInfo           `json:"metadata"`
	ExtendInfo             map[string]interface{}  `json:"extend_info,omitempty"`
	OperatorInfo           *OperatorInfo           `json:"operator_info"` // 算子信息
	OperatorExecuteControl *OperatorExecuteControl `json:"operator_execute_control"`
	CreateUser             string                  `json:"create_user"`
	CreateTime             int64                   `json:"create_time"`
	UpdateUser             string                  `json:"update_user"`
	UpdateTime             int64                   `json:"update_time"`
	ReleaseUser            string                  `json:"release_user,omitempty"` // 发布人
	ReleaseTime            int64                   `json:"release_time,omitempty"` // 发布时间
	Tag                    int                     `json:"tag,omitempty"`          // 版本号
	IsInternal             bool                    `json:"is_internal"`            // 是否内部算子
}

// OperatorType 算子类型
type OperatorType string

const (
	OperatorTypeBase      OperatorType = "basic"     // 基础算子
	OperatorTypeComposite OperatorType = "composite" // 组合算子
)

// OperatorExecuteControl 算子执行控制
type OperatorExecuteControl struct {
	Timeout     int64               `json:"timeout" form:"timeout" default:"3000"` // 超时时间
	RetryPolicy OperatorRetryPolicy `json:"retry_policy" form:"retry_policy"`      // 重试策略
}

// OperatorRetryPolicy 算子重试策略
type OperatorRetryPolicy struct {
	MaxAttempts     int64           `json:"max_attempts" form:"max_attempts" default:"3"`      // 最大重试次数
	InitialDelay    int64           `json:"initial_delay" form:"initial_delay" default:"1000"` // 初始延迟时间（毫秒）
	BackoffFactor   int64           `json:"backoff_factor" form:"backoff_factor" default:"2"`  // 指数退避因子
	MaxDelay        int64           `json:"max_delay" form:"max_delay" default:"6000"`         // 最大延迟时间（毫秒）
	RetryConditions RetryConditions `json:"retry_conditions" form:"retry_conditions"`          // 重试条件
}

// RetryConditions 重试条件
type RetryConditions struct {
	StatusCode []int64  `json:"status_code" form:"status_code"` // 状态码
	ErrorCodes []string `json:"error_codes" form:"error_codes"` // 业务错误码
}

// OperatorInfo 算子信息
type OperatorInfo struct {
	Type          OperatorType  `json:"operator_type" form:"operator_type" default:"basic" validate:"oneof=basic composite"` // 算子类型(basic/composite)
	ExecutionMode ExecutionMode `json:"execution_mode" form:"execution_mode"  default:"sync"  validate:"oneof=sync async"`   // 执行模式(async/sync)
	Category      BizCategory   `json:"category" form:"category" default:"other_category"`                                   // 算子分类(data_process/control)
	CategoryName  string        `json:"category_name,omitempty" form:"category_name,omitempty"`                              // 算子分类名称(支持国际化)
	Source        string        `json:"source" form:"source" default:"unknown"`                                              // 算子来源(system/unknown)
	IsDataSource  *bool         `json:"is_data_source" form:"is_data_source" default:"false"`                                // 是否为数据源算子
}

// PageQueryRequest 分页查询请求
type PageQueryRequest struct {
	BusinessDomainID string       `header:"x-business-domain" validate:"required"` // 业务域ID
	UserID           string       `header:"user_id"`
	Page             int          `form:"page" default:"1" validate:"min=1"`
	PageSize         int          `form:"page_size" default:"10" validate:"max=100"`
	SortBy           string       `form:"sort_by" default:"update_time" validate:"oneof=update_time create_time name"`
	SortOrder        string       `form:"sort_order" default:"desc" validate:"oneof=asc desc"`
	Name             string       `form:"name"`
	Status           BizStatus    `form:"status" validate:"omitempty,oneof=unpublish published offline editing"`
	CreateUser       string       `form:"create_user"`                                              // 创建人
	Category         BizCategory  `form:"category"`                                                 // 分类
	OperatorType     OperatorType `form:"operator_type" validate:"omitempty,oneof=basic composite"` // 算子类型(basic/composite)
	All              bool         `form:"all"`
	IsDataSource     *bool        `form:"is_data_source"` // 是否为数据源算子
}

// PageQueryResponse 分页查询响应
type PageQueryResponse struct {
	CommonPageResult `json:",inline"`
	Data             []*OperatorDataInfo `json:"data"` // 数据列表
}

// OperatorHistoryDetailReq 查询操作历史详情请求
type OperatorHistoryDetailReq struct {
	UserID     string `header:"user_id"` // 非必填
	OperatorID string `uri:"operator_id" validate:"required"`
	Version    string `uri:"version" validate:"required"`
	Tag        int    `form:"tag"`
}

// OperatorMarketDetailReq 算子市场详情查询请求
type OperatorMarketDetailReq struct {
	UserID     string `header:"user_id"` // 非必填
	OperatorID string `uri:"operator_id" validate:"required"`
}

// DebugOperatorReq 调试请求
type DebugOperatorReq struct {
	UserID            string `header:"user_id" validate:"required"` // 用户ID,内部使用
	OperatorID        string `json:"operator_id" validate:"required,uuid4"`
	Version           string `json:"version" validate:"required,uuid4"`
	Timeout           int    `json:"timeout"` // 超时时间，单位秒
	HTTPRequestParams `json:",inline"`
}

// ExecuteOperatorReq 执行请求
type ExecuteOperatorReq struct {
	UserID            string `header:"user_id" validate:"required"`        // 用户ID,内部使用
	OperatorID        string `uri:"operator_id" validate:"required,uuid4"` // 算子ID
	Timeout           int    `json:"timeout"`                              // 超时时间，单位秒
	HTTPRequestParams `json:",inline"`
}

// OperatorHistoryListReq 获取历史版本列表
type OperatorHistoryListReq struct {
	UserID     string `header:"user_id"` // 非必填
	OperatorID string `uri:"operator_id" validate:"required"`
}

// PageQueryOperatorMarketReq 算子市场查询请求
type PageQueryOperatorMarketReq struct {
	BusinessDomainID string        `header:"x-business-domain" validate:"required"` // 业务域ID
	UserID           string        `header:"user_id"`                               // 非必填
	Page             int           `form:"page" default:"1" validate:"min=1"`
	PageSize         int           `form:"page_size" default:"10" validate:"max=100"`
	SortBy           string        `form:"sort_by" default:"update_time" validate:"oneof=update_time create_time name"`
	SortOrder        string        `form:"sort_order" default:"desc" validate:"oneof=asc desc"`
	All              bool          `form:"all"`
	Status           BizStatus     `form:"status" validate:"omitempty,oneof=published offline"`       // 状态
	Name             string        `form:"name"`                                                      // 算子名称
	CreateUser       string        `form:"create_user"`                                               // 创建人
	ReleaseUser      string        `form:"release_user"`                                              // 发布人
	Category         BizCategory   `form:"category"`                                                  // 分类
	OperatorType     OperatorType  `form:"operator_type" validate:"omitempty,oneof=basic composite"`  // 算子类型(basic/composite)
	IsDataSource     *bool         `form:"is_data_source"`                                            // 是否为数据源算子
	ExecutionMode    ExecutionMode `form:"execution_mode" validate:"omitempty,oneof=sync async"`      // 执行模式(async/sync)
	MetadataType     MetadataType  `form:"metadata_type" validate:"omitempty,oneof=openapi function"` // 元数据类型(openapi/function)
}

// GetOperatorInfoByOperatorIDReq 获取算子信息请求
type GetOperatorInfoByOperatorIDReq struct {
	UserID     string `header:"user_id"` // 非必填
	OperatorID string `uri:"operator_id" validate:"required,uuid4"`
}

// OperatorManager 算子管理接口
type OperatorManager interface {
	RegisterOperatorByOpenAPI(ctx context.Context, req *OperatorRegisterReq, userID string) ([]*OperatorRegisterResp, error)
	// UpdateOperatorStatus 更新算子状态
	UpdateOperatorStatus(ctx context.Context, req *OperatorStatusUpdateReq, userID string) error
	// GetOperatorInfoByOperatorID 获取算子信息
	GetOperatorInfoByOperatorID(ctx context.Context, req *GetOperatorInfoByOperatorIDReq) (*OperatorDataInfo, error)
	GetOperatorQueryPage(ctx context.Context, req *PageQueryRequest) (*PageQueryResponse, error)
	// EditOperator 编辑算子
	EditOperator(ctx context.Context, req *OperatorEditReq) (*OperatorEditResp, error)
	// DeleteOperator 删除算子
	DeleteOperator(ctx context.Context, req OperatorDeleteReq, userID string) error
	UpdateOperatorByOpenAPI(ctx context.Context, req *OperatorUpdateReq, userID string) (resultList []*OperatorRegisterResp, err error)
	// 调试接口
	DebugOperator(ctx context.Context, req *DebugOperatorReq) (resp *HTTPResponse, err error)
	// 执行算子
	ExecuteOperator(ctx context.Context, req *ExecuteOperatorReq) (resp *HTTPResponse, err error)
	// 更具ID，version 获取已经发布版本算子信息
	QueryOperatorHistoryDetail(ctx context.Context, req *OperatorHistoryDetailReq) (*OperatorDataInfo, error)
	QueryOperatorHistoryList(ctx context.Context, req *OperatorHistoryListReq) ([]*OperatorDataInfo, error)
	// QueryOperatorMarketList 算子市场查询
	QueryOperatorMarketList(ctx context.Context, req *PageQueryOperatorMarketReq) (*PageQueryResponse, error)
	// QueryOperatorMarketDetail 算子市场详情查询
	QueryOperatorMarketDetail(ctx context.Context, req *OperatorMarketDetailReq) (*OperatorDataInfo, error)
	// 注册内置算子
	RegisterInternalOperator(ctx context.Context, req *RegisterInternalOperatorReq) (resp *OperatorRegisterResp, err error)
	/*导入导出*/
	// Impex[*OperatorImpexData]
	Export(ctx context.Context, req *ExportReq) (data *ComponentImpexConfigModel, err error)
	Import(ctx context.Context, tx *sql.Tx, mode ImportType, data *OperatorImpexConfig, userID string) (err error)
	// 内部操作接口
	InternalOperatorManager
}

// CheckAddAsToolResp 检查算子是否允许添加为工具响应
type CheckAddAsToolResp struct {
	OperatorID string      `json:"operator_id"`
	Name       string      `json:"name"`
	Metadata   IMetadataDB `json:"metadata"`
}

// InternalOperatorManager 内部操作接口
type InternalOperatorManager interface {
	// 检查是否允许添加为工具
	CheckAddAsTool(ctx context.Context, operatorID, userID string) (resp *CheckAddAsToolResp, err error)
}

// RegisterInternalOperatorReq 注册内置算子请求
type RegisterInternalOperatorReq struct {
	BusinessDomainID       string                  `header:"x-business-domain" validate:"required"`                                             // 业务域ID
	UserID                 string                  `header:"user_id"`                                                                           // 用户ID
	OperatorID             string                  `json:"operator_id" form:"operator_id" validate:"required,uuid4"`                            // 算子ID
	MetadataType           MetadataType            `json:"metadata_type" form:"metadata_type" validate:"required" oneof:"openapi function"`     // 算子元数据类型(强制参数)
	Name                   string                  `json:"name" form:"name" validate:"required"`                                                // 算子数据
	OperatorType           OperatorType            `json:"operator_type" form:"operator_type" default:"basic" validate:"oneof=basic composite"` // 算子类型(basic/composite)
	ExecutionMode          ExecutionMode           `json:"execution_mode" form:"execution_mode" default:"sync"  validate:"oneof=sync async"`    // 执行模式(async/sync)
	Source                 string                  `json:"source" form:"source" default:"internal" validate:"required"`                         // 算子来源
	OperatorExecuteControl *OperatorExecuteControl `json:"operator_execute_control" form:"operator_execute_control"`                            // 控制参数
	ExtendInfo             map[string]interface{}  `json:"extend_info,omitempty" form:"extend_info,omitempty"`                                  // 拓展信息
	ConfigSource           ConfigSourceType        `json:"config_source" form:"config_source" validate:"required,oneof=auto manual"`            // 配置来源(自动/手动)
	ConfigVersion          string                  `json:"config_version" form:"config_version" validate:"required"`                            // 配置版本
	ProtectedFlag          bool                    `json:"protected_flag" form:"protected_flag"`                                                // 版本周期内保护标志，true表示保护，false表示不保护，这个字段主要作用于内置MCPServer
	IsDataSource           *bool                   `json:"is_data_source" form:"is_data_source"`                                                // 是否为数据源算子
	*OpenAPIInput          `json:",inline"`
	Functions              []*FunctionInput `json:"functions,omitempty"` // 函数列表
	IsPublic               bool             // 判断是否是外部接口
}
