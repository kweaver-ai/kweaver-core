// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

var (
	CONCPET_GROUP_SORT = map[string]string{
		"name":        "f_name",
		"update_time": "f_update_time",
	}
)

// concept_group
type ConceptGroup struct {
	CGID   string `json:"id" mapstructure:"id"`
	CGName string `json:"name" mapstructure:"name"`

	CommonInfo `mapstructure:",squash"`
	KNID       string `json:"kn_id" mapstructure:"kn_id"`
	Branch     string `json:"branch" mapstructure:"branch"`

	ObjectTypes   []*ObjectType   `json:"object_types,omitempty" mapstructure:"object_types"`
	RelationTypes []*RelationType `json:"relation_types,omitempty" mapstructure:"relation_types"`
	ActionTypes   []*ActionType   `json:"action_types,omitempty" mapstructure:"action_types"`

	Creator    AccountInfo `json:"creator,omitempty" mapstructure:"creator"`
	CreateTime int64       `json:"create_time,omitempty" mapstructure:"create_time"`
	Updater    AccountInfo `json:"updater,omitempty" mapstructure:"updater"`
	UpdateTime int64       `json:"update_time,omitempty" mapstructure:"update_time"`

	ModuleType string `json:"module_type" mapstructure:"module_type"`

	// 统计信息
	Statistics *Statistics `json:"statistics,omitempty"`
	// 操作权限
	Operations []string `json:"operations,omitempty"`

	// 向量
	Vector []float32 `json:"_vector,omitempty"`
	Score  *float64  `json:"_score,omitempty"` // opensearch检索的得分，在概念搜索时使用

	IfNameModify bool `json:"-"`
}

// concept_group_relation
type ConceptGroupRelation struct {
	ID          string `json:"id" mapstructure:"id"`
	KNID        string `json:"kn_id" mapstructure:"kn_id"`
	Branch      string `json:"branch" mapstructure:"branch"`
	CGID        string `json:"cg_id" mapstructure:"cg_id"`
	ConceptType string `json:"concept_type" mapstructure:"concept_type"`
	ConceptID   string `json:"concept_id" mapstructure:"concept_id"`
	CreateTime  int64  `json:"create_time" mapstructure:"create_time"`

	ModuleType string `json:"module_type" mapstructure:"module_type"`
	// 向量
	Vector []float32 `json:"_vector,omitempty"`
	Score  *float64  `json:"_score,omitempty"` // opensearch检索的得分，在概念搜索时使用
}

// 概念分组的分页查询
type ConceptGroupsQueryParams struct {
	PaginationQueryParameters
	NamePattern string
	Tag         string
	KNID        string
	Branch      string
	CGIDs       []string
}

// 概念与分组关系的分页查询
type ConceptGroupRelationsQueryParams struct {
	PaginationQueryParameters
	KNID        string
	Branch      string
	CGIDs       []string
	ConceptType string
	OTIDs       []string
}

// 对ID去重
func GetUniqueIDs(ids []ID) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(ids))

	for _, id := range ids {
		if !seen[id.ID] {
			seen[id.ID] = true
			result = append(result, id.ID)
		}
	}

	return result
}
