// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import "context"

// ConnectorTypeService 定义 connector 类型业务逻辑接口
//
//go:generate mockgen -source ../interfaces/connector_type_service.go -destination ../interfaces/mock/mock_connector_type_service.go
type ConnectorTypeService interface {
	// Register 注册 connector 类型
	Register(ctx context.Context, ct *ConnectorTypeReq) error
	// Update 更新 connector 类型
	Update(ctx context.Context, ct *ConnectorTypeReq) error
	// Delete 删除 connector 类型
	DeleteByType(ctx context.Context, tp string) error
	// GetByType 根据类型获取 connector 类型
	GetByType(ctx context.Context, tp string) (*ConnectorType, error)
	// List 列出 connector 类型
	List(ctx context.Context, params ConnectorTypesQueryParams) ([]*ConnectorType, int64, error)
	// CheckExistByType 检查 connector 类型是否存在
	CheckExistByType(ctx context.Context, tp string) (bool, error)
	// CheckExistByName 检查 connector 类型名称是否存在
	CheckExistByName(ctx context.Context, name string) (bool, error)
	// SetEnabled 启用/禁用 connector 类型
	SetEnabled(ctx context.Context, tp string, enabled bool) error
}
