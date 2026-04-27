// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

const (
	ConnectorModeLocal  string = "local"  // 内置在 vega-backend 进程内运行
	ConnectorModeRemote string = "remote" // 作为独立服务运行，通过 HTTP 调用
)

const (
	ConnectorCategoryTable   string = "table"   // 关系型数据库
	ConnectorCategoryIndex   string = "index"   // 搜索引擎
	ConnectorCategoryTopic   string = "topic"   // 消息队列
	ConnectorCategoryFile    string = "file"    // 文件
	ConnectorCategoryFileset string = "fileset" // 文件集
	ConnectorCategoryMetric  string = "metric"  // 时序数据库
	ConnectorCategoryAPI     string = "api"     // API 服务
)

var (
	CONNECTOR_TYPE_SORT = map[string]string{
		"name": "f_name",
	}
)

// 连接器类型常量定义
const (
	ConnectorTypeMySQL      string = "mysql"
	ConnectorTypeMariaDB    string = "mariadb"
	ConnectorTypePostgreSQL string = "postgresql"
	ConnectorTypeOpenSearch string = "opensearch"
	ConnectorTypeOracle     string = "oracle"
	ConnectorTypeAnyShare   string = "anyshare"
)

// 当前统一查询接口支持的连接器类型列表
// 注意：系统支持更多连接器类型，但当前统一查询接口仅支持以下类型
var SupportedConnectorTypesForQuery = map[string]bool{
	ConnectorTypeMySQL:      true,
	ConnectorTypeMariaDB:    true,
	ConnectorTypePostgreSQL: true,
	ConnectorTypeOpenSearch: true,
}

// GetSupportedConnectorTypesForQuery 返回当前统一查询接口支持的连接器类型列表
// 注意：系统支持更多连接器类型，但当前统一查询接口仅支持以下类型
func GetSupportedConnectorTypesForQuery() []string {
	return []string{
		ConnectorTypeMySQL,
		ConnectorTypeMariaDB,
		ConnectorTypePostgreSQL,
		ConnectorTypeOpenSearch,
	}
}

// IsConnectorTypeSupportedForQuery 检查给定的连接器类型是否被当前统一查询接口支持
// 注意：系统支持更多连接器类型，但当前统一查询接口仅支持部分类型
func IsConnectorTypeSupportedForQuery(connectorType string) bool {
	return SupportedConnectorTypesForQuery[connectorType]
}

// ConnectorFieldConfig 定义连接器配置字段的元数据（兼容 JSON Schema properties）
type ConnectorFieldConfig struct {
	Name        string `json:"name"`        // 字段显示名
	Type        string `json:"type"`        // 字段类型: string, integer, number, boolean, object, array
	Description string `json:"description"` // 字段描述
	Required    bool   `json:"required"`    // 是否必填
	Encrypted   bool   `json:"encrypted"`   // 是否需要加密存储（自定义扩展）
}

// ConnectorType 表示一个已注册的 connector 类型
type ConnectorType struct {
	Type        string                          `json:"type"`
	Name        string                          `json:"name"`         // mysql, postgresql, kafka...
	Tags        []string                        `json:"tags"`         // 标签
	Description string                          `json:"description"`  // 类型描述
	Mode        string                          `json:"mode"`         // local | remote
	Category    string                          `json:"category"`     // table | index | topic | file | fileset | metric | api
	Endpoint    string                          `json:"endpoint"`     // 仅 remote 模式，远程服务地址
	FieldConfig map[string]ConnectorFieldConfig `json:"field_config"` // 字段配置（兼容 JSON Schema properties）
	Enabled     bool                            `json:"enabled"`      // 是否启用

	Operations []string `json:"operations"`
}

// ConnectorTypesQueryParams 查询参数
type ConnectorTypesQueryParams struct {
	PaginationQueryParams
	Tag      string `json:"tag"`      // 按标签筛选
	Mode     string `json:"mode"`     // 按模式筛选
	Category string `json:"category"` // 按分类筛选
	Enabled  *bool  `json:"enabled"`  // 按启用状态筛选
}

// ConnectorTypeReq 表示创建/更新 connector 类型的请求
type ConnectorTypeReq struct {
	Type        string                          `json:"type"`
	Name        string                          `json:"name"`         // mysql, postgresql, kafka...
	Tags        []string                        `json:"tags"`         // 标签
	Description string                          `json:"description"`  // 类型描述
	Mode        string                          `json:"mode"`         // local | remote
	Category    string                          `json:"category"`     // table | index | topic | file | fileset | metric | api
	Endpoint    string                          `json:"endpoint"`     // 仅 remote 模式，远程服务地址
	FieldConfig map[string]ConnectorFieldConfig `json:"field_config"` // 字段配置（兼容 JSON Schema properties）
	Enabled     bool                            `json:"enabled"`      // 是否启用

	IfNameModify        bool           `json:"-"`
	OriginConnectorType *ConnectorType `json:"-"`
}
