// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// QueryExecuteRequest 统一查询请求；query_id 必填，用于游标 session
type QueryExecuteRequest struct {
	QueryID         string         `json:"query_id"`                   // 必填，同一轮查询一致
	Tables          []TableInQuery `json:"tables"`                     // 表列表
	Joins           []JoinSpec     `json:"joins,omitempty"`            // JOIN 定义；单表时为空
	OutputFields    []string       `json:"output_fields,omitempty"`    // 输出列，可带表别名
	FilterCondition any            `json:"filter_condition,omitempty"` // 过滤条件
	Sort            []*SortField   `json:"sort,omitempty"`             // 排序
	Offset          int            `json:"offset,omitempty"`           // 分页偏移
	Limit           int            `json:"limit,omitempty"`            // 每页条数，最大 10000
	NeedTotal       bool           `json:"need_total,omitempty"`       // 是否返回总条数
}

// TableInQuery 查询中的表定义
type TableInQuery struct {
	ResourceID string `json:"resource_id"`
	Alias      string `json:"alias,omitempty"`
}

// JoinSpec JOIN 定义
type JoinSpec struct {
	Type            string       `json:"type"` // inner, left, right, full
	LeftTableAlias  string       `json:"left_table_alias"`
	RightTableAlias string       `json:"right_table_alias"`
	On              []JoinOnCond `json:"on"`
}

// JoinOnCond JOIN ON 条件
type JoinOnCond struct {
	LeftField  string `json:"left_field"`
	RightField string `json:"right_field"`
}

// QueryExecuteResponse 统一查询响应
type QueryExecuteResponse struct {
	QueryID    string           `json:"query_id"` // 后端生成或回传，用于后续分页
	Entries    []map[string]any `json:"entries"`
	TotalCount *int64           `json:"total_count,omitempty"`
	NextOffset int              `json:"next_offset"`
	HasMore    bool             `json:"has_more"`
}

// ResourceDataQueryParams 扩展：支持 keyset 游标
// 在 interfaces/resource_data.go 中已有定义，此处扩展 Cursor 字段通过 JoinQueryParams 传递

// JoinQueryParams JOIN 查询参数，供 ExecuteJoinQuery 使用
type JoinQueryParams struct {
	Resources         []*Resource
	ResourceIDToAlias map[string]string // resource_id -> alias
	Joins             []*JoinSpec
	OutputFields      []string
	FilterCondCfg     *FilterCondCfg
	ActualFilterCond  FilterCondition
	Sort              []*SortField
	Offset            int
	Limit             int
	NeedTotal         bool
	CursorEncoded     string // keyset 游标值，空则用 OFFSET/LIMIT
}

// QuerySessionStore 游标 session 存储抽象，便于测试与替换
type QuerySessionStore interface {
	GetCursor(ctx context.Context, queryID string, offset int) (cursorEncoded string, ok bool)
	SetCursor(ctx context.Context, queryID string, offset int, cursorEncoded string) error
	Touch(ctx context.Context, queryID string) error // 刷新最后访问时间，用于 TTL
}
