// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// ConnectorTypeAccess 定义 connector 类型数据访问接口
//
//go:generate mockgen -source ../interfaces/connector_type_access.go -destination ../interfaces/mock/mock_connector_type_access.go
type ConnectorTypeAccess interface {
	// Create 创建 connector 类型
	Create(ctx context.Context, ct *ConnectorType) error
	// Update 更新 connector 类型
	Update(ctx context.Context, ct *ConnectorType) error
	// Delete 删除 connector 类型
	DeleteByType(ctx context.Context, tp string) error
	// GetByType 根据类型获取 connector 类型
	GetByType(ctx context.Context, tp string) (*ConnectorType, error)
	// GetByName 根据名称获取 connector 类型
	GetByName(ctx context.Context, name string) (*ConnectorType, error)
	// List 列出 connector 类型
	List(ctx context.Context, params ConnectorTypesQueryParams) ([]*ConnectorType, int64, error)
	// SetEnabled 启用/禁用 connector 类型
	SetEnabled(ctx context.Context, tp string, enabled bool) error
}
