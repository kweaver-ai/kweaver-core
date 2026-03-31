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
	// GetByName retrieves a Resource by catalog and name.
	GetByName(ctx context.Context, catalogID string, name string) (*Resource, error)
	// GetByCatalogID retrieves all Resources under a Catalog.
	GetByCatalogID(ctx context.Context, catalogID string) ([]*Resource, error)
	// List lists Resources with filters.
	List(ctx context.Context, params ResourcesQueryParams) ([]*Resource, int64, error)
	// Update updates a Resource.
	Update(ctx context.Context, resource *Resource) error
	// UpdateStatus updates a Resource's status.
	UpdateStatus(ctx context.Context, id string, status string, statusMessage string) error
	// DeleteByIDs deletes Resources by IDs.
	DeleteByIDs(ctx context.Context, ids []string) error

	// ListResourceSrcs lists Resource Sources with filters.
	ListResourceSrcs(ctx context.Context, params ListResourcesQueryParams) ([]*ListResourceEntry, int64, error)

	// CheckExistByCategories checks if Resources exists by catalog ID and categories.
	CheckExistByCategories(ctx context.Context, catalogID string, categories []string) (bool, error)

	// DeleteByCatalogIDs deletes Resources by catalog IDs.
	DeleteByCatalogIDs(ctx context.Context, catalogIDs []string) error
}
