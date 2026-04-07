// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// ResourceService defines resource business logic interface.
//
//go:generate mockgen -source ../interfaces/resource_data_service.go -destination ../interfaces/mock/mock_resource_data_service.go
type ResourceDataService interface {
	// Query queries Resource data.
	Query(ctx context.Context, resource *Resource, params *ResourceDataQueryParams) ([]map[string]any, int64, error)
}
