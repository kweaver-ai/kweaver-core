// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

// ResolvedMetricExecutionPayload is the merged internal representation before vega (DESIGN §3.1.2).
type ResolvedMetricExecutionPayload struct {
	MetricDefinition *MetricDefinition
	ObjectType       *ObjectType
	ResourceID       string
	VegaParams       *ResourceDataQueryParams
}

// MetricQueryService exposes BKN native metric query and dry-run (DESIGN §3.3).
//
//go:generate mockgen -source ../interfaces/metric_query_service.go -destination ../interfaces/mock/mock_metric_query_service.go
type MetricQueryService interface {
	// GetMetricDefinition loads metric definition by id (for handler-layer validation before ExecuteMetricQuery).
	GetMetricDefinition(ctx context.Context, knID string, branch string, metricID string) (*MetricDefinition, bool, error)
	// MetricQueryRequest.FillNull is set by the handler from URL query fill_null (json:"-").
	QueryMetricData(ctx context.Context, knID string, branch string, metricID string, body *MetricQueryRequest) (MetricData, error)
	DryRunMetricData(ctx context.Context, knID string, branch string, body *MetricDryRunRequest) (MetricData, error)
}
