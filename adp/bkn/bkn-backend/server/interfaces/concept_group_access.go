// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/concept_group_access.go -destination ../interfaces/mock/mock_concept_group_access.go
type ConceptGroupAccess interface {
	CheckConceptGroupExistByID(ctx context.Context, knID string, branch string, cgID string) (string, bool, error)
	CheckConceptGroupExistByName(ctx context.Context, knID string, branch string, cgName string) (string, bool, error)
	CreateConceptGroup(ctx context.Context, tx *sql.Tx, conceptGroup *ConceptGroup) error
	ListConceptGroups(ctx context.Context, query ConceptGroupsQueryParams) ([]*ConceptGroup, error)
	GetConceptGroupByID(ctx context.Context, knID string, branch string, cgID string) (*ConceptGroup, error)
	UpdateConceptGroup(ctx context.Context, tx *sql.Tx, conceptGroup *ConceptGroup) error
	UpdateConceptGroupDetail(ctx context.Context, knID string, branch string, cgID string, detail string) error
	DeleteConceptGroupByID(ctx context.Context, tx *sql.Tx, knID string, branch string, cgID string) (int64, error)

	GetConceptGroupsByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, cgIDs []string) ([]*ConceptGroup, error)
	GetConceptGroupsTotal(ctx context.Context, query ConceptGroupsQueryParams) (int, error)
	GetAllConceptGroupsByKnID(ctx context.Context, knID string, branch string) (map[string]*ConceptGroup, error)
	GetConceptGroupIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error)
	DeleteConceptGroupsByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error)
	DeleteConceptGroupRelationsByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error)

	ListConceptGroupRelations(ctx context.Context, tx *sql.Tx, query ConceptGroupRelationsQueryParams) ([]ConceptGroupRelation, error)
	CreateConceptGroupRelation(ctx context.Context, tx *sql.Tx, kn *ConceptGroupRelation) error
	DeleteObjectTypesFromGroup(ctx context.Context, tx *sql.Tx, query ConceptGroupRelationsQueryParams) (int64, error)
	// DeleteObjectTypesFromGroup(ctx context.Context, tx *sql.Tx, knID string, branch string, cgID string, otIDs []string) (int64, error)

	// 从概念与分组关系中获取对象类id（join了对象类表）

	GetConceptIDsByConceptGroupIDs(ctx context.Context, knID string, branch string, cgIDs []string, conceptType string) ([]string, error)
	GetRelationTypeIDsFromConceptGroupRelation(ctx context.Context, query ConceptGroupRelationsQueryParams) ([]string, error)
	GetActionTypeIDsFromConceptGroupRelation(ctx context.Context, query ConceptGroupRelationsQueryParams) ([]string, error)
	GetConceptGroupsByOTIDs(ctx context.Context, tx *sql.Tx, query ConceptGroupRelationsQueryParams) (map[string][]*ConceptGroup, error)
}
