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

	DEFAULT_DATA_LIMIT = 10

	// 日历间隔常量 - 参照OpenSearch的calendar_interval枚举定义
	CALENDAR_UNIT_MINUTE  = "minute"
	CALENDAR_UNIT_HOUR    = "hour"
	CALENDAR_UNIT_DAY     = "day"
	CALENDAR_UNIT_WEEK    = "week"
	CALENDAR_UNIT_MONTH   = "month"
	CALENDAR_UNIT_QUARTER = "quarter"
	CALENDAR_UNIT_YEAR    = "year"
)

// SortField represents a field to sort by.
type SortField struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// Aggregation represents an aggregation operation.
type Aggregation struct {
	Property string `json:"property"` // 被聚合的资源字段名
	Aggr     string `json:"aggr"`     // 聚合函数: count, count_distinct, sum, max, min, avg
	Alias    string `json:"alias,omitempty"`
}

// GroupByItem represents a group by dimension.
type GroupByItem struct {
	Property         string `json:"property"`                    // 分组维度
	Description      string `json:"description,omitempty"`       // 仅文档/调试
	CalendarInterval string `json:"calendar_interval,omitempty"` // date_histogram 的 calendar_interval 参数，支持：minute, hour, day, week, month, quarter, year
}

// HavingClause represents a HAVING clause for aggregation filtering.
type HavingClause struct {
	Field     string `json:"field"`     // 固定为 "__value"
	Operation string `json:"operation"` // ==, !=, >, >=, <, <=, in, not_in, range, out_range
	Value     any    `json:"value"`
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

	// 聚合查询相关字段
	Aggregation *Aggregation   `json:"aggregation,omitempty"` // 聚合度量
	GroupBy     []*GroupByItem `json:"group_by,omitempty"`    // 分组维度
	Having      *HavingClause  `json:"having,omitempty"`      // 对聚合结果过滤（HAVING）
}
