// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package helpers

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"vega-backend-tests/testutil"
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

// resourceReadOnlyFields Resource更新时需要移除的只读字段
var resourceReadOnlyFields = []string{
	"id", "catalog_id", "create_time", "update_time",
	"creator", "updater",
}

// BuildUpdatePayload 基于现有resource数据构建更新payload
func BuildUpdatePayload(original map[string]any, updates map[string]any) map[string]any {
	// 深拷贝原始数据
	payload := DeepCopyMap(original)

	// 移除只读字段
	for _, field := range resourceReadOnlyFields {
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

// ========== 清理辅助函数 ==========

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

// ========== 响应提取辅助函数 ==========

// ExtractFromEntriesResponse 从entries格式响应中提取第一个对象
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

	// 如果响应直接就是对象（向后兼容）
	if _, hasID := resp.Body["id"]; hasID {
		return resp.Body
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
