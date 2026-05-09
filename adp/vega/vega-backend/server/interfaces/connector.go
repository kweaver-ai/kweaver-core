// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

// ConnectorConfig holds data source connection configuration.
type ConnectorConfig map[string]any

// TableMeta represents table/asset metadata.
type TableMeta struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Database    string                `json:"database"`   // 所属数据库名称（实例级连接时使用）
	Schema      string                `json:"schema"`     // 所属schema名称（实例级连接时使用）
	TableType   string                `json:"table_type"` // table | view | materialized_view
	Properties  map[string]any        `json:"properties"` // 扩展属性：charset, collation, engine, row_count 等
	Columns     []TableColumnMeta     `json:"columns"`
	PKs         []string              `json:"primary_keys"`
	Indices     []TableIndexMeta      `json:"indices"`      // 索引列表
	ForeignKeys []TableForeignKeyMeta `json:"foreign_keys"` // 外键列表

}

// TableColumnMeta represents column metadata.
type TableColumnMeta struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	OrigType          string `json:"orig_type"`
	Nullable          bool   `json:"nullable"`
	DefaultValue      string `json:"default_value,omitempty"` // 默认值
	Description       string `json:"description"`
	CharMaxLen        int    `json:"char_max_len,omitempty"`       // 字符最大长度
	NumPrecision      int    `json:"num_precision,omitempty"`      // 数值精度
	NumScale          int    `json:"num_scale,omitempty"`          // 数值小数位
	DatetimePrecision int    `json:"datetime_precision,omitempty"` // 日期时间精度
	Charset           string `json:"charset,omitempty"`            // 字符集
	Collation         string `json:"collation,omitempty"`          // 排序规则
	OrdinalPosition   int    `json:"ordinal_position"`             // 列位置（从1开始）
	ColumnKey         string `json:"column_key"`                   // 列键
}

// TableIndexMeta represents index metadata.
type TableIndexMeta struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Primary bool     `json:"primary"`
}

// TableForeignKeyMeta represents foreign key metadata.
type TableForeignKeyMeta struct {
	Name       string   `json:"name"`
	Columns    []string `json:"columns"`
	RefTable   string   `json:"ref_table"`
	RefColumns []string `json:"ref_columns"`
	OnDelete   string   `json:"on_delete,omitempty"`
	OnUpdate   string   `json:"on_update,omitempty"`
}

// QueryResult represents query execution result.
type QueryResult struct {
	Columns     []string         `json:"columns"`
	Rows        []map[string]any `json:"rows"`
	Total       int64            `json:"total"`
}

// FileMeta represents file metadata.
type FileMeta struct {
	Path         string `json:"path"`
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	LastModified int64  `json:"last_modified"`
	ContentType  string `json:"content_type"`
}

// FilesetMeta represents a file or folder object from a fileset-capable source (e.g. AnyShare).
type FilesetMeta struct {
	ID             string              `json:"id"`
	Name           string              `json:"name"`
	DisplayPath    string              `json:"display_path"` // human-readable path hint for UI / source_identifier option
	SourceMetadata map[string]any      `json:"-"`            // flattened into Resource.SourceMetadata on discover
	Columns        []FilesetColumnMeta `json:"columns"`
}

// FilesetColumnMeta represents column metadata.
type FilesetColumnMeta struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	OrigType string `json:"orig_type"`
}

// TopicMeta represents message topic metadata.
type TopicMeta struct {
	Name       string `json:"name"`
	Partitions int    `json:"partitions"`
	Replicas   int    `json:"replicas"`
}

// MetricResult represents time-series query result.
type MetricResult struct {
	Metric string            `json:"metric"`
	Values []MetricValue     `json:"values"`
	Labels map[string]string `json:"labels"`
}

// MetricValue represents a single metric data point.
type MetricValue struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

// IndexMeta represents search index metadata.
type IndexMeta struct {
	Name       string                    `json:"name"`
	Properties map[string]any            `json:"properties"`
	Mapping    map[string]IndexFieldMeta `json:"mapping"`
}

// IndexFieldMeta represents index field metadata.
type IndexFieldMeta struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	OrigType   string                 `json:"orig_type"`
	Analyzer   string                 `json:"analyzer,omitempty"`
	Searchable bool                   `json:"searchable"`
	Features   map[string]interface{} `json:"features"`
}

// HealthStatus represents connection health status.
type HealthStatus struct {
	Status    string `json:"status"` // green, yellow, red, unknown
	Message   string `json:"message"`
	LatencyMs int64  `json:"latency_ms"`
}
