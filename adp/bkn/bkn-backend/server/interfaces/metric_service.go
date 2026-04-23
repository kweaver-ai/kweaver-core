// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

// MetricService exposes metric CRUD and concept search (Task 3, IMPLEMENTATION_PLAN).
//
//go:generate mockgen -source metric_service.go -destination mock/mock_metric_service.go
type MetricService interface {
	// CheckMetricExistByID 按 kn/branch/id 判断指标是否存在；存在时返回库中名称（与 ObjectTypeService.CheckObjectTypeExistByID 对齐，供 handler 提前校验）。
	CheckMetricExistByID(ctx context.Context, knID, branch, metricID string) (name string, exist bool, err error)
	// CheckMetricExistByName 按名称判断是否存在；存在时返回该指标的 id（与 CheckObjectTypeExistByName 对齐）。
	CheckMetricExistByName(ctx context.Context, knID, branch, name string) (metricID string, exist bool, err error)

	CreateMetrics(ctx context.Context, tx *sql.Tx, entries []*MetricDefinition, strictMode bool, importMode string) ([]string, error)
	ListMetrics(ctx context.Context, query MetricsListQueryParams) (*MetricsList, error)
	GetMetricByID(ctx context.Context, knID string, branch string, metricID string) (*MetricDefinition, error)
	GetMetricsByIDs(ctx context.Context, knID string, branch string, metricIDs []string) ([]*MetricDefinition, error)
	UpdateMetric(ctx context.Context, tx *sql.Tx, req *MetricDefinition, strictMode bool) error
	DeleteMetricsByIDs(ctx context.Context, tx *sql.Tx, knID string, branch string, metricIDs []string) error
	// DeleteMetricsByKnID 内部接口：按知识网络删除全部指标，不校验权限；tx 必须非 nil（与 DeleteActionTypesByKnID 一致）。
	DeleteMetricsByKnID(ctx context.Context, tx *sql.Tx, knID string, branch string) error

	SearchMetrics(ctx context.Context, query *ConceptsQuery) (MetricSearchResult, error)

	// InsertDatasetData 将指标写入 BKN 概念数据集（与 ObjectTypeService.InsertDatasetData 一致；供概念同步等无用户上下文场景调用）。
	InsertDatasetData(ctx context.Context, metrics []*MetricDefinition) error

	// ValidateMetrics 在 strictMode 下校验 scope 与对象类字段等（与 CreateMetrics 中严格校验一致），不写库；请求体合法性由 handler 先通过 ValidateMetricRequests 校验。
	ValidateMetrics(ctx context.Context, entries []*MetricDefinition, strictMode bool, importMode string) error
}
