// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	cond "ontology-query/common/condition"
	"sync"
)

var (
	DIRECTION_MAP = map[string]bool{
		DIRECTION_FORWARD:       true,
		DIRECTION_BACKWARD:      true,
		DIRECTION_BIDIRECTIONAL: true,
	}
)

// 基于起点、方向和路径长度获取对象子图的请求体
type SubGraphQueryBaseOnSource struct {
	ConceptGroups         []string       `json:"concept_groups,omitempty"`
	SourceObjecTypeId     string         `json:"source_object_type_id"`
	Condition             map[string]any `json:"condition,omitempty"`
	Direction             string         `json:"direction"`
	PathLength            int            `json:"path_length"`
	IncludeIncompletePath bool           `json:"include_incomplete_path,omitempty"` // 当路径探索无下一个节点时，是否拼接当前路径（true：拼接，false：跳过）
	PageQuery

	ActualCondition *cond.CondCfg `json:"-"`
	KNID            string        `json:"-"`
	Branch          string        `json:"-"`
	CommonQueryParameters
	*PathQuotaManager
	BatchQueryState
}

// 基于路径查子图的请求体
type SubGraphQueryBaseOnTypePath struct {
	Paths  QueryRelationTypePaths
	KNID   string
	Branch string
	CommonQueryParameters
}

// 基于一组对象实例组织关系子图的请求体
type SubGraphQueryBaseOnObjects struct {
	Entries []InputObjectInstance `json:"entries"`
	KNID    string                `json:"-"`
	Branch  string                `json:"-"`
	CommonQueryParameters
}

// 输入的对象实例
type InputObjectInstance struct {
	ObjectTypeID     string         `json:"object_type_id"`
	InstanceIdentity map[string]any `json:"_instance_identity"`
}

type QueryRelationTypePaths struct {
	TypePaths []QueryRelationTypePath `json:"relation_type_paths"`
}

type QueryRelationTypePath struct {
	ObjectTypes []ObjectTypeWithKeyField `json:"object_types"` // key是对象类id
	Edges       []TypeEdge               `json:"relation_types"`
	Limit       int                      `json:"limit"` // 当前路径返回的数量
	// ObjectTypes   []ObjectTypeWithKeyField `json:"-"`     // key是对象类id
}

// type ObjectTypeInRequestPath struct {
// 	OTID      string   `json:"id"`
// 	Condition *CondCfg `json:"condition,omitempty"` // 对起点对象类的过滤条件
// 	PageQuery          // 对路径起点对象类的排序和limit
// }

// 路径配额管理策略
type PathQuotaManager struct {
	TotalLimit         int64    `json:"-"` // 总路径数量限制，当前是全局配置 1w
	GlobalCount        int64    `json:"-"` // 全局已经添加了的对象路径数量
	UsedQuota          sync.Map `json:"-"` // 已使用配额
	RequestPathTypeNum int      `json:"-"` // 当前请求的概念路径的数量
}

// 批量查询的中间状态
type BatchQueryState struct {
	Visited   map[string]bool `json:"-"`
	BatchSize int             `json:"-"`
}

// 对象子图的返回体
type ObjectSubGraph struct {
	Objects           map[string]ObjectInfoInSubgraph `json:"objects"`
	IsolatedObjects   map[string]ObjectInfoInSubgraph `json:"isolated_objects,omitempty"`
	RelationPaths     []RelationPath                  `json:"relation_paths"`
	TotalCount        int64                           `json:"total_count,omitempty"`
	SearchAfter       []any                           `json:"search_after,omitempty"`
	CuurentPathNumber int                             `json:"current_path_number,omitempty"`
	OverallMs         int64                           `json:"overall_ms"`
}

// 基于路径查询的返回体
type PathsEntries struct {
	Entries []ObjectSubGraph `json:"entries"`
}

type ObjectSystemInfo struct {
	InstanceID       any            `json:"_instance_id"`
	InstanceIdentity map[string]any `json:"_instance_identity"`
	Display          any            `json:"_display"`
}

// 在对象子图中的对象信息
type ObjectInfoInSubgraph struct {
	ObjectSystemInfo
	ObjectTypeId   string         `json:"object_type_id"`
	ObjectTypeName string         `json:"object_type_name"`
	Properties     map[string]any `json:"properties"`
}

// 由关系实例组成的路径
type RelationPath struct {
	Relations []Relation `json:"relations"`
	Length    int        `json:"length"`
}

// 关系实例
type Relation struct {
	RelationTypeId   string `json:"relation_type_id"`
	RelationTypeName string `json:"relation_type_name"`
	SourceObjectId   string `json:"source_object_id"`
	TargetObjectId   string `json:"target_object_id"`
}

// 从本体引擎中获取到的概念路径
type RelationTypePath struct {
	ObjectTypes []ObjectTypeWithKeyField `json:"object_types"`
	TypeEdges   []TypeEdge               `json:"relation_types"`
	Length      int                      `json:"length"`

	ID int `json:"-"` // 概念路径id，为后续对象路径配额限制使用
}

type TypeEdge struct {
	RelationTypeId     string       `json:"relation_type_id"`
	RelationType       RelationType `json:"relation_type"`
	SourceObjectTypeId string       `json:"source_object_type_id"`
	TargetObjectTypeId string       `json:"target_object_type_id"`
	Direction          string       `json:"direction"`
}

type LevelObject struct {
	ObjectID   string
	ObjectUK   map[string]any // 对象主键
	ObjectData map[string]any
	ObjectType *ObjectType
	PathFrom   string // 记录从哪个对象来的，用于构建路径
}

type LevelObjectWithPath struct {
	LevelObject
	Paths []RelationPath // 从起点到当前对象的所有路径
}
