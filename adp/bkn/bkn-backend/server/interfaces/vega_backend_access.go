// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

// Property represents a field definition in vega-backend schema
type Property struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	DisplayName  string            `json:"display_name"`
	OriginalName string            `json:"original_name"`
	Description  string            `json:"description"`
	Features     []PropertyFeature `json:"features,omitempty"`
}

// PropertyFeature represents a field feature (keyword, fulltext, vector)
type PropertyFeature struct {
	FeatureName string         `json:"name"`
	DisplayName string         `json:"display_name"`
	FeatureType string         `json:"feature_type"` // keyword, fulltext, vector
	Description string         `json:"description"`
	RefProperty string         `json:"ref_property"`
	IsDefault   bool           `json:"is_default"`
	IsNative    bool           `json:"is_native"`
	Config      map[string]any `json:"config"`
}

// CatalogRequest represents create catalog request
type CatalogRequest struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	// ConnectorType string         `json:"connector_type"`
	// ConnectorCfg  map[string]any `json:"connector_config"`
}

// Catalog represents a Catalog entity
type Catalog struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Tags          []string `json:"tags"`
	Description   string   `json:"description"`
	Type          string   `json:"type"`
	Enabled       bool     `json:"enabled"`
	ConnectorType string   `json:"connector_type"`
}

// VegaResource represents a Resource entity in vega-backend
type VegaResource struct {
	ID          string   `json:"id"`
	CatalogID   string   `json:"catalog_id"`
	Name        string   `json:"name"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	// Status           string      `json:"status"`
	SchemaDefinition []*Property `json:"schema_definition,omitempty"`
}

// CatalogsListResponse represents catalogs list response
type CatalogsListResponse struct {
	Data   []*Catalog `json:"data"`
	Total  int        `json:"total"`
	Offset int        `json:"offset"`
	Limit  int        `json:"limit"`
}

// ResourcesListResponse represents resources list response
type ResourcesListResponse struct {
	Data   []*VegaResource `json:"data"`
	Total  int             `json:"total"`
	Offset int             `json:"offset"`
	Limit  int             `json:"limit"`
}

// DatasetQueryResponse represents dataset query response
type DatasetQueryResponse struct {
	Entries     []map[string]any `json:"entries"`
	TotalCount  int64            `json:"total_count"`
	SearchAfter []any            `json:"search_after"`
}

// DatasetQueryParams represents query parameters for dataset data retrieval
type DatasetQueryParams struct {
	FilterCondition map[string]any `json:"filter_condition,omitempty"`
	SearchAfter     []any          `json:"search_after,omitempty"`
	Offset          int            `json:"offset,omitempty"`
	Limit           int            `json:"limit,omitempty"`
	NeedTotal       bool           `json:"need_total,omitempty"`
	Sort            []*SortParams  `json:"sort,omitempty"`
	OutputFields    []string       `json:"output_fields,omitempty"`
}

// VegaBackendAccess defines the interface for accessing vega-backend service
//
//go:generate mockgen -source ../interfaces/vega_backend_access.go -destination ../interfaces/mock/mock_vega_backend_access.go
type VegaBackendAccess interface {
	// GetCatalogByID gets catalog by ID
	GetCatalogByID(ctx context.Context, id string) (*Catalog, error)

	// CreateCatalog creates a new catalog
	CreateCatalog(ctx context.Context, req *CatalogRequest) (*Catalog, error)

	// GetResourceByID gets resource by ID
	GetResourceByID(ctx context.Context, id string) (*VegaResource, error)

	// CreateResource creates a new resource
	CreateResource(ctx context.Context, req *VegaResource) error

	// DeleteResource deletes a resource by ID
	DeleteResource(ctx context.Context, id string) error

	// QueryDatasetData queries data from dataset
	QueryDatasetData(ctx context.Context, datasetID string, params *DatasetQueryParams) (*DatasetQueryResponse, error)

	// WriteDatasetDocuments writes documents to dataset
	WriteDatasetDocuments(ctx context.Context, datasetID string, documents []map[string]any) error

	// DeleteDatasetDocumentByID deletes a document by ID from dataset
	DeleteDatasetDocumentByID(ctx context.Context, datasetID string, docID string) error

	// DeleteDatasetDocumentsByQuery deletes documents by query condition from dataset
	DeleteDatasetDocumentsByQuery(ctx context.Context, datasetID string, filterCondition map[string]any) error
}
