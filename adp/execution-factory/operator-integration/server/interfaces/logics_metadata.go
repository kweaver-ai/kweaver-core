package interfaces

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/getkin/kin-openapi/openapi3"
	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
)

// MetadataInfo 元数据信息
type MetadataInfo struct {
	Version         string           `json:"version" validate:"required"`     // 版本
	Summary         string           `json:"summary" validate:"required"`     // 摘要
	Description     string           `json:"description"`                     // 描述
	ServerURL       string           `json:"server_url" validate:"required"`  // 服务URL
	Path            string           `json:"path" validate:"required"`        // 路径
	Method          string           `json:"method" validate:"required"`      // 方法
	CreateTime      int64            `json:"create_time" validate:"required"` // 创建时间
	UpdateTime      int64            `json:"update_time" validate:"required"` // 更新时间
	CreateUser      string           `json:"create_user" validate:"required"` // 创建人
	UpdateUser      string           `json:"update_user" validate:"required"` // 更新人
	APISpec         *APISpec         `json:"api_spec" validate:"required"`    // OpenAPI 格式
	FunctionContent *FunctionContent `json:"function_content,omitempty"`      // 函数内容
}

/*OpenAPI 格式定义*/

// OpenAPIContent OpenAPI内容
type OpenAPIContent struct {
	SererURL string `json:"server_url" validate:"required"` // 服务器URL
	// Info 信息
	// @description: 信息
	Info *openapi3.Info `json:"info"`
	// PathItems 路径项内容
	// @description: 路径项内容
	PathItems []*PathItemContent `json:"path_items"`
}

// GetPathItemByMethodAndPath 获取路径项内容
// @description: 根据方法和路径获取路径项内容
// @param method 方法
// @param path 路径
// @return []*PathItemContent 路径项内容
func (o *OpenAPIContent) GetPathItemByMethodAndPath(method, path string) *PathItemContent {
	// 获取指定路径项
	for _, item := range o.PathItems {
		if item.Path != path || item.Method != method {
			continue
		}

		return item
	}
	return nil
}

// PathItemContent 路径项内容
type PathItemContent struct {
	Summary     string   `json:"summary" validate:"required"`
	Path        string   `json:"path" validate:"required"`
	Method      string   `json:"method" validate:"required"`
	Description string   `json:"description"`
	APISpec     *APISpec `json:"api_spec"`
	ServerURL   string   `json:"server_url" validate:"required"` // 服务器URL
	ErrMessage  string   `json:"err_message,omitempty"`
}

// APISpec OpenAPI 格式
type APISpec struct {
	Parameters   []*Parameter `json:"parameters"`    // 结构化参数
	RequestBody  *RequestBody `json:"request_body"`  // 请求体结构
	Responses    []*Response  `json:"responses"`     // 响应结构
	Components   *Components  `json:"components"`    // 组件定义
	Callbacks    interface{}  `json:"callbacks"`     // 回调函数定义
	Security     interface{}  `json:"security"`      // 安全要求
	Tags         []string     `json:"tags"`          // 标签
	ExternalDocs interface{}  `json:"external_docs"` // 外部文档
}

// ToJSON 将APISpec转换为JSON字符串
func (a *APISpec) ToJSON() string {
	jsonBytes, _ := jsoniter.Marshal(a)
	return string(jsonBytes)
}

// Components 组件定义
type Components struct {
	Schemas interface{} `json:"schemas"` // 引用的结构体定义
}

// Parameter 参数类型
type Parameter struct {
	Name        string              `json:"name"`
	In          string              `json:"in"` // path/query/header/cookie
	Description string              `json:"description"`
	Required    bool                `json:"required"`
	Schema      *openapi3.SchemaRef `json:"schema,omitempty"`
	Example     any                 `json:"example,omitempty"`
	Examples    openapi3.Examples   `json:"examples,omitempty"`
	Content     openapi3.Content    `json:"content,omitempty"`
}

// RequestBody 请求体结构
type RequestBody struct {
	Description string           `json:"description"`
	Content     openapi3.Content `json:"content"` // 按媒体类型分类
	Required    bool             `json:"required"`
}

// Response 响应结构
type Response struct {
	StatusCode  string           `json:"status_code"` // 200/400等
	Description string           `json:"description"`
	Content     openapi3.Content `json:"content"`
}

/*函数相关参数定义*/

// ScriptType 脚本类型
type ScriptType string

const (
	ScriptTypePython ScriptType = "python" // Python 脚本类型
)

// FunctionContent 函数内容定义
type FunctionContent struct {
	ScriptType      ScriptType       `json:"script_type" form:"script_type" default:"python" validate:"required,oneof=python"` // 脚本类型
	Code            string           `json:"code" form:"code" validate:"required"`                                             // Python 代码（必填）
	Dependencies    []DependencyInfo `json:"dependencies" form:"dependencies"`                                                 // 依赖库列表
	DependenciesURL string           `json:"dependencies_url" form:"dependencies_url"`                                         // 依赖库安装源地址
}

// ParameterType 参数类型
type ParameterType string

const (
	ParameterTypeString  ParameterType = "string"  // 字符串类型
	ParameterTypeNumber  ParameterType = "number"  // 数字类型
	ParameterTypeBoolean ParameterType = "boolean" // 布尔类型
	ParameterTypeArray   ParameterType = "array"   // 数组类型
	ParameterTypeObject  ParameterType = "object"  // 对象类型
)

// ParameterDef 参数定义
// 支持通过 SubParameters 字段实现多层嵌套,适用于 Object 和 Array 类型
type ParameterDef struct {
	Name        string `json:"name"`                  // 参数名称
	Description string `json:"description,omitempty"` // 参数描述
	Required    bool   `json:"required"`              // 是否必填

	// 参数类型: string, number, boolean, array, object
	Type ParameterType `json:"type,omitempty" validate:"omitempty,oneof=string number boolean array object"` // 参数类型

	// 简单类型的约束字段
	Default any   `json:"default,omitempty"` // 默认值
	Enum    []any `json:"enum,omitempty"`    // 枚举值(可选)
	Example any   `json:"example,omitempty"` // 示例值

	// 复杂类型的嵌套定义
	// 使用场景:
	//   - Object 类型: SubParameters 定义对象的属性列表
	//   - Array 类型: SubParameters 只包含一个元素,定义数组元素的结构
	//     (数组元素名称建议使用 "items")
	SubParameters []*ParameterDef `json:"sub_parameters,omitempty"` // 子参数列表(用于 object 和 array 类型)
}

// FunctionInput  函数输入定义
type FunctionInput struct {
	// 基础信息
	Name        string `json:"name" form:"name"`                         // 函数名称
	Description string `json:"description,omitempty" form:"description"` // 函数描述，用于说明函数的功能和行为
	// 参数定义
	Inputs  []*ParameterDef `json:"inputs,omitempty" form:"inputs"`   // 输入参数列表
	Outputs []*ParameterDef `json:"outputs,omitempty" form:"outputs"` // 输出参数列表
	// 代码相关
	ScriptType      ScriptType        `json:"script_type" form:"script_type" default:"python" validate:"required,oneof=python"` // 脚本类型
	Code            string            `json:"code" form:"code"`                                                                 // Python 代码（必填）
	Dependencies    []*DependencyInfo `json:"dependencies,omitempty" form:"dependencies"`                                       // 依赖库列表
	DependenciesURL string            `json:"dependencies_url,omitempty" form:"dependencies_url"`                               // 依赖库安装源地址
}

// FunctionInputEdit 函数输入编辑定义
type FunctionInputEdit struct {
	// 参数定义
	Inputs  []*ParameterDef `json:"inputs,omitempty" form:"inputs"`   // 输入参数列表
	Outputs []*ParameterDef `json:"outputs,omitempty" form:"outputs"` // 输出参数列表
	// 代码相关
	ScriptType      ScriptType        `json:"script_type" form:"script_type" default:"python" validate:"required,oneof=python"` // 脚本类型
	Code            string            `json:"code" form:"code"`                                                                 // Python 代码（必填）
	Dependencies    []*DependencyInfo `json:"dependencies,omitempty" form:"dependencies"`                                       // 依赖库列表
	DependenciesURL string            `json:"dependencies_url,omitempty" form:"dependencies_url"`                               // 依赖库安装源地址
}

// OpenAPIInput OpenAPI 输入定义
type OpenAPIInput struct {
	// 基础信息
	Data json.RawMessage `json:"data" form:"data"` // 原始内容（OpenAPI JSON/YAML）
}

// IMetadataService 统一元数据管理接口
type IMetadataService interface {
	// 注册元数据
	RegisterMetadata(ctx context.Context, tx *sql.Tx, metadata IMetadataDB) (version string, err error)
	// 批量注册元数据
	BatchRegisterMetadata(ctx context.Context, tx *sql.Tx, metadatas []IMetadataDB) (versions []string, err error)
	// 根据版本查询元数据
	GetMetadataByVersion(ctx context.Context, metadataType MetadataType, version string) (IMetadataDB, error)
	// 批量查询元数据
	BatchGetMetadata(ctx context.Context, apiVersions, funcVersions []string) ([]IMetadataDB, error)
	// 更新元数据
	UpdateMetadata(ctx context.Context, tx *sql.Tx, metadata IMetadataDB) error
	// 删除元数据
	// DeleteMetadata(ctx context.Context, tx *sql.Tx, metadataType MetadataType, version string) error
	// 批量删除元数据
	BatchDeleteMetadata(ctx context.Context, tx *sql.Tx, metadataType MetadataType, versions []string) error
	// 验证元数据格式
	ValidateMetadata(ctx context.Context, metadata IMetadataDB) error
	// 元数据解析
	ParseMetadata(ctx context.Context, metadataType MetadataType, input any) ([]IMetadataDB, error)
	// 获取解析后的原始内容
	ParseRawContent(ctx context.Context, metadataType MetadataType, input any) (content any, err error)
	// 根据SourceID、SourceType查询元数据
	GetMetadataBySource(ctx context.Context, sourceID string, sourceType model.SourceType) (bool, IMetadataDB, error)
	// 批量根据SourceID、SourceType查询元数据
	BatchGetMetadataBySourceIDs(ctx context.Context, sourceMap map[model.SourceType][]string) (sourceIDToMetadataMap map[string]IMetadataDB, err error)
	// 检查并返回元数据是否存在
	CheckMetadataExists(ctx context.Context, metadataType MetadataType, version string) (bool, IMetadataDB, error)
}
