package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source=logics_toolbox.go -destination=../mocks/toolbox.go -package=mocks

// ToolStatusType 工具状态类型
type ToolStatusType string

func (t ToolStatusType) String() string {
	return string(t)
}

const (
	ToolStatusTypeDisabled ToolStatusType = "disabled" // 禁用
	ToolStatusTypeEnabled  ToolStatusType = "enabled"  // 启用
)

// UpsertInternalToolBoxReq 内置工具箱注册请求
type UpsertInternalToolBoxReq struct {
	BoxID    string      `json:"box_id" validate:"required"`                                // 唯一ID,由业务方注册，保证唯一性
	BoxName  string      `json:"box_name" validate:"required"`                              // 工具箱名称
	Desc     string      `json:"box_desc" validate:"required"`                              // 工具箱描述
	Category BizCategory `json:"box_category" form:"box_category" default:"other_category"` // 分类
}

// CreateToolBoxReq 新建工具箱请求
type CreateToolBoxReq struct {
	BusinessDomainID string       `header:"x-business-domain" validate:"required"`                                       // 业务域ID
	UserID           string       `header:"user_id" validate:"required"`                                                 // 用户ID,内部使用
	BoxName          string       `json:"box_name" form:"box_name"`                                                      // 工具箱名称
	BoxDesc          string       `json:"box_desc" form:"box_desc"`                                                      // 工具箱描述
	BoxSvcURL        string       `json:"box_svc_url" form:"box_svc_url"`                                                // 工具箱服务地址
	Category         BizCategory  `json:"box_category" form:"box_category" default:"other_category"`                     // 分类
	MetadataType     MetadataType `json:"metadata_type" form:"metadata_type" validate:"required,oneof=openapi function"` // 元数据类型(强制参数)
	Source           string       `json:"source" form:"source" default:"custom"`                                         // 工具箱来源(默认custom)
	*OpenAPIInput    `json:",inline"`
}

// CreateToolBoxResp 新建工具返回结果
type CreateToolBoxResp struct {
	BoxID string `json:"box_id"` // 工具箱ID
}

// UpdateToolBoxReq 更新工具箱请求
type UpdateToolBoxReq struct {
	UserID        string       `header:"user_id" validate:"required"`                                                 // 用户ID,内部使用
	BoxID         string       `uri:"box_id" validate:"required"`                                                     // 工具箱ID
	BoxName       string       `json:"box_name" form:"box_name" validate:"required"`                                  // 工具箱名称
	BoxDesc       string       `json:"box_desc" form:"box_desc" validate:"required"`                                  // 工具箱描述
	BoxSvcURL     string       `json:"box_svc_url" form:"box_svc_url"`                                                // 工具箱服务地址(当metadata_type为openapi时必填)
	Category      BizCategory  `json:"box_category" form:"box_category" default:"other_category" validate:"required"` // 分类
	MetadataType  MetadataType `json:"metadata_type" form:"metadata_type" validate:"oneof=openapi function"`          // 元数据类型(可选参数)
	*OpenAPIInput `json:",inline"`
}

// UpdateToolBoxResp 更新工具箱返回结果
type UpdateToolBoxResp struct {
	BoxID     string          `json:"box_id"`     // 工具箱ID
	EditTools []*EditToolInfo `json:"edit_tools"` // 工具箱下的工具列表
}

// EditToolInfo 已编辑的工具基本信息
type EditToolInfo struct {
	ToolID string         `json:"tool_id"`     // 工具ID
	Status ToolStatusType `json:"status"`      // 工具状态
	Name   string         `json:"name"`        // 工具名称
	Desc   string         `json:"description"` // 工具描述
}

// ToolBoxToolInfo 工具箱信息
type ToolBoxToolInfo struct {
	MetadataType     MetadataType `json:"metadata_type" validate:"required,oneof=openapi function"` // 元数据类型
	BusinessDomainID string       `json:"business_domain_id"`                                       // 业务域ID
	BoxID            string       `json:"box_id"`                                                   // 工具箱ID
	BoxName          string       `json:"box_name"`                                                 // 工具箱名称
	BoxDesc          string       `json:"box_desc"`                                                 // 工具箱描述
	Status           BizStatus    `json:"status" validate:"oneof=unpublish published offline"`      // 工具箱状态
	BoxSvcURL        string       `json:"box_svc_url"`                                              // 工具箱服务地址
	CategoryType     string       `json:"category_type"`                                            // 分类
	CategoryName     string       `json:"category_name"`                                            // 分类名称
	IsInternal       bool         `json:"is_internal"`                                              // 是否为内部工具箱
	Source           string       `json:"source" default:"custom" validate:"oneof=custom internal"` // 工具箱来源
	Tools            []*ToolInfo  `json:"tools"`                                                    // 工具箱下的工具列表
	CreateTime       int64        `json:"create_time"`                                              // 创建时间
	UpdateTime       int64        `json:"update_time"`                                              // 更新时间
	CreateUser       string       `json:"create_user"`                                              // 创建用户
	UpdateUser       string       `json:"update_user"`                                              // 更新用户
	ReleaseUser      string       `json:"release_user,omitempty"`                                   // 发布人
	ReleaseTime      int64        `json:"release_time,omitempty"`                                   // 发布时间
}

// ToolInfo 工具信息
type ToolInfo struct {
	ToolID           string                 `json:"tool_id"`                                                                    // 工具ID
	Name             string                 `json:"name"`                                                                       // 工具名称
	Description      string                 `json:"description"`                                                                // 工具描述
	Status           ToolStatusType         `json:"status" default:"disabled" validate:"oneof=disabled enabled"`                // 工具状态
	MetadataType     MetadataType           `json:"metadata_type" default:"openapi" validate:"required,oneof=openapi function"` // 元数据类型
	Metadata         *MetadataInfo          `json:"metadata"`                                                                   // 元数据
	UseRule          string                 `json:"use_rule"`                                                                   // 使用规则
	GlobalParameters *ParametersStruct      `json:"global_parameters"`                                                          // 全局参数
	CreateTime       int64                  `json:"create_time"`                                                                // 创建时间
	UpdateTime       int64                  `json:"update_time"`                                                                // 更新时间
	CreateUser       string                 `json:"create_user"`                                                                // 创建用户
	UpdateUser       string                 `json:"update_user"`                                                                // 更新用户
	ExtendInfo       map[string]interface{} `json:"extend_info"`                                                                // 扩展信息
	// 资源类型
	ResourceObject ResourceObjectType `json:"resource_object"` // 资源类型
}

// GetToolBoxReq 获取工具箱请求
type GetToolBoxReq struct {
	UserID   string `header:"user_id"`                 // 用户ID,内部使用
	BoxID    string `uri:"box_id" validate:"required"` // 工具箱ID
	IsPublic bool   `header:"is_public"`               // 是否公开接口
}

// DeleteBoxReq 删除工具箱请求
type DeleteBoxReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"` // 业务域ID
	UserID           string `header:"user_id" validate:"required"`           // 用户ID,内部使用
	BoxID            string `uri:"box_id" validate:"required"`
}

// DeleteBoxResp 删除工具箱返回结果
type DeleteBoxResp struct {
	BoxID string `json:"box_id"` // 工具箱ID
}

// QueryToolBoxListReq 获取工具箱列表请求
type QueryToolBoxListReq struct {
	BusinessDomainID string      `header:"x-business-domain" validate:"required"`                       // 业务域ID
	UserID           string      `header:"user_id"`                                                     // 用户ID,内部使用
	IsPublic         bool        `header:"is_public"`                                                   // 是否公开接口
	CreateUser       string      `form:"create_user"`                                                   // 创建人
	ReleaseUser      string      `form:"release_user"`                                                  // 发布人
	BoxCategory      BizCategory `form:"category"`                                                      // 分类
	Status           BizStatus   `form:"status" validate:"omitempty,oneof=unpublish published offline"` // 工具箱状态
	BoxName          string      `form:"name"`                                                          // 工具箱名称
	CommonPageParams
}

// QueryMarketToolBoxListReq 查询市场工具箱列表请求
type QueryMarketToolBoxListReq struct {
	BusinessDomainID string      `header:"x-business-domain" validate:"required"` // 业务域ID
	UserID           string      `header:"user_id"`                               // 用户ID,内部使用
	IsPublic         bool        `header:"is_public"`                             // 是否公开接口
	CreateUser       string      `form:"create_user"`                             // 创建人
	ReleaseUser      string      `form:"release_user"`                            // 发布人
	BoxCategory      BizCategory `form:"category"`                                // 分类
	BoxName          string      `form:"name"`                                    // 工具箱名称
	CommonPageParams
}

// CommonPageParams 通用分页参数
type CommonPageParams struct {
	Page      int    `form:"page" default:"1" validate:"min=1"`                                           // 页码，从1开始
	PageSize  int    `form:"page_size" default:"10" validate:"min=1,max=100"`                             // 每页大小
	All       bool   `form:"all"`                                                                         // 是否查询所有工具箱
	SortBy    string `form:"sort_by" default:"update_time" validate:"oneof=create_time update_time name"` // 排序字段，默认为创建时间
	SortOrder string `form:"sort_order" default:"desc" validate:"oneof=asc desc"`                         // 排序顺序，默认为降序
}

// ToolBoxInfo 工具箱信息
type ToolBoxInfo struct {
	MetadataType     MetadataType `json:"metadata_type" validate:"required,oneof=openapi function"` // 元数据类型
	BusinessDomainID string       `json:"business_domain_id"`                                       // 业务域ID
	BoxID            string       `json:"box_id"`                                                   // 工具箱ID
	BoxName          string       `json:"box_name"`                                                 // 工具箱名称
	BoxDesc          string       `json:"box_desc"`                                                 // 工具箱描述
	BoxSvcURL        string       `json:"box_svc_url"`                                              // 工具箱服务地址
	Status           BizStatus    `json:"status" validate:"oneof=unpublish published offline"`      // 工具箱状态
	CategoryType     string       `json:"category_type"`                                            // 分类
	CategoryName     string       `json:"category_name"`                                            // 分类名称
	IsInternal       bool         `json:"is_internal"`                                              // 是否为内部工具箱
	Source           string       `json:"source" default:"custom" validate:"oneof=custom internal"` // 工具箱来源
	Tools            []string     `json:"tools"`                                                    // 工具箱下的工具列表
	CreateTime       int64        `json:"create_time"`                                              // 创建时间
	UpdateTime       int64        `json:"update_time"`                                              // 更新时间
	CreateUser       string       `json:"create_user"`                                              // 创建用户
	UpdateUser       string       `json:"update_user"`                                              // 更新用户
	ReleaseUser      string       `json:"release_user,omitempty"`                                   // 发布人
	ReleaseTime      int64        `json:"release_time,omitempty"`                                   // 发布时间
}

// QueryToolBoxListResp 获取工具箱列表返回结果
type QueryToolBoxListResp struct {
	CommonPageResult `json:",inline"`
	Data             []*ToolBoxInfo `json:"data"` // 工具箱列表
}

// ParametersStruct 参数结构体
type ParametersStruct struct {
	Name        string      `json:"name" validate:"required"`                                           // 参数名称
	Description string      `json:"description" validate:"required"`                                    // 参数描述
	Required    bool        `json:"required"`                                                           // 是否必填
	In          string      `json:"in" validate:"required,oneof=query path header cookie body"`         // 参数位置，例如：query, path, header, cookie, body
	Type        string      `json:"type" validate:"required,oneof=string integer boolean array object"` // 参数类型，例如：string, integer, boolean, array, object
	Value       interface{} `json:"value"`                                                              // 参数值
}

// CreateToolReq 创建工具请求
type CreateToolReq struct {
	UserID           string                 `header:"user_id" validate:"required"`                                                 // 用户ID,内部使用
	BoxID            string                 `uri:"box_id" validate:"required"`                                                     // 工具箱ID
	MetadataType     MetadataType           `json:"metadata_type" form:"metadata_type" validate:"required,oneof=openapi function"` // 元数据类型(强制参数)
	UseRule          string                 `json:"use_rule" form:"use_rule"`                                                      // 使用规则
	GlobalParameters *ParametersStruct      `json:"global_parameters" form:"global_parameters" validate:"omitempty"`               // 全局参数
	ExtendInfo       map[string]interface{} `json:"extend_info" form:"extend_info"`                                                // 扩展信息
	FunctionInput    *FunctionInput         `json:"function_input,omitempty"`
	*OpenAPIInput    `json:",inline"`
}

// CreateToolResp 创建工具返回结果
type CreateToolResp struct {
	BoxID        string                    `json:"box_id"`                // 工具箱ID
	SuccessCount int64                     `json:"success_count"`         // 成功数量
	SuccessIDs   []string                  `json:"success_ids,omitempty"` // 成功的工具ID列表
	FailureCount int64                     `json:"failure_count"`         // 失败数量
	Failures     []CreateToolFailureResult `json:"failures,omitempty"`    // 创建失败的工具ID列表及错误信息
}

// CreateToolFailureResult 创建工具失败结果
type CreateToolFailureResult struct {
	ToolName string `json:"tool_name"` // 失败的工具名称
	Error    error  `json:"error_msg"` // 失败原因
}

// UpdateToolReq 更新工具请求
type UpdateToolReq struct {
	UserID            string                 `header:"user_id" validate:"required"` // 用户ID,内部使用
	BoxID             string                 `uri:"box_id" validate:"required"`
	ToolID            string                 `uri:"tool_id" validate:"required"`
	ToolName          string                 `json:"name" form:"name" validate:"required"`
	ToolDesc          string                 `json:"description" form:"description" validate:"required"`
	UseRule           string                 `json:"use_rule" form:"use_rule"`                                                      // 使用规则
	GlobalParameters  *ParametersStruct      `json:"global_parameters" form:"global_parameters"`                                    // 全局参数
	ExtendInfo        map[string]interface{} `json:"extend_info" form:"extend_info"`                                                // 扩展信息
	MetadataType      MetadataType           `json:"metadata_type" form:"metadata_type" validate:"required,oneof=openapi function"` // 元数据类型(可选参数)
	FunctionInputEdit *FunctionInputEdit     `json:"function_input,omitempty"`
	*OpenAPIInput     `json:",inline"`
}

// UpdateToolResp 更新工具返回结果
type UpdateToolResp struct {
	BoxID  string `json:"box_id"`  // 工具箱ID
	ToolID string `json:"tool_id"` // 工具ID
}

// GetToolReq 获取工具请求
type GetToolReq struct {
	UserID string `header:"user_id"` // 用户ID,内部使用
	BoxID  string `uri:"box_id" validate:"required"`
	ToolID string `uri:"tool_id" validate:"required"`
}

// BatchDeleteToolReq 批量删除工具请求
type BatchDeleteToolReq struct {
	UserID  string   `header:"user_id" validate:"required"` // 用户ID,内部使用
	BoxID   string   `uri:"box_id" validate:"required"`
	ToolIDs []string `json:"tool_ids" validate:"required"`
}

type BatchDeleteToolResp struct {
	BoxID  string   `json:"box_id"`   // 工具箱ID
	ToolID []string `json:"tool_ids"` // 工具ID
}

// QueryToolListReq 获取工具列表请求
type QueryToolListReq struct {
	UserID      string         `header:"user_id"` // 用户ID,内部使用
	Page        int            `form:"page" default:"1" validate:"min=1"`
	PageSize    int            `form:"page_size" default:"10" validate:"min=1,max=100"`
	SortBy      string         `form:"sort_by" default:"create_time" validate:"oneof=create_time update_time tool_name"` // 排序字段，默认为创建时间
	SortOrder   string         `form:"sort_order" default:"desc" validate:"oneof=asc desc"`                              // 排序顺序，默认为降序
	ToolName    string         `form:"name"`                                                                             // 工具名称
	Status      ToolStatusType `form:"status" validate:"omitempty,oneof=disabled enabled"`                               // 工具状态                                      // 工具状态
	QueryUserID string         `form:"user_id"`                                                                          // 查询用户ID
	All         bool           `form:"all"`                                                                              // 是否查询所有工具
	BoxID       string         `uri:"box_id" validate:"required"`                                                        // 工具箱ID                                                               // 是否查询所有工具
}

// QueryToolListResp 获取工具列表返回结果
type QueryToolListResp struct {
	CommonPageResult `json:",inline"`
	BoxID            string      `json:"box_id"` // 工具箱ID
	Tools            []*ToolInfo `json:"tools"`  // 工具箱下的工具列表
}

// QueryMarketToolListReq 获取市场工具列表请求
type QueryMarketToolListReq struct {
	UserID    string         `header:"user_id"` // 用户ID,内部使用
	Page      int            `form:"page" default:"1" validate:"min=1"`
	PageSize  int            `form:"page_size" default:"10" validate:"min=1,max=100"`
	SortBy    string         `form:"sort_by" default:"update_time" validate:"oneof=create_time update_time tool_name"` // 排序字段，默认为创建时间
	SortOrder string         `form:"sort_order" default:"desc" validate:"oneof=asc desc"`                              // 排序顺序，默认为降序
	ToolName  string         `form:"tool_name" validate:"required"`                                                    // 工具名称
	Status    ToolStatusType `form:"status" validate:"omitempty,oneof=disabled enabled"`                               // 工具状态
	All       bool           `form:"all"`                                                                              // 是否查询所有工具
}

// QueryMarketToolListResp 获取市场工具列表返回结果
type QueryMarketToolListResp struct {
	CommonPageResult `json:",inline"`
	Data             []*ToolBoxToolInfo `json:"data"` // 工具详情列表
}

type ToolStatus struct {
	ToolID string         `json:"tool_id" validate:"required"`
	Status ToolStatusType `json:"status" validate:"required,oneof=disabled enabled"`
}

// UpdateToolStatusReq 更新工具状态请求
type UpdateToolStatusReq struct {
	UserID         string        `header:"user_id" validate:"required"` // 用户ID,内部使用
	BoxID          string        `uri:"box_id" validate:"required"`
	ToolStatusList []*ToolStatus `json:",inline"`
}

// ExecuteToolReq 执行工具请求
type ExecuteToolReq struct {
	UserID            string `header:"user_id" validate:"required"` // 用户ID,内部使用
	BoxID             string `uri:"box_id" validate:"required"`
	ToolID            string `uri:"tool_id" validate:"required"`
	Timeout           int    `json:"timeout"` // 超时时间，单位秒
	HTTPRequestParams `json:",inline"`
}

// ConvertOperatorToToolReq 算子转换成工具请求
type ConvertOperatorToToolReq struct {
	UserID           string            `header:"user_id" validate:"required"` // 用户ID,内部使用
	OperatorID       string            `json:"operator_id" validate:"required"`
	BoxID            string            `json:"box_id" validate:"required"`
	UseRule          string            `json:"use_rule"`          // 使用规则
	ExtendInfo       map[string]string `json:"extend_info"`       // 扩展信息
	GlobalParameters *ParametersStruct `json:"global_parameters"` // 全局参数
}

// ConvertOperatorToToolResp 算子转换成工具返回结果
type ConvertOperatorToToolResp struct {
	BoxID  string `json:"box_id"`
	ToolID string `json:"tool_id"` // 工具ID
}

// UpdateToolBoxStatusReq 更新工具箱状态请求
type UpdateToolBoxStatusReq struct {
	UserID string    `header:"user_id" validate:"required"` // 用户ID,内部使用
	BoxID  string    `uri:"box_id" validate:"required"`
	Status BizStatus `json:"status" validate:"required,oneof=unpublish published offline"` // 工具箱状态
}

// UpdateToolBoxStatusResp 更新工具箱状态响应
type UpdateToolBoxStatusResp struct {
	BoxID  string    `json:"box_id"`
	Status BizStatus `json:"status"`
}

// GetReleaseToolBoxInfoReq 获取工具箱信息请求
type GetReleaseToolBoxInfoReq struct {
	UserID string `header:"user_id"`                 // 用户ID,内部使用
	BoxIDs string `uri:"box_id" validate:"required"` // 工具箱ID
	Fields string `uri:"fields" validate:"required"` // 字段
}

// GetReleaseToolBoxInfoResp 获取工具箱信息响应
type GetReleaseToolBoxInfoResp struct {
	MetadataType MetadataType `json:"metadata_type" validate:"required,oneof=openapi function"`
	BoxID        string       `json:"box_id" validate:"required"`
	BoxName      string       `json:"box_name,omitempty"`
	BoxDesc      string       `json:"box_desc,omitempty"`
	BoxSvcURL    string       `json:"box_svc_url,omitempty"`
	Status       string       `json:"status,omitempty"`
	Category     BizCategory  `json:"category_type,omitempty"`
	CategoryName string       `json:"category_name,omitempty"`
	IsInternal   *bool        `json:"is_internal,omitempty"`
	Source       string       `json:"source,omitempty"`
	Tools        []*ToolInfo  `json:"tools,omitempty"`
	CreateUser   string       `json:"create_user,omitempty"`
	UpdateUser   string       `json:"update_user,omitempty"`
	ReleaseUser  string       `json:"release_user,omitempty"`
}

// IToolService 工具箱服务接口
type IToolService interface {
	// 工具箱管理
	CreateToolBox(ctx context.Context, req *CreateToolBoxReq) (resp *CreateToolBoxResp, err error)
	UpdateToolBox(ctx context.Context, req *UpdateToolBoxReq) (resp *UpdateToolBoxResp, err error)
	GetToolBox(ctx context.Context, req *GetToolBoxReq, isMarket bool) (resp *ToolBoxToolInfo, err error)
	DeleteBoxByID(ctx context.Context, req *DeleteBoxReq) (resp *DeleteBoxResp, err error)
	QueryToolBoxList(ctx context.Context, req *QueryToolBoxListReq) (resp *QueryToolBoxListResp, err error)
	QueryMarketToolBoxList(ctx context.Context, req *QueryMarketToolBoxListReq) (resp *QueryToolBoxListResp, err error)
	UpdateToolBoxStatus(ctx context.Context, req *UpdateToolBoxStatusReq) (resp *UpdateToolBoxStatusResp, err error)
	// 工具管理
	CreateTool(ctx context.Context, req *CreateToolReq) (resp *CreateToolResp, err error)
	UpdateTool(ctx context.Context, req *UpdateToolReq) (resp *UpdateToolResp, err error)
	GetBoxTool(ctx context.Context, req *GetToolReq) (resp *ToolInfo, err error)
	DeleteBoxTool(ctx context.Context, req *BatchDeleteToolReq) (resp *BatchDeleteToolResp, err error)
	QueryToolList(ctx context.Context, req *QueryToolListReq) (resp *QueryToolListResp, err error)
	UpdateToolStatus(ctx context.Context, req *UpdateToolStatusReq) (resp []*ToolStatus, err error)
	// 工具调试
	DebugTool(ctx context.Context, req *ExecuteToolReq) (resp *HTTPResponse, err error)
	// 工具执行
	ExecuteTool(ctx context.Context, req *ExecuteToolReq) (resp *HTTPResponse, err error)
	// 工具执行（不包含权限校验和审计日志）
	ExecuteToolCore(ctx context.Context, req *ExecuteToolReq) (resp *HTTPResponse, err error)
	// 算子转换成工具
	ConvertOperatorToTool(ctx context.Context, req *ConvertOperatorToToolReq) (resp *ConvertOperatorToToolResp, err error)
	GetReleaseToolBoxInfo(ctx context.Context, req *GetReleaseToolBoxInfoReq) (resp []*GetReleaseToolBoxInfoResp, err error)
	// 内部接口
	CreateInternalToolBox(ctx context.Context, req *CreateInternalToolBoxReq) (resp *CreateInternalToolBoxResp, err error)

	// 市场接口
	GetMarketToolList(ctx context.Context, req *QueryMarketToolListReq) (resp *QueryMarketToolListResp, err error) // 获取所有的工具

	// 导入超出
	// Impex[*ToolBoxImpexData]
	Import(ctx context.Context, tx *sql.Tx, mode ImportType, data *ComponentImpexConfigModel, userID string) (err error)
	Export(ctx context.Context, req *ExportReq) (data *ComponentImpexConfigModel, err error)
	// 事件处理
	ToolBoxEventHandler
}

// ToolBoxEventHandler 事件处理接口
type ToolBoxEventHandler interface {
	HandleOperatorDeleteEvent(ctx context.Context, message []byte) error
}

// CreateInternalToolBoxReq 内部工具箱注册请求
// 仅供内部接口使用
// MetadataType 目前仅支持 openapi
// Data 为 openapi 元数据
// Version 为版本号:由业务方注册,版本号格式为: x.x.x,例如: 1.0.0，如果版本号不变化，认为没有新版本，不更新元数据
// BoxID 为工具箱ID,
// TODO: 内置工具权限验证，可见范围、使用范围、使用规则
type CreateInternalToolBoxReq struct {
	BusinessDomainID string `header:"x-business-domain" validate:"required"`      // 业务域ID
	UserID           string `header:"user_id"`                                    // 用户ID,内部使用
	BoxID            string `json:"box_id" form:"box_id" validate:"required"`     // 工具箱ID
	BoxName          string `json:"box_name" form:"box_name" validate:"required"` // 工具箱名称
	BoxDesc          string `json:"box_desc" form:"box_desc" validate:"required"` // 描述
	Source           string `json:"source" form:"source" validate:"required"`     // 来源
	// 配置信息
	ConfigVersion string           `json:"config_version" form:"config_version" validate:"required"`                 // 配置版本
	ConfigSource  ConfigSourceType `json:"config_source" form:"config_source" validate:"required,oneof=auto manual"` // 配置来源(自动/手动)
	ProtectedFlag bool             `json:"protected_flag" form:"protected_flag"`                                     // 手动配置保护锁(内部)
	IsPublic      bool             // 判断是否是外部接口
	// 数据源
	MetadataType  MetadataType `json:"metadata_type" form:"metadata_type" validate:"required,oneof=openapi function"` // 元数据类型
	*OpenAPIInput `json:",inline"`
	Functions     []*FunctionInput `json:"functions,omitempty"` // 函数列表
}

// CreateInternalToolBoxResp 创建内部工具箱返回
// 当注册资源冲突时，根据版本等信息判断是否升级。
// 当工具箱名称冲突时，返回409, 详情里面返回 “CreateInternalToolBoxResp” 信息
type CreateInternalToolBoxResp struct {
	BoxID   string      `json:"box_id"`   // 工具箱ID
	BoxName string      `json:"box_name"` // 工具箱名称
	Tools   []*ToolInfo `json:"tools"`    // 工具ID列表
}
