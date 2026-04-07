// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package opensearch

import (
	"vega-backend-tests/at/catalog/helpers"
	"vega-backend-tests/at/setup"
)

// OpenSearchPayloadBuilder OpenSearch catalog payload构建器
type OpenSearchPayloadBuilder struct {
	config     setup.OpenSearchConfig
	testConfig *setup.TestConfig
}

// NewOpenSearchPayloadBuilder 创建OpenSearch payload构建器
func NewOpenSearchPayloadBuilder(config setup.OpenSearchConfig) *OpenSearchPayloadBuilder {
	return &OpenSearchPayloadBuilder{config: config}
}

// SetTestConfig 设置测试配置（包含加密器）
func (b *OpenSearchPayloadBuilder) SetTestConfig(tc *setup.TestConfig) {
	b.testConfig = tc
}

// encryptPassword 加密密码
func (b *OpenSearchPayloadBuilder) encryptPassword(password string) string {
	if b.testConfig != nil {
		return b.testConfig.EncryptString(password)
	}
	return password
}

// GetEncryptedPassword 返回加密后的正确密码
func (b *OpenSearchPayloadBuilder) GetEncryptedPassword() string {
	return b.encryptPassword(b.config.Password)
}

// GetConnectorType 返回connector类型
func (b *OpenSearchPayloadBuilder) GetConnectorType() string {
	return "opensearch"
}

// BuildCreatePayload 构建基本的OpenSearch catalog创建payload
func (b *OpenSearchPayloadBuilder) BuildCreatePayload() map[string]any {
	return map[string]any{
		"name":           helpers.GenerateUniqueName("test-opensearch-catalog"),
		"connector_type": "opensearch",
		"connector_config": map[string]any{
			"host":     b.config.Host,
			"port":     b.config.Port,
			"username": b.config.Username,
			"password": b.encryptPassword(b.config.Password),
			"use_ssl":  b.config.UseSSL,
		},
	}
}

// BuildFullCreatePayload 构建完整的OpenSearch catalog创建payload（包含所有可选字段）
func (b *OpenSearchPayloadBuilder) BuildFullCreatePayload() map[string]any {
	payload := b.BuildCreatePayload()
	payload["description"] = "完整的OpenSearch测试catalog"
	payload["tags"] = []string{"test", "opensearch", "at", "full-fields"}

	// 添加OpenSearch options
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["options"] = map[string]any{
		"timeout":           "30s",
		"max_retries":       3,
		"compress":          true,
		"discover_interval": "5m",
	}

	return payload
}

// BuildCreatePayloadWithOptions 构建包含options的OpenSearch catalog payload
func (b *OpenSearchPayloadBuilder) BuildCreatePayloadWithOptions(options map[string]any) map[string]any {
	payload := b.BuildCreatePayload()
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["options"] = options
	return payload
}

// BuildCreatePayloadWithWrongCredentials 构建错误凭证的OpenSearch catalog payload
func (b *OpenSearchPayloadBuilder) BuildCreatePayloadWithWrongCredentials() map[string]any {
	payload := b.BuildCreatePayload()
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["password"] = b.encryptPassword("wrong_password_123")
	return payload
}

// BuildCreatePayloadWithInvalidConfig 构建无效配置的OpenSearch catalog payload（无效的host）
func (b *OpenSearchPayloadBuilder) BuildCreatePayloadWithInvalidConfig() map[string]any {
	payload := b.BuildCreatePayload()
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["host"] = "invalid-host-does-not-exist.local"
	return payload
}

// SupportsTestConnection OpenSearch支持连接测试
func (b *OpenSearchPayloadBuilder) SupportsTestConnection() bool {
	return true
}

// GetRequiredConfigFields 返回OpenSearch connector_config必需的字段
func (b *OpenSearchPayloadBuilder) GetRequiredConfigFields() []string {
	return []string{"host", "port", "username", "password"}
}

// ========== OpenSearch特定Payload生成函数 ==========

// BuildCreatePayloadWithSSL 构建带SSL配置的OpenSearch catalog payload
func (b *OpenSearchPayloadBuilder) BuildCreatePayloadWithSSL(useSSL bool) map[string]any {
	payload := b.BuildCreatePayload()
	connectorConfig := payload["connector_config"].(map[string]any)
	connectorConfig["use_ssl"] = useSSL
	return payload
}

// BuildCreatePayloadWithInvalidPort 构建无效port的OpenSearch payload
func (b *OpenSearchPayloadBuilder) BuildCreatePayloadWithInvalidPort() map[string]any {
	return map[string]any{
		"name":           helpers.GenerateUniqueName("invalid-port-catalog"),
		"connector_type": "opensearch",
		"connector_config": map[string]any{
			"host":     b.config.Host,
			"port":     "not_a_number",
			"username": b.config.Username,
			"password": b.encryptPassword(b.config.Password),
			"use_ssl":  b.config.UseSSL,
		},
	}
}

// BuildCreatePayloadWithMissingAuth 构建缺少认证信息的OpenSearch payload
func (b *OpenSearchPayloadBuilder) BuildCreatePayloadWithMissingAuth() map[string]any {
	return map[string]any{
		"name":           helpers.GenerateUniqueName("missing-auth-catalog"),
		"connector_type": "opensearch",
		"connector_config": map[string]any{
			"host":    b.config.Host,
			"port":    b.config.Port,
			"use_ssl": b.config.UseSSL,
			// 缺少username和password
		},
	}
}

// BuildCreatePayloadWithInvalidHost 构建无效host的OpenSearch payload
func (b *OpenSearchPayloadBuilder) BuildCreatePayloadWithInvalidHost() map[string]any {
	return b.BuildCreatePayloadWithInvalidConfig()
}

// BuildCreatePayloadWithOutOfRangePort 构建超出范围port的OpenSearch payload
func (b *OpenSearchPayloadBuilder) BuildCreatePayloadWithOutOfRangePort() map[string]any {
	return map[string]any{
		"name":           helpers.GenerateUniqueName("out-of-range-port-catalog"),
		"connector_type": "opensearch",
		"connector_config": map[string]any{
			"host":     b.config.Host,
			"port":     99999, // 超出有效端口范围
			"username": b.config.Username,
			"password": b.encryptPassword(b.config.Password),
			"use_ssl":  b.config.UseSSL,
		},
	}
}

// GetConfig 返回OpenSearch配置（供测试中直接使用）
func (b *OpenSearchPayloadBuilder) GetConfig() setup.OpenSearchConfig {
	return b.config
}
