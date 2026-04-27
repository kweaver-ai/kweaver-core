// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/relation_type_service.go -destination ../interfaces/mock/mock_relation_type_service.go
type RelationTypeService interface {
	CheckRelationTypeExistByID(ctx context.Context, knID string, branch string, rtID string) (string, bool, error)
	CreateRelationTypes(ctx context.Context, tx *sql.Tx, relationTypes []*RelationType, mode string, validateDependency bool) ([]string, error)
	ListRelationTypes(ctx context.Context, query RelationTypesQueryParams) ([]*RelationType, int, error)
	GetRelationTypesByIDs(ctx context.Context, knID string, branch string, rtIDs []string) ([]*RelationType, error)
	UpdateRelationType(ctx context.Context, tx *sql.Tx, relationType *RelationType, strictMode bool) error
	DeleteRelationTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, rtIDs []string) error

	GetRelationTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error)
	DeleteRelationTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error

	SearchRelationTypes(ctx context.Context, query *ConceptsQuery) (RelationTypes, error)

	// 写关系类到索引中
	InsertDatasetData(ctx context.Context, relationTypes []*RelationType) error

	// ValidateRelationTypes 仅校验依赖存在性，不写库
	ValidateRelationTypes(ctx context.Context, knID string, branch string, relationTypes []*RelationType, strictMode bool, batch *BatchIDIndex, mode string) error
}
