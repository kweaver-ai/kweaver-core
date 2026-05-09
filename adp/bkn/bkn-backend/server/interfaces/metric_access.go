// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"database/sql"
)

// MetricAccess persists and queries MetricDefinition rows (Task 3, IMPLEMENTATION_PLAN).
//
//go:generate mockgen -source ../interfaces/metric_access.go -destination ../interfaces/mock/mock_metric_access.go
type MetricAccess interface {
	CreateMetric(ctx context.Context, tx *sql.Tx, def *MetricDefinition) error
	GetMetricByID(ctx context.Context, knID, branch, metricID string) (*MetricDefinition, error)
	GetMetricsByIDs(ctx context.Context, knID, branch string, metricIDs []string) ([]*MetricDefinition, error)
	UpdateMetric(ctx context.Context, tx *sql.Tx, def *MetricDefinition) error
	DeleteMetricsByIDs(ctx context.Context, tx *sql.Tx, knID, branch string, metricIDs []string) error
	ListMetrics(ctx context.Context, query MetricsListQueryParams) ([]*MetricDefinition, error)
	GetMetricsTotal(ctx context.Context, query MetricsListQueryParams) (int, error)

	CheckMetricExistByID(ctx context.Context, knID, branch, metricID string) (string, bool, error)
	CheckMetricExistByName(ctx context.Context, knID, branch, name string) (string, bool, error)

	GetMetricIDsByKnID(ctx context.Context, knID, branch string) ([]string, error)
	DeleteMetricsByKnID(ctx context.Context, tx *sql.Tx, knID, branch string) (int64, error)
}
