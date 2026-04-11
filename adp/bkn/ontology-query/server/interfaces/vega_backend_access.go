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

// DatasetQueryParams is the JSON body for POST /resources/:id/data.
type DatasetQueryParams struct {
	FilterCondition map[string]any `json:"filter_condition,omitempty"`
	SearchAfter     []any          `json:"search_after,omitempty"`
	Offset          int            `json:"offset,omitempty"`
	Limit           int            `json:"limit,omitempty"`
	NeedTotal       bool           `json:"need_total,omitempty"`
	Sort            []*SortParams  `json:"sort,omitempty"`
	OutputFields    []string       `json:"output_fields,omitempty"`
}

// ResourceDataQueryParams is an alias for the same payload as dataset query (bkn-backend aligned).
type ResourceDataQueryParams = DatasetQueryParams

// VegaBackendAccess calls vega-backend resource data APIs (aligned with bkn-backend).
type VegaBackendAccess interface {
	QueryResourceData(ctx context.Context, resourceID string, params *ResourceDataQueryParams) (*DatasetQueryResponse, error)
}
