// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

type Hit struct {
	Source map[string]interface{} `json:"_source"`
	Sort   []any                  `json:"sort"`
	Score  float64                `json:"_score"`
}

//go:generate mockgen -source ../interfaces/opensearch_access.go -destination ../interfaces/mock/mock_opensearch_access.go

// OpenSearchAccess 定义OpenSearch访问接口
type OpenSearchAccess interface {
	CreateIndex(ctx context.Context, indexName string, body any) error

	// IndexExists 检查指定索引是否存在
	IndexExists(ctx context.Context, indexName string) (bool, error)

	// DeleteIndex 按照indexName删除索引
	DeleteIndex(ctx context.Context, indexName string) error

	// InsertData 向指定索引写入数据，并指定文档ID
	InsertData(ctx context.Context, indexName string, docID string, data any) error

	// BulkInsertData 批量写入数据到指定索引
	BulkInsertData(ctx context.Context, indexName string, dataList []any) error

	// SearchData 搜索指定索引中的数据
	SearchData(ctx context.Context, indexName string, query any) ([]Hit, error)

	// DeleteData 按照索引名和数据id删除数据
	DeleteData(ctx context.Context, indexName string, docID string) error

	// BulkDeleteData 批量删除数据，参数为docID数组
	BulkDeleteData(ctx context.Context, indexName string, docIDs []string) error

	Count(ctx context.Context, indexName string, query any) ([]byte, error)
}
