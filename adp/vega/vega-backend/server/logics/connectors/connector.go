// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package connectors defines the port interfaces for hexagonal architecture.
package connectors

import (
	"context"

	"vega-backend/interfaces"
)

// Connector 定义基础连接器接口
type Connector interface {
	GetType() string
	GetName() string
	GetMode() string
	GetCategory() string

	GetEnabled() bool
	SetEnabled(bool)

	// GetSensitiveFields 返回该 connector 的敏感字段列表（如 password）
	GetSensitiveFields() []string
	// GetFieldConfig 返回该 connector 的字段配置定义（兼容 JSON Schema properties）
	GetFieldConfig() map[string]interfaces.ConnectorFieldConfig

	New(cfg interfaces.ConnectorConfig) (Connector, error)

	Connect(ctx context.Context) error
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	TestConnection(ctx context.Context) error

	GetMetadata(ctx context.Context) (map[string]any, error)
}

// LocalConnectorBuilder 本地 connector 构建函数
type LocalConnectorBuilder func(cfg *interfaces.ConnectorConfig) (Connector, error)

// TableConnector defines the interface for relational database connectors.
// Implementations: mysql, postgresql, dameng, oracle, clickhouse, etc.
type TableConnector interface {
	Connector

	// ListDatabases 列出实例下所有可访问的数据库
	ListDatabases(ctx context.Context) ([]string, error)
	ListTables(ctx context.Context) ([]*interfaces.TableMeta, error)
	GetTableMeta(ctx context.Context, table *interfaces.TableMeta) error

	// ExecuteQuery 执行单表查询语句
	ExecuteQuery(ctx context.Context, resource *interfaces.Resource,
		params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error)

	// ExecuteJoinQuery 执行多表 JOIN 查询；单表时 resources 长度 1、joins 为空
	ExecuteJoinQuery(ctx context.Context, catalog *interfaces.Catalog, params *interfaces.JoinQueryParams) (*interfaces.QueryResult, error)
}

// FileConnector defines the interface for file/document storage connectors.
// Implementations: s3, hdfs, minio, feishu, notion, etc.
type FileConnector interface {
	Connector
}

// FilesetConnector defines the interface for file/document storage connectors.
// Implementations: anyshare, s3, hdfs, minio, feishu, notion, etc.
type FilesetConnector interface {
	Connector
	// ListFilesets lists file and folder objects for discovery (typically one level per parent).
	ListFilesets(ctx context.Context) ([]*interfaces.FilesetMeta, error)
	// ExecuteQuery executes a query on the fileset
	ExecuteQuery(ctx context.Context, resource *interfaces.Resource, params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error)
}

// TopicConnector defines the interface for message queue connectors.
// Implementations: kafka, pulsar, etc.
type TopicConnector interface {
	Connector
}

// MetricConnector defines the interface for time-series database connectors.
// Implementations: prometheus, influxdb, etc.
type MetricConnector interface {
	Connector
}

// IndexConnector defines the interface for search engine connectors.
// Implementations: opensearch, elasticsearch, etc.
type IndexConnector interface {
	Connector

	ListIndexes(ctx context.Context) ([]*interfaces.IndexMeta, error)
	GetIndexMeta(ctx context.Context, index *interfaces.IndexMeta) error

	// ExecuteQuery executes a query on the index
	ExecuteQuery(ctx context.Context, indexName string, resource *interfaces.Resource, params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error)
	ExecuteQueryWithDsl(ctx context.Context, resourceName string, dsl string) (*interfaces.QueryResult, error)
	ExecuteRawQuery(ctx context.Context, index string, query map[string]any) (*interfaces.SQLQueryResponse, error)
	// for dataset
	Create(ctx context.Context, name string, schemaDefinition []*interfaces.Property) error
	Update(ctx context.Context, name string, schemaDefinition []*interfaces.Property) error
	Delete(ctx context.Context, name string) error
	CheckExist(ctx context.Context, name string) (bool, error)
	CreateDocuments(ctx context.Context, name string, documents []map[string]any) ([]string, error)
	GetDocument(ctx context.Context, name string, docID string) (map[string]any, error)
	DeleteDocument(ctx context.Context, name string, docID string) error
	UpsertDocuments(ctx context.Context, name string, updateRequests []map[string]any) ([]string, error)
	DeleteDocuments(ctx context.Context, name string, docIDs string) error
	DeleteDocumentsByQuery(ctx context.Context, name string, params *interfaces.ResourceDataQueryParams, schemaDefinition []*interfaces.Property) error
}

// APIConnector defines the interface for REST/GraphQL API connectors.
// Implementations: rest, graphql, etc.
type APIConnector interface {
	Connector
}

// MariaDBSQLExecutor defines the interface for executing raw SQL on MariaDB
type MariaDBSQLExecutor interface {
	ExecuteRawSQL(ctx context.Context, sql string) (*interfaces.SQLQueryResponse, error)
}
