// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package helpers provides common utilities for catalog tests (both physical and logical).
package helpers

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"vega-backend-tests/testutil"
)

const (
	CatalogTypeLogical  = "logical"
	CatalogTypePhysical = "physical"
)

// ========== 名称生成辅助函数 ==========

// GenerateUniqueName 生成唯一名称（带时间戳）
func GenerateUniqueName(prefix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d", prefix, timestamp)
}

// GenerateUniqueNameWithSuffix 生成唯一名称（带自定义后缀）
func GenerateUniqueNameWithSuffix(prefix, suffix string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d-%s", prefix, timestamp, suffix)
}

// ========== 字符串构建辅助函数 ==========

// BuildStringWithLength 构建指定长度的字符串
func BuildStringWithLength(char string, length int) string {
	return strings.Repeat(char, length)
}

// BuildUniqueStringWithLength 构建指定长度的唯一字符串（包含时间戳）
func BuildUniqueStringWithLength(length int) string {
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	if length <= len(timestamp) {
		return strings.Repeat("a", length)
	}
	padding := strings.Repeat("a", length-len(timestamp))
	return padding + timestamp
}

// ========== 深拷贝辅助函数 ==========

// DeepCopyMap 深拷贝map
func DeepCopyMap(original map[string]any) map[string]any {
	if original == nil {
		return nil
	}

	result := make(map[string]any)
	for key, value := range original {
		switch v := value.(type) {
		case map[string]any:
			result[key] = DeepCopyMap(v)
		case []any:
			result[key] = DeepCopySlice(v)
		default:
			result[key] = value
		}
	}
	return result
}

// DeepCopySlice 深拷贝slice
func DeepCopySlice(original []any) []any {
	if original == nil {
		return nil
	}

	result := make([]any, len(original))
	for i, value := range original {
		switch v := value.(type) {
		case map[string]any:
			result[i] = DeepCopyMap(v)
		case []any:
			result[i] = DeepCopySlice(v)
		default:
			result[i] = value
		}
	}
	return result
}

// ========== Update辅助函数 ==========

// catalogReadOnlyFields Catalog更新时需要移除的只读字段
var catalogReadOnlyFields = []string{
	"id", "type", "create_time", "update_time",
	"creator", "updater",
	"health_check_status", "health_check_time", "health_check_message",
}

// BuildUpdatePayload 基于现有catalog数据构建更新payload
func BuildUpdatePayload(original map[string]any, updates map[string]any) map[string]any {
	// 深拷贝原始数据
	payload := DeepCopyMap(original)

	// 移除只读字段
	for _, field := range catalogReadOnlyFields {
		delete(payload, field)
	}

	// 应用更新
	for key, value := range updates {
		if value == nil {
			delete(payload, key)
		} else {
			payload[key] = value
		}
	}

	return payload
}

// InjectEncryptedPassword 向payload的connector_config中注入加密后的password
func InjectEncryptedPassword(payload map[string]any, encryptedPassword string) {
	if connCfg, ok := payload["connector_config"].(map[string]any); ok {
		connCfg["password"] = encryptedPassword
	}
}

// ========== 清理辅助函数 ==========

// CleanupCatalogs 清理所有现有的catalog，确保测试环境干净
func CleanupCatalogs(client *testutil.HTTPClient, t *testing.T) {
	// 获取所有catalog（使用offset+limit分页）
	listResp := client.GET("/api/vega-backend/v1/catalogs?offset=0&limit=1000")
	if listResp.StatusCode != http.StatusOK {
		t.Logf("⚠ 获取catalog列表失败，状态码: %d", listResp.StatusCode)
		return
	}

	if listResp.Body == nil {
		t.Logf("✓ catalog列表为空，无需清理")
		return
	}

	entries, ok := listResp.Body["entries"].([]any)
	if !ok || len(entries) == 0 {
		t.Logf("✓ catalog列表为空，无需清理")
		return
	}

	t.Logf("⏳ 开始清理 %d 个现有catalog...", len(entries))

	deletedCount := 0
	for _, entry := range entries {
		catalogMap, ok := entry.(map[string]any)
		if !ok {
			continue
		}

		catalogID, ok := catalogMap["id"].(string)
		if !ok {
			continue
		}

		deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
		if deleteResp.StatusCode == http.StatusOK || deleteResp.StatusCode == http.StatusNoContent {
			deletedCount++
		} else {
			t.Logf("⚠ 删除catalog %s 失败，状态码: %d", catalogID, deleteResp.StatusCode)
		}
	}

	t.Logf("✓ 清理完成，成功删除 %d 个catalog", deletedCount)
}

// CleanupCatalogsByType 按类型清理catalog
func CleanupCatalogsByType(client *testutil.HTTPClient, t *testing.T, catalogType string) {
	listResp := client.GET("/api/vega-backend/v1/catalogs?offset=0&limit=1000&type=" + catalogType)
	if listResp.StatusCode != http.StatusOK {
		return
	}

	if listResp.Body == nil {
		return
	}

	entries, ok := listResp.Body["entries"].([]any)
	if !ok {
		return
	}

	deletedCount := 0
	for _, entry := range entries {
		catalog, ok := entry.(map[string]any)
		if !ok {
			continue
		}

		catalogID, ok := catalog["id"].(string)
		if !ok {
			continue
		}

		deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
		if deleteResp.StatusCode == http.StatusOK || deleteResp.StatusCode == http.StatusNoContent {
			deletedCount++
		}
	}

	if deletedCount > 0 {
		t.Logf("✓ 清理了 %d 个 %s catalog", deletedCount, catalogType)
	}
}

// CleanupResources 清理所有现有的resource，确保测试环境干净
func CleanupResources(client *testutil.HTTPClient, t *testing.T) {
	// 获取所有resource（使用offset+limit分页）
	listResp := client.GET("/api/vega-backend/v1/resources?offset=0&limit=1000")
	if listResp.StatusCode != http.StatusOK {
		t.Logf("⚠ 获取resource列表失败，状态码: %d", listResp.StatusCode)
		return
	}

	if listResp.Body == nil {
		t.Logf("✓ resource列表为空，无需清理")
		return
	}

	entries, ok := listResp.Body["entries"].([]any)
	if !ok || len(entries) == 0 {
		t.Logf("✓ resource列表为空，无需清理")
		return
	}

	t.Logf("⏳ 开始清理 %d 个现有resource...", len(entries))

	deletedCount := 0
	for _, entry := range entries {
		resourceMap, ok := entry.(map[string]any)
		if !ok {
			continue
		}

		resourceID, ok := resourceMap["id"].(string)
		if !ok {
			continue
		}

		deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
		if deleteResp.StatusCode == http.StatusOK || deleteResp.StatusCode == http.StatusNoContent {
			deletedCount++
		} else {
			t.Logf("⚠ 删除resource %s 失败，状态码: %d", resourceID, deleteResp.StatusCode)
		}
	}

	t.Logf("✓ 清理完成，成功删除 %d 个resource", deletedCount)
}

// ========== 响应提取辅助函数 ==========

// ExtractFromEntriesResponse 从响应中提取catalog对象
// 支持两种格式：
// - entries 格式：{"entries": [{...}]}（List 响应）
// - 直接返回：{"id": "...", "name": "..."}（Get 单个资源响应）
func ExtractFromEntriesResponse(resp testutil.HTTPResponse) map[string]any {
	if resp.Body == nil {
		return nil
	}

	// 尝试从entries数组中提取
	if entries, ok := resp.Body["entries"].([]any); ok && len(entries) > 0 {
		if item, ok := entries[0].(map[string]any); ok {
			return item
		}
	}

	// 直接返回格式（GET 单个资源），深拷贝避免修改原始数据
	if _, ok := resp.Body["id"]; ok {
		return DeepCopyMap(resp.Body)
	}

	return nil
}

// ExtractID 从创建响应中提取ID
func ExtractID(resp testutil.HTTPResponse) string {
	if resp.Body == nil {
		return ""
	}

	if id, ok := resp.Body["id"].(string); ok {
		return id
	}

	return ""
}

// WaitForCondition 等待条件满足（用于异步操作）
func WaitForCondition(condition func() bool, timeout time.Duration, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}

// ========== Logical Catalog 构建函数 ==========

// BuildLogicalCatalogPayload 构建logical catalog创建payload
func BuildLogicalCatalogPayload() map[string]any {
	return map[string]any{
		"name":        GenerateUniqueName("logical-catalog"),
		"type":        "logical",
		"description": "逻辑Catalog测试",
		"tags":        []string{"test", "logical"},
	}
}

// BuildLogicalCatalogPayloadWithName 构建带指定名称的logical catalog
func BuildLogicalCatalogPayloadWithName(name string) map[string]any {
	return map[string]any{
		"name":        name,
		"type":        "logical",
		"description": "逻辑Catalog测试",
		"tags":        []string{"test", "logical"},
	}
}

// BuildFullLogicalCatalogPayload 构建完整字段的logical catalog
func BuildFullLogicalCatalogPayload() map[string]any {
	return map[string]any{
		"name":        GenerateUniqueName("full-logical-catalog"),
		"type":        "logical",
		"description": "完整的逻辑Catalog测试，包含所有可选字段",
		"tags":        []string{"test", "logical", "full"},
	}
}
