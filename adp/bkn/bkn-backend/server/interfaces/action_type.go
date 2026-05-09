// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	cond "bkn-backend/common/condition"
)

const (
	// 行动资源类型
	ACTION_SOURCE_TYPE_TOOL = "tool"
	ACTION_SOURCE_TYPE_MCP  = "mcp"

	// 行动类型
	ACTION_TYPE_ADD    = "add"
	ACTION_TYPE_MODIFY = "modify"
	ACTION_TYPE_DELETE = "delete"
)

var (
	ACTION_TYPE_SORT = map[string]string{
		"name":        "f_name",
		"update_time": "f_update_time",
	}

	// 行动类型
	ActionTypeMap = map[string]bool{
		ACTION_TYPE_ADD:    true,
		ACTION_TYPE_MODIFY: true,
		ACTION_TYPE_DELETE: true,
	}

	// 行动条件操作符
	ActionCondOperationMap = map[string]struct{}{
		cond.OperationAnd:        {},
		cond.OperationOr:         {},
		cond.OperationEq:         {},
		cond.OperationNotEq:      {},
		cond.OperationGt:         {},
		cond.OperationGte:        {},
		cond.OperationLt:         {},
		cond.OperationLte:        {},
		cond.OperationIn:         {},
		cond.OperationNotIn:      {},
		cond.OperationEmpty:      {},
		cond.OperationNotEmpty:   {},
		cond.OperationTrue:       {},
		cond.OperationFalse:      {},
		cond.OperationRange:      {},
		cond.OperationOutRange:   {},
		cond.OperationBefore:     {},
		cond.OperationBetween:    {},
		cond.OperationExist:      {},
		cond.OperationNotExist:   {},
		cond.OperationLike:       {},
		cond.OperationNotLike:    {},
		cond.OperationPrefix:     {},
		cond.OperationNotPrefix:  {},
		cond.OperationNull:       {},
		cond.OperationNotNull:    {},
		cond.OperationRegex:      {},
		cond.OperationContain:    {},
		cond.OperationNotContain: {},
		cond.OperationCurrent:    {},
	}
)

type ActionTypeWithKeyField struct {
	ATID         string           `json:"id" mapstructure:"id"`
	ATName       string           `json:"name" mapstructure:"name"`
	ActionType   string           `json:"action_type" mapstructure:"action_type"`
	ObjectTypeID string           `json:"object_type_id" mapstructure:"object_type_id"`
	ObjectType   SimpleObjectType `json:"object_type,omitempty" mapstructure:"object_type"` // 翻译绑定的对象类
	Condition    *ActionCondCfg   `json:"cond,omitempty" mapstructure:"cond"`
	Affect       *ActionAffect    `json:"affect" mapstructure:"affect"`
	ActionSource ActionSource     `json:"action_source" mapstructure:"action_source"`
	Parameters   []Parameter      `json:"parameters" mapstructure:"parameters"`
	Schedule     Schedule         `json:"schedule" mapstructure:"schedule"`
}

// knowledge_network
type ActionType struct {
	ActionTypeWithKeyField `mapstructure:",squash"`
	CommonInfo             `mapstructure:",squash"`
	KNID                   string `json:"kn_id" mapstructure:"kn_id"`
	Branch                 string `json:"branch" mapstructure:"branch"`

	Creator    AccountInfo `json:"creator" mapstructure:"creator"`
	CreateTime int64       `json:"create_time" mapstructure:"create_time"`
	Updater    AccountInfo `json:"updater" mapstructure:"updater"`
	UpdateTime int64       `json:"update_time" mapstructure:"update_time"`

	ModuleType string `json:"module_type" mapstructure:"module_type"`

	IfNameModify bool `json:"-"`
	// 向量
	Vector []float32 `json:"_vector,omitempty"`
	Score  *float64  `json:"_score,omitempty"` // opensearch检索的得分，在概念搜索时使用
}

type ActionSource struct {
	Type string `json:"type" mapstructure:"type"`
	// 互斥字段，根据Type选择
	// type 为 tool
	BoxID  string `json:"box_id,omitempty" mapstructure:"box_id"`
	ToolID string `json:"tool_id,omitempty" mapstructure:"tool_id"`
	// type 为 mcp
	McpID    string `json:"mcp_id,omitempty" mapstructure:"mcp_id"`
	ToolName string `json:"tool_name,omitempty" mapstructure:"tool_name"`
}

type ActionCondCfg struct {
	ObjectTypeID string `json:"object_type_id,omitempty" mapstructure:"object_type_id"` // 行动条件需要标记是哪个行动类的

	Field            string           `json:"field,omitempty" mapstructure:"field"`
	Operation        string           `json:"operation,omitempty" mapstructure:"operation"`
	SubConds         []*ActionCondCfg `json:"sub_conditions,omitempty" mapstructure:"sub_conditions"`
	cond.ValueOptCfg `mapstructure:",squash"`

	RemainCfg map[string]any `json:",omitempty" mapstructure:",remain,squash"`

	NameField *ViewField `json:"-" mapstructure:"-"`
}

type ActionAffect struct {
	ObjectTypeID string           `json:"object_type_id,omitempty" mapstructure:"object_type_id"` // 翻译影响的对象类
	ObjectType   SimpleObjectType `json:"object_type,omitempty" mapstructure:"object_type"`
	Comment      string           `json:"comment,omitempty" mapstructure:"comment"`
}

type Schedule struct {
	Type       string `json:"type" mapstructure:"type"`
	Expression string `json:"expression" mapstructure:"expression"`
}

// 对象类的分页查询
type ActionTypesQueryParams struct {
	PaginationQueryParameters
	NamePattern   string
	Tag           string
	Branch        string
	KNID          string
	ObjectTypeIDs []string
	ActionType    string
}

// 检索行动类列表
type ActionTypes struct {
	Entries     []*ActionType `json:"entries"`
	TotalCount  int64         `json:"total_count,omitempty"`
	SearchAfter []any         `json:"search_after,omitempty"`
	OverallMs   int64         `json:"overall_ms"`
}

func IsValidActionSourceType(m string) bool {
	return m == ACTION_SOURCE_TYPE_TOOL || m == ACTION_SOURCE_TYPE_MCP
}
