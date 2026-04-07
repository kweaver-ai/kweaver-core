// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

// 基于起点、方向和路径长度获取对象子图的请求体
type PathsQueryBaseOnSource struct {
	ConceptGroups     []string `json:"concept_groups,omitempty"`
	SourceObjecTypeId string   `json:"source_object_type_id"`
	Direction         string   `json:"direction"`
	PathLength        int      `json:"path_length"`

	KNID string `json:"-"`
	// IncludeTypeInfo bool   `json:"-"`
}

// 关系类列表查询参数
type RelationTypesQuery struct {
	// 单个对象类型ID查询（向后兼容）
	SourceObjectTypeID string `json:"source_object_type_id,omitempty"`
	TargetObjectTypeID string `json:"target_object_type_id,omitempty"`

	// 多个对象类型ID查询（新增）
	SourceObjectTypeIDs []string `json:"source_object_type_ids,omitempty"`
	TargetObjectTypeIDs []string `json:"target_object_type_ids,omitempty"`
}

//go:generate mockgen -source ../interfaces/ontology_manager_access.go -destination ../interfaces/mock/mock_ontology_manager_access.go
type OntologyManagerAccess interface {
	GetObjectType(ctx context.Context, knID string, branch string, otId string) (ObjectType, bool, error)
	GetRelationType(ctx context.Context, knID string, branch string, rtId string) (RelationType, bool, error)
	GetActionType(ctx context.Context, knID string, branch string, atId string) (ActionType, map[string]any, bool, error)
	GetRelationTypePathsBaseOnSource(ctx context.Context, knID string, branch string, query PathsQueryBaseOnSource) ([]RelationTypePath, error)
	ListRelationTypes(ctx context.Context, knID string, branch string, query RelationTypesQuery) ([]RelationType, error)
	GetRiskTypesByIDs(ctx context.Context, knID string, branch string, riskTypeIDs []string) ([]RiskType, error)
}
