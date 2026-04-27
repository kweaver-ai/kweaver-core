// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

type contextKey string // 自定义专属的key类型

const (
	CONTENT_TYPE_NAME = "Content-Type"
	CONTENT_TYPE_JSON = "application/json"

	HTTP_HEADER_METHOD_OVERRIDE = "x-http-method-override"
	HTTP_HEADER_ACCOUNT_ID      = "x-account-id"
	HTTP_HEADER_ACCOUNT_TYPE    = "x-account-type"
	HTTP_HEADER_BUSINESS_DOMAIN = "x-business-domain"

	ACCOUNT_INFO_KEY    contextKey = "x-account-info"    // 避免直接使用string
	BUSINESS_DOMAIN_KEY contextKey = "x-business-domain" // 业务域ID

	SERVICE_NAME = "ontology-query"

	// 默认的分支
	MAIN_BRANCH = "main"

	// 指标数据查询是否包含模型信息
	DEFAULT_INCLUDE_TYPE_INFO    = "false"
	DEFAULT_INCLUDE_LOGIC_PARAMS = "false"

	// 参数来源
	VALUE_FROM_INPUT    = "input"
	VALUE_FROM_PROPERTY = "property"

	// 属性类型
	PROPERTY_TYPE_METRIC   = "metric"
	PROPERTY_TYPE_OPERATOR = "operator"

	// 排序方向
	DESC_DIRECTION = "desc"
	ASC_DIRECTION  = "asc"

	// 得分排序字段
	SORT_FIELD_SCORE = "_score"

	// 最大路径数(探索子图时，探索的最大路径数)
	MAX_PATHS = 2000

	// 默认的路径查询数量 - 基于路径查询时
	DEFAULT_PATHS = 2000

	// 路径子图查询
	QUERY_TYPE_RELATION_TYPE_PATH = "relation_path"

	// limit的最大值
	MAX_LIMIT = 10000

	// 对象类的对象实例检索时，默认的查询数量
	DEFAULT_OBJECT_LIMIT = 10

	// 子图查询时，节点对象类的对象实例的默认查询数量
	DEFAULT_LIMIT = 1000

	// 算子的执行模式 execution_mode
	OPERATOR_EXECUTION_MODE_SYNC = "sync"

	// 边的方向
	DIRECTION_FORWARD       = "forward"
	DIRECTION_BACKWARD      = "backward"
	DIRECTION_BIDIRECTIONAL = "bidirectional"

	// search after默认的limit大小
	SearchAfter_Limit = 10000

	// 对象类的对象数据查询使用缓存（持久化）数据的默认值
	DEFAULT_IGNORING_STORE_CACHE = "false"
)

const (
	OBJECT_TYPE_KN            = "knowledge_network"
	OBJECT_TYPE_OBJECT_TYPE   = "object_type"
	OBJECT_TYPE_RELATION_TYPE = "relation_type"
	OBJECT_TYPE_ACTION_TYPE   = "action_type"
)

type ResourceInfo struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type CommonInfo struct {
	Tags    []string `json:"tags" mapstructure:"tags"`
	Comment string   `json:"comment" mapstructure:"comment"`
	Icon    string   `json:"icon" mapstructure:"icon"`
	Color   string   `json:"color" mapstructure:"color"`
	Detail  string   `json:"detail" mapstructure:"detail"`
}

type AccountInfo struct {
	ID   string `json:"id" mapstructure:"id"`
	Type string `json:"type" mapstructure:"type"`
	Name string `json:"name" mapstructure:"name"`
}

type PageQuery struct {
	// 分页信息
	NeedTotal bool `json:"need_total"`
	Limit     int  `json:"limit"`
	// UseSearchAfter bool          `json:"use_search_after"` // 业务知识网络只提供search after的方式，不需要提供这个参数
	Sort []*SortParams `json:"sort"`
	SearchAfterParams
}
