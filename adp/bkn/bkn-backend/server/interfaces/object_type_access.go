// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source ../interfaces/object_type_access.go -destination ../interfaces/mock/mock_object_type_access.go
type ObjectTypeAccess interface {
	CheckObjectTypeExistByID(ctx context.Context, knID string, branch string, otID string) (string, bool, error)
	CheckObjectTypeExistByName(ctx context.Context, knID string, branch string, otName string) (string, bool, error)

	CreateObjectType(ctx context.Context, tx *sql.Tx, objectType *ObjectType) error
	CreateObjectTypeStatus(ctx context.Context, tx *sql.Tx, objectType *ObjectType) error
	ListObjectTypes(ctx context.Context, tx *sql.Tx, query ObjectTypesQueryParams) ([]*ObjectType, error)
	GetObjectTypesTotal(ctx context.Context, query ObjectTypesQueryParams) (int, error)
	GetObjectTypeByID(ctx context.Context, tx *sql.Tx, knID string, branch string, otID string) (*ObjectType, error)
	GetObjectTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, otIDs []string) ([]*ObjectType, error)
	UpdateObjectType(ctx context.Context, tx *sql.Tx, objectType *ObjectType) error
	DeleteObjectTypesByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, otIDs []string) (int64, error)
	DeleteObjectTypeStatusByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, otIDs []string) (int64, error)
	DeleteObjectTypesByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error)
	DeleteObjectTypeStatusByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) (int64, error)
	UpdateDataProperties(ctx context.Context, tx *sql.Tx, objectType *ObjectType) error

	GetAllObjectTypesByKnID(ctx context.Context, knID string, branch string) (map[string]*ObjectType, error)
	GetObjectTypeIDsByKnID(ctx context.Context, knID string, branch string) ([]string, error)
	UpdateObjectTypeStatus(ctx context.Context, tx *sql.Tx, knID string, branch string, otID string, otStatus ObjectTypeStatus) error
}
