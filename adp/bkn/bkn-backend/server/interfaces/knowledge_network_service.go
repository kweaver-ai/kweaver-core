// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/knowledge_network_service.go -destination ../interfaces/mock/mock_knowledge_network_service.go
type KNService interface {
	CheckKNExistByID(ctx context.Context, knID string, branch string) (string, bool, error)
	CheckKNExistByName(ctx context.Context, knName string, branch string) (string, bool, error)
	CreateKN(ctx context.Context, kn *KN, mode string, validateDependency bool) (string, error)
	ListKNs(ctx context.Context, query KNsQueryParams) ([]*KN, int, error)
	GetKNByID(ctx context.Context, knID string, branch string, mode string) (*KN, error)
	UpdateKN(ctx context.Context, tx *sql.Tx, kn *KN) error
	UpdateKNDetail(ctx context.Context, knID string, branch string, detail string) error
	DeleteKN(ctx context.Context, kn *KN) error

	GetStatByKN(ctx context.Context, kn *KN) (*Statistics, error)
	GetRelationTypePaths(ctx context.Context, query RelationTypePathsBaseOnSource) ([]RelationTypePath, error)

	ListKnSrcs(ctx context.Context, query KNsQueryParams) ([]Resource, int, error)
}
