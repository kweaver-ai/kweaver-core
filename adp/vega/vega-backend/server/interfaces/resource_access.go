// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// ResourceAccess defines resource data access interface.
//
//go:generate mockgen -source ../interfaces/resource_access.go -destination ../interfaces/mock/mock_resource_access.go
type ResourceAccess interface {
	// Create creates a new Resource.
	Create(ctx context.Context, resource *Resource) error
	// GetByID retrieves a Resource by ID.
	GetByID(ctx context.Context, id string) (*Resource, error)
	// GetByIDs retrieves Resources by IDs.
	GetByIDs(ctx context.Context, ids []string) ([]*Resource, error)
	// AttachListExtensions 按列表查询参数加载根级 extensions（供 List 在 GetByIDsBasic 之后调用）。
	AttachListExtensions(ctx context.Context, params ResourcesQueryParams, resources []*Resource) error
	// GetByIDsBasic retrieves Resources by IDs without parsing sourceMetadata, schemaDefinition and logicDefinition.
	GetByIDsBasic(ctx context.Context, ids []string) ([]*Resource, error)
	// GetByName retrieves a Resource by catalog and name.
	GetByName(ctx context.Context, catalogID string, name string) (*Resource, error)
	// GetByCatalogID retrieves all Resources under a Catalog.
	GetByCatalogID(ctx context.Context, catalogID string) ([]*Resource, error)
	// List lists Resources with filters.
	List(ctx context.Context, params ResourcesQueryParams) ([]*Resource, int64, error)
	// ListIDs lists Resource IDs with filters.
	ListIDs(ctx context.Context, params ResourcesQueryParams) ([]string, error)
	// Update updates a Resource.
	Update(ctx context.Context, resource *Resource) error
	// UpdateStatus updates a Resource's status.
	UpdateStatus(ctx context.Context, id string, status string, statusMessage string) error
	// DeleteByIDs deletes Resources by IDs.
	DeleteByIDs(ctx context.Context, ids []string) error

	// ListResourceSrcs lists Resource Sources with filters.
	ListResourceSrcs(ctx context.Context, params ListResourcesQueryParams) ([]*ListResourceEntry, int64, error)
	// ListResourceSrcsIDs lists Resource Source IDs with filters.
	ListResourceSrcsIDs(ctx context.Context, params ListResourcesQueryParams) ([]string, error)
	// ListResourceSrcsByIDs lists Resource Sources by IDs.
	ListResourceSrcsByIDs(ctx context.Context, ids []string) ([]*ListResourceEntry, error)

	// CheckExistByCategories checks if Resources exists by catalog ID and categories.
	CheckExistByCategories(ctx context.Context, catalogID string, categories []string) (bool, error)

	// DeleteByCatalogIDs deletes Resources by catalog IDs.
	DeleteByCatalogIDs(ctx context.Context, catalogIDs []string) error
}
