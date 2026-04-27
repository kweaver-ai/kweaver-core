// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package helpers

import (
	"vega-backend-tests/at/setup"
)

// ========== MariaDB Catalog Builder ==========

// MariaDBCatalogBuilder MariaDB Catalog payload构建器
type MariaDBCatalogBuilder struct {
	config        setup.MariaDBConfig
	testConfig    *setup.TestConfig
	connectorType string
}

// NewMariaDBCatalogBuilder 创建MariaDB Catalog构建器
func NewMariaDBCatalogBuilder(config setup.MariaDBConfig) *MariaDBCatalogBuilder {
	return &MariaDBCatalogBuilder{
		config:        config,
		connectorType: "mariadb",
	}
}

// SetTestConfig 设置测试配置（用于密码加密等）
func (b *MariaDBCatalogBuilder) SetTestConfig(config *setup.TestConfig) {
	b.testConfig = config
}

// GetConnectorType 返回connector类型
func (b *MariaDBCatalogBuilder) GetConnectorType() string {
	return b.connectorType
}

// GetEncryptedPassword 获取加密后的密码
func (b *MariaDBCatalogBuilder) GetEncryptedPassword() string {
	if b.testConfig != nil {
		return b.testConfig.EncryptString(b.config.Password)
	}
	return b.config.Password
}

// BuildCreatePayload 构建MariaDB catalog创建payload
func (b *MariaDBCatalogBuilder) BuildCreatePayload() map[string]any {
	return map[string]any{
		"name":           GenerateUniqueName("test-mariadb-catalog"),
		"connector_type": b.connectorType,
		"connector_config": map[string]any{
			"host":     b.config.Host,
			"port":     b.config.Port,
			"username": b.config.Username,
			"password": b.GetEncryptedPassword(),
			"database": b.config.Database,
		},
	}
}

// ========== OpenSearch Catalog Builder ==========

// OpenSearchCatalogBuilder OpenSearch Catalog payload构建器
type OpenSearchCatalogBuilder struct {
	config        setup.OpenSearchConfig
	testConfig    *setup.TestConfig
	connectorType string
}

// NewOpenSearchCatalogBuilder 创建OpenSearch Catalog构建器
func NewOpenSearchCatalogBuilder(config setup.OpenSearchConfig) *OpenSearchCatalogBuilder {
	return &OpenSearchCatalogBuilder{
		config:        config,
		connectorType: "opensearch",
	}
}

// SetTestConfig 设置测试配置（用于密码加密等）
func (b *OpenSearchCatalogBuilder) SetTestConfig(config *setup.TestConfig) {
	b.testConfig = config
}

// GetConnectorType 返回connector类型
func (b *OpenSearchCatalogBuilder) GetConnectorType() string {
	return b.connectorType
}

// GetEncryptedPassword 获取加密后的密码
func (b *OpenSearchCatalogBuilder) GetEncryptedPassword() string {
	if b.testConfig != nil {
		return b.testConfig.EncryptString(b.config.Password)
	}
	return b.config.Password
}

// BuildCreatePayload 构建OpenSearch catalog创建payload
func (b *OpenSearchCatalogBuilder) BuildCreatePayload() map[string]any {
	return map[string]any{
		"name":           GenerateUniqueName("test-opensearch-catalog"),
		"connector_type": b.connectorType,
		"connector_config": map[string]any{
			"host":     b.config.Host,
			"port":     b.config.Port,
			"username": b.config.Username,
			"password": b.GetEncryptedPassword(),
			"use_ssl":  b.config.UseSSL,
		},
	}
}

// ========== Factory Function ==========

// NewCatalogPayloadBuilder 根据connector类型创建对应的CatalogPayloadBuilder
func NewCatalogPayloadBuilder(connectorType string, config *setup.TestConfig) CatalogPayloadBuilder {
	switch connectorType {
	case "mariadb":
		b := NewMariaDBCatalogBuilder(config.TargetMariaDB)
		b.SetTestConfig(config)
		return b
	case "opensearch":
		b := NewOpenSearchCatalogBuilder(config.TargetOpenSearch)
		b.SetTestConfig(config)
		return b
	default:
		return nil
	}
}
