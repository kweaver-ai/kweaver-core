// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// DatasetService 定义 dataset 业务逻辑接口
//
//go:generate mockgen -source ../interfaces/dataset_service.go -destination ../interfaces/mock/mock_dataset_service.go
type DatasetService interface {
	Create(ctx context.Context, res *Resource) error
	Update(ctx context.Context, res *Resource) error
	Delete(ctx context.Context, res *Resource) error

	ListDocuments(ctx context.Context, indexName string, res *Resource, params *ResourceDataQueryParams) ([]map[string]any, int64, error)
	GetDocument(ctx context.Context, id string, docID string) (map[string]any, error)

	CreateDocuments(ctx context.Context, id string, documents []map[string]any) ([]string, error)
	DeleteDocument(ctx context.Context, id string, docID string) error
	UpsertDocuments(ctx context.Context, id string, updateRequests []map[string]any) ([]string, error)
	DeleteDocuments(ctx context.Context, id string, docIDs string) error
	DeleteDocumentsByQuery(ctx context.Context, indexName string, res *Resource, params *ResourceDataQueryParams) error
}
