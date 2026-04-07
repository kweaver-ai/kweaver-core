// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package remote provides HTTP-based remote connector implementations.
package remote

import (
	"context"

	"vega-backend/interfaces"
	"vega-backend/logics/connectors"
)

// ============================================
// RemoteConnector 基础远程连接器
// ============================================

// RemoteConnector 实现基础的远程连接器代理
type RemoteConnector struct {
	enabled  bool
	connType *interfaces.ConnectorType
	config   interfaces.ConnectorConfig
}

// NewRemoteConnector 创建基础远程连接器
func NewRemoteConnector(ct *interfaces.ConnectorType) *RemoteConnector {
	return &RemoteConnector{
		enabled:  ct.Enabled,
		connType: ct,
	}
}

// GetType 返回连接器类型
func (r *RemoteConnector) GetType() string {
	return r.connType.Type
}

// GetName 返回连接器名称
func (r *RemoteConnector) GetName() string {
	return r.connType.Name
}

// GetMode 返回连接器模式
func (r *RemoteConnector) GetMode() string {
	return r.connType.Mode
}

// GetCategory 返回连接器类别
func (r *RemoteConnector) GetCategory() string {
	return r.connType.Category
}

// GetEnabled 返回连接器是否启用
func (r *RemoteConnector) GetEnabled() bool {
	return r.enabled
}

func (rc *RemoteConnector) SetEnabled(enabled bool) {
	rc.enabled = enabled
}

// GetSensitiveFields returns the sensitive fields for this remote connector.
// Falls back to default ["password"] if not configured.
func (rc *RemoteConnector) GetSensitiveFields() []string {
	return []string{"password"}
}

// GetFieldConfig returns the field configuration for this remote connector.
// 从 ConnectorType 中获取字段配置
func (rc *RemoteConnector) GetFieldConfig() map[string]interfaces.ConnectorFieldConfig {
	return rc.connType.FieldConfig
}

// New 创建新的连接器实例
func (rc *RemoteConnector) New(cfg interfaces.ConnectorConfig) (connectors.Connector, error) {
	return &RemoteConnector{
		enabled:  rc.enabled,
		connType: rc.connType,
		config:   cfg,
	}, nil
}

func (rc *RemoteConnector) Connect(ctx context.Context) error {
	return nil
}

func (rc *RemoteConnector) Close(ctx context.Context) error {
	return nil
}

func (rc *RemoteConnector) Ping(ctx context.Context) error {
	return nil
}

func (rc *RemoteConnector) TestConnection(ctx context.Context) error {
	return nil
}

// GetMetadata returns the metadata for the catalog (stub).
func (rc *RemoteConnector) GetMetadata(ctx context.Context) (map[string]any, error) {
	return nil, nil
}
