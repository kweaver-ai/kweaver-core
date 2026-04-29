// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

const (
	QueryType_Standard = "standard" // 标准查询
	QueryType_Stream   = "stream"   // 流式查询
)

// SQLQueryRequest SQL查询请求
type SQLQueryRequest struct {
	Query        any    `json:"query"`                   // SQL查询语句（字符串）或OpenSearch DSL查询（JSON对象）
	QueryType    string `json:"query_type,omitempty"`    // 可选，指定查询类型（如"standard"、"stream"）
	QueryID      string `json:"query_id,omitempty"`      // 可选，指定查询 ID，用于游标 session
	ResourceType string `json:"resource_type,omitempty"` // 可选，指定资源类型（如"opensearch"、"mysql"、"postgresql"）
	StreamSize   int    `json:"stream_size,omitempty"`   // 流式查询每批数据量，默认10000，流式查询必填
	QueryTimeout int    `json:"query_timeout,omitempty"` // 查询超时时间（秒），默认60，最小1，最大3600
}

// SQLQueryResponse SQL查询响应
type SQLQueryResponse struct {
	Columns    []ColumnInfo     `json:"columns"`     // 列信息
	Entries    []map[string]any `json:"entries"`     // 查询结果
	Stats      QueryStats       `json:"stats"`       // 查询统计
	TotalCount int64            `json:"total_count"` // 总条数
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name string `json:"name"` // 列名
	Type string `json:"type"` // 列类型
}

// QueryStats 查询统计
type QueryStats struct {
	IsTimeout   bool   `json:"is_timeout"`             // 是否超时
	QueryID     string `json:"query_id,omitempty"`     // 查询 ID
	HasMore     bool   `json:"has_more"`               // 是否还有更多数据
	SearchAfter []any  `json:"search_after,omitempty"` // OpenSearch流式查询的search_after值
	Offset      int    `json:"offset"`                 // 已获取到的数据总数（流式查询模式）
}

//go:generate mockgen -source ../interfaces/query.go -destination ../interfaces/mock/mock_query.go

// SQLQueryService SQL查询服务接口
type SQLQueryService interface {
	// Execute 执行SQL查询
	Execute(ctx context.Context, req *SQLQueryRequest) (*SQLQueryResponse, error)
}
