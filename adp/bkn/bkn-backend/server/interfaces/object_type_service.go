// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/object_type_service.go -destination ../interfaces/mock/mock_object_type_service.go
type ObjectTypeService interface {
	CheckObjectTypeExistByID(ctx context.Context, knID string, branch string, otID string) (string, bool, error)
	CheckObjectTypeExistByName(ctx context.Context, knID string, branch string, otName string) (string, bool, error)
	CreateObjectTypes(ctx context.Context, tx *sql.Tx, objectTypes []*ObjectType, mode string, needCreateConceptGroupRelation bool, strictMode bool) ([]string, error)
	ListObjectTypes(ctx context.Context, tx *sql.Tx, query ObjectTypesQueryParams) ([]*ObjectType, int, error)
	GetObjectTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, otIDs []string) ([]*ObjectType, error)
	UpdateObjectType(ctx context.Context, tx *sql.Tx, objectType *ObjectType, strictMode bool) error
	// UpdateDataProperties updates object type data properties. When strictMode is true, vector index config (if enabled) is validated against the embedding small model service.
	UpdateDataProperties(ctx context.Context, objectType *ObjectType, dataProperties []*DataProperty, strictMode bool) error
	DeleteObjectTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, otIDs []string) error

	GetObjectTypeByID(ctx context.Context, tx *sql.Tx, knID string, branch string, otID string) (*ObjectType, error)
	GetAllObjectTypesByKnID(ctx context.Context, knID string, branch string) (map[string]*ObjectType, error)
	GetObjectTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error)
	DeleteObjectTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error

	SearchObjectTypes(ctx context.Context, query *ConceptsQuery) (ObjectTypes, error)

	// 获取对象类基本信息（无翻译依赖资源）
	GetObjectTypesMapByIDs(ctx context.Context, knID string, branch string, otIDs []string, needPropMap bool) (map[string]*ObjectType, error)

	// 对象类写索引
	InsertDatasetData(ctx context.Context, objectTypes []*ObjectType) error

	// ValidateObjectTypes 仅校验依赖存在性，不写库
	// When batch is nil, behavior matches history (DB existence only). For full-KN / concept-group preflight, pass an index from batchindex.CollectKNFromPayload / batchindex.CollectFromConceptGroups.
	ValidateObjectTypes(ctx context.Context, knID string, branch string, objectTypes []*ObjectType, strictMode bool, batch *BatchIDIndex, mode string) error
}
