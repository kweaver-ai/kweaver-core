// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/relation_type_access.go -destination ../interfaces/mock/mock_relation_type_access.go
type RelationTypeAccess interface {
	CheckRelationTypeExistByID(ctx context.Context, knID string, branch string, rtID string) (string, bool, error)

	CreateRelationType(ctx context.Context, tx *sql.Tx, relationType *RelationType) error
	ListRelationTypes(ctx context.Context, query RelationTypesQueryParams) ([]*RelationType, error)
	GetRelationTypesTotal(ctx context.Context, query RelationTypesQueryParams) (int, error)
	GetRelationTypeByID(ctx context.Context, knID string, branch string, rtID string) (*RelationType, error)
	GetRelationTypesByIDs(ctx context.Context, knID string, branch string, rtIDs []string) ([]*RelationType, error)
	UpdateRelationType(ctx context.Context, tx *sql.Tx, relationType *RelationType) error
	DeleteRelationTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, rtIDs []string) (int64, error)

	GetAllRelationTypesByKnID(ctx context.Context, knID string, branch string) (map[string]*RelationType, error)
	GetRelationTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error)
	DeleteRelationTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error)
}
