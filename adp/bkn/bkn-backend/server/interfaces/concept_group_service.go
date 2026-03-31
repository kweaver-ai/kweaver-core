// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/concept_group_service.go -destination ../interfaces/mock/mock_concept_group_service.go
type ConceptGroupService interface {
	CheckConceptGroupExistByID(ctx context.Context, knID string, branch string, cgID string) (string, bool, error)
	CheckConceptGroupExistByName(ctx context.Context, knID string, branch string, cgName string) (string, bool, error)
	CreateConceptGroup(ctx context.Context, tx *sql.Tx, conceptGroup *ConceptGroup, mode string, validateDependency bool) (string, error)
	ListConceptGroups(ctx context.Context, query ConceptGroupsQueryParams) ([]*ConceptGroup, int, error)
	GetConceptGroupByID(ctx context.Context, knID string, branch string, cgID string, mode string) (*ConceptGroup, error)
	UpdateConceptGroup(ctx context.Context, tx *sql.Tx, conceptGroup *ConceptGroup) error
	UpdateConceptGroupDetail(ctx context.Context, knID string, branch string, cgID string, detail string) error
	DeleteConceptGroupByID(ctx context.Context, tx *sql.Tx, knID string, branch string, cgID string) error

	GetStatByConceptGroup(ctx context.Context, conceptGroup *ConceptGroup) (*Statistics, error)
	GetConceptGroupIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error)
	DeleteConceptGroupsByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error

	AddObjectTypesToConceptGroup(ctx context.Context, tx *sql.Tx, knID string, branch string, cgID string, otIDs []ID, importMode string) ([]string, error)
	ListConceptGroupRelations(ctx context.Context, query ConceptGroupRelationsQueryParams) ([]ConceptGroupRelation, error)
	DeleteObjectTypesFromGroup(ctx context.Context, tx *sql.Tx, knID string, branch string, cgID string, otIDs []string) error
}
