// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// CatalogAccess defines catalog data access interface.
//
//go:generate mockgen -source ../interfaces/catalog_access.go -destination ../interfaces/mock/mock_catalog_access.go
type CatalogAccess interface {
	// Create creates a new Catalog.
	Create(ctx context.Context, catalog *Catalog) error
	// GetByID retrieves a Catalog by ID.
	GetByID(ctx context.Context, id string) (*Catalog, error)
	// GetByIDs retrieves a Catalog by IDs.
	GetByIDs(ctx context.Context, ids []string) ([]*Catalog, error)
	// AttachListExtensions 按列表查询参数加载或清空根级 extensions（供 List 在 GetByIDs 之后调用）。
	AttachListExtensions(ctx context.Context, params CatalogsQueryParams, catalogs []*Catalog) error
	// GetByName retrieves a Catalog by name.
	GetByName(ctx context.Context, name string) (*Catalog, error)
	// List lists Catalogs with filters.
	List(ctx context.Context, params CatalogsQueryParams) ([]*Catalog, int64, error)
	// ListIDs lists Catalog IDs with filters.
	ListIDs(ctx context.Context, params CatalogsQueryParams) ([]string, error)
	// Update updates a Catalog.
	Update(ctx context.Context, catalog *Catalog) error
	// DeleteByIDs deletes Catalogs by IDs.
	DeleteByIDs(ctx context.Context, ids []string) error
	// UpdateHealthCheckStatus updates Catalog health check status.
	UpdateHealthCheckStatus(ctx context.Context, id string, status CatalogHealthCheckStatus) error

	// UpdateMetadata updates a Catalog metadata.
	UpdateMetadata(ctx context.Context, id string, metadata map[string]any) error

	// ListCatalogSrcs lists Catalog Sources with filters.
	ListCatalogSrcs(ctx context.Context, params ListCatalogsQueryParams) ([]*ListCatalogEntry, int64, error)
	// ListCatalogSrcsIDs lists Catalog Source IDs with filters.
	ListCatalogSrcsIDs(ctx context.Context, params ListCatalogsQueryParams) ([]string, error)
	// ListCatalogSrcsByIDs lists Catalog Sources by IDs.
	ListCatalogSrcsByIDs(ctx context.Context, ids []string) ([]*ListCatalogEntry, error)
}
