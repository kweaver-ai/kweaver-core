// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
)

// DatasetQueryResponse matches bkn-backend / vega resource data response shape.
type DatasetQueryResponse struct {
	Entries     []map[string]any `json:"entries"`
	TotalCount  int64            `json:"total_count"`
	SearchAfter []any            `json:"search_after"`
}

// ResourceDataQueryParams is the JSON body for POST /resources/:id/data.
// Analytics fields align with resource_data_query_analytics_schema.md (aggregate mode).
type ResourceDataQueryParams struct {
	FilterCondition map[string]any `json:"filter_condition,omitempty"`
	SearchAfter     []any          `json:"search_after,omitempty"`
	Offset          int            `json:"offset,omitempty"`
	Limit           int            `json:"limit,omitempty"`
	NeedTotal       bool           `json:"need_total,omitempty"`
	Sort            []*SortParams  `json:"sort,omitempty"`
	OutputFields    []string       `json:"output_fields,omitempty"`

	Aggregation map[string]any   `json:"aggregation,omitempty"`
	GroupBy     []map[string]any `json:"group_by,omitempty"`
	// OrderBy     []map[string]any `json:"order_by,omitempty"` // 同 sort
	Having map[string]any `json:"having,omitempty"`
}

//go:generate mockgen -source vega_backend_access.go -destination mock/mock_vega_backend_access.go -package mock_interfaces
type VegaBackendAccess interface {
	QueryResourceData(ctx context.Context, resourceID string, params *ResourceDataQueryParams) (*DatasetQueryResponse, error)
}
