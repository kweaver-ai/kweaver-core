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

	ListDocuments(ctx context.Context, res *Resource, params *ResourceDataQueryParams) ([]map[string]any, int64, error)
	GetDocument(ctx context.Context, id string, docID string) (map[string]any, error)

	CreateDocuments(ctx context.Context, id string, documents []map[string]any) ([]string, error)
	UpdateDocument(ctx context.Context, id string, docID string, document map[string]any) error
	DeleteDocument(ctx context.Context, id string, docID string) error
	UpdateDocuments(ctx context.Context, id string, updateRequests []map[string]any) error
	DeleteDocuments(ctx context.Context, id string, docIDs string) error
	DeleteDocumentsByQuery(ctx context.Context, res *Resource, params *ResourceDataQueryParams) error

	// Build task management methods
	CreateBuildTask(ctx context.Context, id string, req *BuildTaskRequest) (string, error)
	GetBuildTaskByID(ctx context.Context, id string) (*BuildTask, error)
	GetBuildTaskByResourceID(ctx context.Context, resourceID string) (*BuildTask, error)
	GetBuildTasks(ctx context.Context, offset, limit int) ([]*BuildTask, int64, error)
	UpdateBuildTaskStatus(ctx context.Context, taskID string, req *UpdateBuildTaskStatusRequest) error
	DeleteBuildTask(ctx context.Context, taskID string) error
}
