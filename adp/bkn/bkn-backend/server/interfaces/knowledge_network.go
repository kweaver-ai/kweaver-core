// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	cond "bkn-backend/common/condition"
	"bkn-backend/interfaces/data_type"
)

const (
	// 指标数据查询是否包含模型信息
	DEFAULT_INCLUDE_TYPE_INFO = "false"

	// 路径方向
	DIRECTION_FORWARD       = "forward"
	DIRECTION_BACKWARD      = "backward"
	DIRECTION_BIDIRECTIONAL = "bidirectional"
)

var (
	KN_SORT = map[string]string{
		"name":        "f_name",
		"update_time": "f_update_time",
	}

	// 字段集为 kn_id, module_type, id, name, property_name, property_display_name, comment
	CONCPET_QUERY_FIELD_STR = []string{
		"kn_id",
		"module_type",
		"id",
		"name",
		"comment",
		"detail",
		"data_properties.name",
		"data_properties.display_name",
		"data_properties.comment",
		"logic_properties.name",
		"logic_properties.display_name",
		"logic_properties.comment",
	}
	CONCPET_QUERY_FIELD = map[string]*cond.FieldCfg{
		"kn_id": {
			Name: "kn_id",
			Type: data_type.DATATYPE_KEYWORD,
		},
		"module_type": {
			Name: "module_type",
			Type: data_type.DATATYPE_KEYWORD,
		},
		"id": {
			Name: "id",
			Type: data_type.DATATYPE_KEYWORD,
		},
		"name": {
			Name: "name",
			Type: data_type.DATATYPE_TEXT,
		},
		"comment": {
			Name: "comment",
			Type: data_type.DATATYPE_TEXT,
		},
		"detail": {
			Name: "detail",
			Type: data_type.DATATYPE_TEXT,
		},
		"data_properties.name": {
			Name: "data_properties.name",
			Type: data_type.DATATYPE_TEXT,
		},
		"data_properties.display_name": {
			Name: "data_properties.display_name",
			Type: data_type.DATATYPE_TEXT,
		},
		"data_properties.comment": {
			Name: "data_properties.comment",
			Type: data_type.DATATYPE_TEXT,
		},
		"logic_properties.name": {
			Name: "data_properties.name",
			Type: data_type.DATATYPE_TEXT,
		},
		"logic_properties.display_name": {
			Name: "data_properties.display_name",
			Type: data_type.DATATYPE_TEXT,
		},
		"logic_properties.comment": {
			Name: "data_properties.comment",
			Type: data_type.DATATYPE_TEXT,
		},
	}

	DIRECTION_MAP = map[string]bool{
		DIRECTION_FORWARD:       true,
		DIRECTION_BACKWARD:      true,
		DIRECTION_BIDIRECTIONAL: true,
	}
)

// knowledge_network
type KN struct {
	KNID       string `json:"id" mapstructure:"id"`
	KNName     string `json:"name" mapstructure:"name"`
	CommonInfo `mapstructure:",squash"`

	SkillContent string `json:"skill_content,omitempty" mapstructure:"skill_content"`

	Branch         string `json:"branch,omitempty" mapstructure:"branch"`
	BusinessDomain string `json:"business_domain,omitempty" mapstructure:"business_domain"`

	ConceptGroups []*ConceptGroup     `json:"concept_groups,omitempty" mapstructure:"concept_groups"`
	ObjectTypes   []*ObjectType       `json:"object_types,omitempty" mapstructure:"object_types"`
	RelationTypes []*RelationType     `json:"relation_types,omitempty" mapstructure:"relation_types"`
	ActionTypes   []*ActionType       `json:"action_types,omitempty" mapstructure:"action_types"`
	RiskTypes     []*RiskType         `json:"risk_types,omitempty" mapstructure:"risk_types"`
	Metrics       []*MetricDefinition `json:"metrics,omitempty" mapstructure:"metrics"`

	Creator    AccountInfo `json:"creator" mapstructure:"creator"`
	CreateTime int64       `json:"create_time" mapstructure:"create_time"`
	Updater    AccountInfo `json:"updater" mapstructure:"updater"`
	UpdateTime int64       `json:"update_time" mapstructure:"update_time"`

	ModuleType string `json:"module_type" mapstructure:"module_type"`

	IfNameModify bool `json:"-"`

	// 统计信息
	Statistics *Statistics `json:"statistics,omitempty"`
	// 操作权限
	Operations []string `json:"operations,omitempty"`

	// 向量
	Vector []float32 `json:"_vector,omitempty"`
	Score  *float64  `json:"_score,omitempty"` // opensearch检索的得分，在概念搜索时使用
}

type Statistics struct {
	CgTotal       int `json:"concept_groups_total"`
	OtTotal       int `json:"object_types_total"`
	RtTotal       int `json:"relation_types_total"`
	AtTotal       int `json:"action_types_total"`
	RiskTypeTotal int `json:"risk_types_total"`
}

// 业务知识网络的分页查询
type KNsQueryParams struct {
	PaginationQueryParameters
	NamePattern    string
	Tag            string
	BusinessDomain string
	Branch         string
}

// 概念搜索
type ConceptsQuery struct {
	ConceptGroups []string       `json:"concept_groups,omitempty"`
	Condition     map[string]any `json:"condition,omitempty"`
	// 分页信息
	NeedTotal bool `json:"need_total"`
	Limit     int  `json:"limit"`
	// UseSearchAfter bool          `json:"use_search_after"` // 业务知识网络只提供search after的方式，不需要提供这个参数
	Sort []*SortParams `json:"sort"`
	SearchAfterParams

	KNID            string        `json:"-"`
	Branch          string        `json:"-"`
	ModuleType      string        `json:"-"`
	ActualCondition *cond.CondCfg `json:"-"`
}

type SortParams struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

type SearchAfterParams struct {
	SearchAfter []any `json:"search_after"`
	// PitID        string `json:"pit_id"`
	// PitKeepAlive string `json:"pit_keep_alive"`
}

// 基于起点、方向和路径长度获取对象子图的请求体
type RelationTypePathsBaseOnSource struct {
	ConceptGroups     []string `json:"concept_groups,omitempty"`
	SourceObjecTypeId string   `json:"source_object_type_id"`
	Direction         string   `json:"direction"`
	PathLength        int      `json:"path_length"`

	KNID   string `json:"-"`
	Branch string `json:"-"`
}

type RelationTypePath struct {
	ObjectTypes []ObjectTypeWithKeyField `json:"object_types"`
	TypeEdges   []TypeEdge               `json:"relation_types"`
	Length      int                      `json:"length"`
}

type TypeEdge struct {
	RelationTypeId      string                   `json:"relation_type_id"`
	RelationType        RelationTypeWithKeyField `json:"relation_type"`
	SourceObjectTypeId  string                   `json:"source_object_type_id"`
	Target_ObjectTypeId string                   `json:"target_object_type_id"`
	Direction           string                   `json:"direction"`
}

type CommonQueryParameters struct {
	IncludeStatistics bool
}

// BatchIDIndex holds concept IDs declared in the current request body (KN / concept group tree, etc.)
// for preflight dependency resolution within the same scope as a creation transaction.
type BatchIDIndex struct {
	KNID   string
	Branch string

	ObjectTypes     map[string]*ObjectType
	RelationTypeIDs map[string]struct{}
	ActionTypeIDs   map[string]struct{}
	ConceptGroupIDs map[string]struct{}
	// Metrics maps metric id -> declaration from the KN payload for duplicate/conflict checks and strict cross-validation against batch.ObjectTypes (same batch OT semantics as relation/action validators).
	Metrics map[string]*MetricDefinition
}
