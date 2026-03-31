// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// ResourceService defines resource business logic interface.
//
//go:generate mockgen -source ../interfaces/resource_service.go -destination ../interfaces/mock/mock_resource_service.go
type ResourceService interface {
	// Create creates a new Resource.
	Create(ctx context.Context, req *ResourceRequest) (string, error)
	// Get retrieves a Resource by ID.
	GetByID(ctx context.Context, id string) (*Resource, error)
	// GetByIDs retrieves Resources by IDs.
	GetByIDs(ctx context.Context, ids []string) ([]*Resource, error)
	// GetByCatalogID retrieves all Resources under a Catalog.
	GetByCatalogID(ctx context.Context, catalogID string) ([]*Resource, error)
	// GetByName retrieves a Resource by catalog and name.
	GetByName(ctx context.Context, catalogID string, name string) (*Resource, error)
	// List lists Resources with filters.
	List(ctx context.Context, params ResourcesQueryParams) ([]*Resource, int64, error)
	// Update updates a Resource.
	Update(ctx context.Context, id string, req *ResourceRequest) error
	// UpdateStatus updates a Resource's status.
	UpdateStatus(ctx context.Context, id string, status string, statusMessage string) error
	// DeleteByIDs deletes Resources by IDs.
	DeleteByIDs(ctx context.Context, ids []string) error
	// CheckExistByID checks if a Resource exists by ID.
	CheckExistByID(ctx context.Context, id string) (bool, error)
	// CheckExistByName checks if a Resource exists by name.
	CheckExistByName(ctx context.Context, catalogID string, name string) (bool, error)

	// UpdateResource updates a Resource directly.
	UpdateResource(ctx context.Context, resource *Resource) error

	// ListResourceSrcs lists Resource Sources with filters.
	ListResourceSrcs(ctx context.Context, params ListResourcesQueryParams) ([]*ListResourceEntry, int64, error)

	// CheckExistByCategories checks if Resources exists by catalog ID and categories.
	CheckExistByCategories(ctx context.Context, catalogID string, categories []string) (bool, error)
}
