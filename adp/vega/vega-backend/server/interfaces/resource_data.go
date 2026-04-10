// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "time"

const (
	Format_Original = "original"
	Format_Flat     = "flat"

	// 最大查询长度设置为10000
	MAX_SEARCH_SIZE = 10000
)

// SortField represents a field to sort by.
type SortField struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// ResourceDataQueryParams represents query parameters for data retrieval.
type ResourceDataQueryParams struct {
	Offset int          `json:"offset,omitempty"`
	Limit  int          `json:"limit,omitempty"`
	Sort   []*SortField `json:"sort,omitempty"`

	FilterCondition any `json:"filter_condition,omitempty"`

	OutputFields []string `json:"output_fields"` // 指定输出的字段列表

	NeedTotal   bool          `json:"need_total,omitempty"`
	Format      string        `json:"-"`
	Timeout     time.Duration `json:"-"`                      // 超时时间，查询参数
	SearchAfter []any         `json:"search_after,omitempty"` // OpenSearch search after参数

	QueryType string `json:"query_type"`

	FilterCondCfg    *FilterCondCfg  `json:"-"`
	ActualFilterCond FilterCondition `json:"-"`

	// CursorEncoded keyset 游标值，由 query session 注入；非空时用 WHERE (sort_cols) > cursor 替代 OFFSET
	CursorEncoded string `json:"-"`

	// SearchKeyword 搜索关键字，用于文件集资源的模糊搜索
	SearchKeyword string `json:"search_keyword,omitempty"`
}
