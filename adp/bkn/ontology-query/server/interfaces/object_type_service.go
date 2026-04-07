// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

//go:generate mockgen -source ../interfaces/object_type_service.go -destination ../interfaces/mock/mock_object_type_service.go
type ObjectTypeService interface {
	GetObjectsByObjectTypeID(ctx context.Context, query *ObjectQueryBaseOnObjectType) (Objects, error)
	GetObjectPropertyValue(ctx context.Context, query *ObjectPropertyValueQuery) (Objects, error)
}
