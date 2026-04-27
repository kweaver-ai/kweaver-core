// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package dataset

import (
	"fmt"
	"math/rand"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend-tests/at/setup"
	"vega-backend-tests/testutil"
)

// ========== 辅助函数 ==========

// generateUniqueName 生成唯一名称
func generateUniqueName(prefix string) string {
	suffix := rand.Intn(10000)
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().Unix(), suffix)
}

// buildDatasetResourcePayload 构建dataset资源payload
func buildDatasetResourcePayload() map[string]any {
	return map[string]any{
		"catalog_id":        "default",
		"name":              generateUniqueName("test-dataset-full"),
		"tags":              []string{"test", "dataset"},
		"description":       "测试数据集资源",
		"category":          "dataset",
		"status":            "active",
		"source_identifier": "at_db",
		"schema_definition": []map[string]any{
			{"name": "id", "type": "keyword", "display_name": "ID", "original_name": "id", "description": "唯一标识符"},
			{"name": "@timestamp", "type": "long", "display_name": "时间戳", "original_name": "@timestamp", "description": "事件发生时间"},
			{"name": "name", "type": "text", "display_name": "名称", "original_name": "name", "description": "用户名称", "features": []map[string]any{
				{
					"name":         "keyword_name",
					"display_name": "关键词名称",
					"feature_type": "keyword",
					"description":  "用户名称的关键词表示",
					"ref_property": "name",
					"is_default":   true,
					"is_native":    true,
					"config": map[string]any{
						"ignore_above": 1024,
					},
				},
				{
					"name":         "fulltext_name",
					"display_name": "全文名称",
					"feature_type": "fulltext",
					"description":  "用户名称的全文表示",
					"ref_property": "name",
					"is_default":   true,
					"is_native":    true,
					"config": map[string]any{
						"analyzer": "standard",
					},
				},
			}},
			{"name": "age", "type": "integer", "display_name": "年龄", "original_name": "age", "description": "用户年龄"},
			{"name": "email", "type": "text", "display_name": "邮箱", "original_name": "email", "description": "用户邮箱"},
			{"name": "active", "type": "boolean", "display_name": "是否激活", "original_name": "active", "description": "用户是否激活"},
			{"name": "tags", "type": "string", "display_name": "标签", "original_name": "tags", "description": "用户标签"},
			{"name": "content", "type": "vector", "display_name": "内容向量", "original_name": "content", "description": "用户内容向量", "features": []map[string]any{
				{
					"name":         "content",
					"display_name": "内容向量",
					"feature_type": "vector",
					"description":  "用户内容向量",
					"ref_property": "content",
					"is_default":   true,
					"is_native":    true,
					"config": map[string]any{
						"dimension": 768,
						"method": map[string]any{
							"name":   "hnsw",
							"engine": "lucene",
							"parameters": map[string]any{
								"ef_construction": 256,
							},
						},
					},
				},
			}},
		},
	}
}

// buildDatasetResourcePayloadWithName 构建指定名称的dataset资源payload
func buildDatasetResourcePayloadWithName(name string) map[string]any {
	payload := buildDatasetResourcePayload()
	payload["name"] = name
	return payload
}

// extractFromEntriesResponse 从响应中提取资源数据
func extractFromEntriesResponse(resp testutil.HTTPResponse) map[string]any {
	if resp.Body != nil {
		if entries, ok := resp.Body["entries"].([]any); ok {
			if len(entries) > 0 {
				if entry, ok := entries[0].(map[string]any); ok {
					return entry
				}
			}
		}
	}
	return nil
}

// buildUpdatePayload 构建更新payload
func buildUpdatePayload(originalData map[string]any, updates map[string]any) map[string]any {
	// 基于原始数据创建更新payload
	payload := make(map[string]any)
	for k, v := range originalData {
		payload[k] = v
	}

	// 应用更新
	for k, v := range updates {
		payload[k] = v
	}

	return payload
}

// buildDatasetDocumentPayload 构建dataset文档payload
func buildDatasetDocumentPayload() map[string]any {
	return map[string]any{
		"@timestamp": int(time.Now().UnixMilli()),
		"name":       "Test User",
		"age":        30,
		"content":    generateVector(768),
	}
}

// generateVector 生成指定维度的向量
func generateVector(dims int) []float64 {
	vector := make([]float64, dims)
	for i := range vector {
		vector[i] = rand.Float64()*2 - 1 // 生成 [-1, 1] 范围内的随机数
	}
	return vector
}

// createTestCatalog 创建测试用的catalog并返回其ID，同时设置清理函数
func createTestCatalog(client *testutil.HTTPClient, t *testing.T) string {
	// 创建测试用的catalog
	catalogPayload := map[string]any{
		"name":        generateUniqueName("test-dataset-catalog"),
		"description": "测试dataset catalog",
		"tags":        []string{"test", "dataset", "catalog"},
		"type":        "mariadb",
		"connection_info": map[string]any{
			"host":     "localhost",
			"port":     3306,
			"user":     "root",
			"password": "password",
			"database": "test",
		},
	}
	catalogResp := client.POST("/api/vega-backend/v1/catalogs", catalogPayload)
	So(catalogResp.StatusCode, ShouldEqual, http.StatusCreated)
	So(catalogResp.Body["id"], ShouldNotBeEmpty)
	catalogID := catalogResp.Body["id"].(string)
	return catalogID
}

// deleteTestResourceAndCatalog 清理测试用的resource和catalog
func deleteTestResourceAndCatalog(client *testutil.HTTPClient, t *testing.T, resourceIDs []string, catalogID string) {
	t.Logf("清理资源IDs: %v, catalog: %s", resourceIDs, catalogID)
	// 先删除resourceIDs中的所有资源
	for _, resourceID := range resourceIDs {
		deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
		if deleteResp.StatusCode != http.StatusOK && deleteResp.StatusCode != http.StatusNoContent {
			t.Logf("清理资源失败 %s: %d, 响应体: %v", resourceID, deleteResp.StatusCode, deleteResp.Body)
		}
	}

	// 再删除catalog
	deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
	if deleteResp.StatusCode != http.StatusOK && deleteResp.StatusCode != http.StatusNoContent {
		t.Logf("清理catalog失败 %s: %d, 响应体: %v", catalogID, deleteResp.StatusCode, deleteResp.Body)
	}
}

// TestDatasetResourceCreateAndQuery Dataset资源创建和查询AT测试
// 测试编号前缀: DS1xx (Dataset Create and Query)
func TestDatasetResourceCreateAndQuery(t *testing.T) {
	Convey("Dataset资源创建AT测试 - 初始化", t, func() {
		// 加载测试配置
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		// 创建HTTP客户端
		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)

		// 验证服务可用性
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		// 创建测试用的catalog
		catalogID := createTestCatalog(client, t)
		resourceIDs := []string{}

		// ========== 正向测试（DS101-DS120） ==========

		Convey("DS101: 创建dataset资源", func() {
			payload := buildDatasetResourcePayload()
			payload["catalog_id"] = catalogID
			resp := client.POST("/api/vega-backend/v1/resources", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
			resourceIDs = append(resourceIDs, resp.Body["id"].(string))
		})

		Convey("DS102: 创建后立即查询", func() {
			payload := buildDatasetResourcePayload()
			payload["catalog_id"] = catalogID
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			resourceID := createResp.Body["id"].(string)
			resourceIDs = append(resourceIDs, resourceID)

			// 立即查询
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)
			resource := extractFromEntriesResponse(getResp)
			So(resource, ShouldNotBeNil)
			So(resource["category"], ShouldEqual, "dataset")
			So(resource["id"], ShouldEqual, resourceID)
			So(resource["name"], ShouldEqual, payload["name"])
		})

		Convey("DS103: 获取不存在的resource", func() {
			resp := client.GET("/api/vega-backend/v1/resources/non-existent-id-12345")
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("DS104: 列表查询 - 按category过滤dataset", func() {
			// 创建1个dataset resource
			payload := buildDatasetResourcePayload()
			payload["catalog_id"] = catalogID
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			resourceIDs = append(resourceIDs, createResp.Body["id"].(string))

			// 查询dataset类型
			datasetResp := client.GET("/api/vega-backend/v1/resources?category=dataset&offset=0&limit=10")
			So(datasetResp.StatusCode, ShouldEqual, http.StatusOK)

			if datasetResp.Body != nil && datasetResp.Body["entries"] != nil {
				entries := datasetResp.Body["entries"].([]any)
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)
			}
		})

		Convey("DS105: 创建dataset资源 - 包含object类型字段", func() {
			// 构建带有object类型字段的payload
			payload := map[string]any{
				"catalog_id":        catalogID,
				"name":              generateUniqueName("test-dataset-object-field"),
				"category":          "dataset",
				"connector_type":    "mariadb",
				"description":       "测试包含object类型字段的数据集",
				"tags":              []string{"test", "dataset", "object"},
				"source_identifier": "at_db",
				"schema_definition": []map[string]any{
					{"name": "id", "type": "keyword"},
					{"name": "name", "type": "text"},
					{"name": "user_info", "type": "object"},
				},
			}

			// 创建资源
			resp := client.POST("/api/vega-backend/v1/resources", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
			resourceIDs = append(resourceIDs, resp.Body["id"].(string))

			// 验证创建的资源
			resourceID := resp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			resource := extractFromEntriesResponse(getResp)
			So(resource, ShouldNotBeNil)
			So(resource["id"], ShouldEqual, resourceID)
			So(resource["name"], ShouldEqual, payload["name"])

			// 写入包含object类型数据的文档
			docPayload := map[string]any{
				"id":   "obj-123",
				"name": "Test Object User",
				"user_info": map[string]any{
					"age":   30,
					"email": "test@example.com",
					"address": map[string]any{
						"city": "New York",
						"zip":  "10001",
					},
				},
			}
			createDocResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", []map[string]any{docPayload})
			So(createDocResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(createDocResp.Body["ids"], ShouldNotBeEmpty)

			// 验证文档创建成功
			ids, ok := createDocResp.Body["ids"].([]interface{})
			So(ok, ShouldBeTrue)
			So(len(ids), ShouldBeGreaterThan, 0)
		})

		Convey("DS106: 创建dataset资源 - object数据类型，自定定义是嵌套，非 object", func() {
			payload := map[string]any{
				"catalog_id":        catalogID,
				"name":              generateUniqueName("test-dataset-specific-schema"),
				"category":          "dataset",
				"source_identifier": "at_db",
				"schema_definition": []map[string]any{
					{"name": "id", "type": "keyword"},
					{"name": "name", "type": "text"},
					{"name": "address.a", "type": "text"},
					{"name": "address.b", "type": "text"},
				},
			}

			// 创建资源
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(createResp.Body["id"], ShouldNotBeEmpty)
			resourceIDs = append(resourceIDs, createResp.Body["id"].(string))

			// 验证创建的资源
			resourceID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			resource := extractFromEntriesResponse(getResp)
			So(resource, ShouldNotBeNil)
			So(resource["id"], ShouldEqual, resourceID)

			// 写入文档
			docPayload := map[string]any{
				"id":   "123456",
				"name": "Test User",
				"address": map[string]any{
					"a": "123 Main St",
					"b": "Anytown, USA",
				},
			}
			createDocResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", []map[string]any{docPayload})
			So(createDocResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(createDocResp.Body["ids"], ShouldNotBeEmpty)
			ids, ok := createDocResp.Body["ids"].([]interface{})
			So(ok, ShouldBeTrue)
			So(len(ids), ShouldBeGreaterThan, 0)
			//docID := ids[0].(string)

			// 通过 keyword 类型的 id 查询验证写入的文档
			client.SetHeader("X-HTTP-Method-Override", "GET")
			getDocResp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", map[string]any{
				"offset": 0,
				"limit":  10,
				"filter_condition": map[string]any{
					"operation": "eq",
					"field":     "id",
					"value":     "123456",
				},
				"output_fields": []string{"id", "name", "address"},
				"need_total":    false,
			})
			client.RemoveHeader("X-HTTP-Method-Override")
			So(getDocResp.StatusCode, ShouldEqual, http.StatusOK)
			So(getDocResp.Body["entries"], ShouldNotBeNil)
			entries := getDocResp.Body["entries"].([]any)
			So(len(entries), ShouldBeGreaterThan, 0)
			doc := entries[0].(map[string]any)
			So(doc["id"], ShouldEqual, "123456")
			So(doc["name"], ShouldEqual, "Test User")
			So(doc["address"], ShouldNotBeEmpty)
			address := doc["address"].(map[string]any)
			So(address["a"], ShouldEqual, "123 Main St")
			So(address["b"], ShouldEqual, "Anytown, USA")

			// 通过 text 类型的 name 查询验证写入的文档
			client.SetHeader("X-HTTP-Method-Override", "GET")
			getDocByNameResp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", map[string]any{
				"offset": 0,
				"limit":  10,
				"filter_condition": map[string]any{
					"operation": "match",
					"field":     "name",
					"value":     "Test User",
				},
				"output_fields": []string{"_id", "name", "address"},
				"need_total":    false,
			})
			client.RemoveHeader("X-HTTP-Method-Override")
			So(getDocByNameResp.StatusCode, ShouldEqual, http.StatusOK)
			So(getDocByNameResp.Body["entries"], ShouldNotBeNil)
			entriesByName := getDocByNameResp.Body["entries"].([]any)
			So(len(entriesByName), ShouldBeGreaterThan, 0)
			docByName := entriesByName[0].(map[string]any)
			So(docByName["name"], ShouldEqual, "Test User")
		})

		Convey("DS107: 指定ID创建dataset资源", func() {
			payload := buildDatasetResourcePayload()
			payload["catalog_id"] = catalogID
			specificID := generateUniqueName("test-dataset-specific-id")
			payload["id"] = specificID
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			resourceIDs = append(resourceIDs, createResp.Body["id"].(string))

			// 立即查询
			getResp := client.GET("/api/vega-backend/v1/resources/" + specificID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)
			resource := extractFromEntriesResponse(getResp)
			So(resource, ShouldNotBeNil)
		})

		Convey("DS108: 测试 VEGA 类型与 OpenSearch 类型的转换", func() {
			// 构建包含各种类型的 dataset 资源
			payload := map[string]any{
				"catalog_id":        catalogID,
				"name":              generateUniqueName("test-dataset-types"),
				"tags":              []string{"test", "dataset", "types"},
				"description":       "测试各种类型的数据集",
				"category":          "dataset",
				"status":            "active",
				"source_identifier": "at_db",
				"schema_definition": []map[string]any{
					{"name": "id", "type": "integer", "display_name": "整数ID", "original_name": "id", "description": "整数类型"},
					{"name": "uid", "type": "unsigned_integer", "display_name": "无符号整数", "original_name": "uid", "description": "无符号整数类型"},
					{"name": "score", "type": "float", "display_name": "浮点数", "original_name": "score", "description": "浮点数类型"},
					{"name": "price", "type": "decimal", "display_name": "小数", "original_name": "price", "description": "小数类型"},
					{"name": "name", "type": "string", "display_name": "字符串", "original_name": "name", "description": "字符串类型"},
					{"name": "created_at", "type": "datetime", "display_name": "日期时间", "original_name": "created_at", "description": "日期时间类型"},
					{"name": "event_time", "type": "time", "display_name": "时间", "original_name": "event_time", "description": "时间类型"},
					{"name": "metadata", "type": "json", "display_name": "JSON", "original_name": "metadata", "description": "JSON类型"},
					{"name": "location", "type": "point", "display_name": "地理位置", "original_name": "location", "description": "地理位置点类型"},
					{"name": "area", "type": "shape", "display_name": "地理形状", "original_name": "area", "description": "地理形状类型"},
					{"name": "embedding", "type": "vector", "display_name": "向量", "original_name": "embedding", "description": "向量类型", "features": []map[string]any{
						{
							"name":         "embedding",
							"display_name": "向量",
							"feature_type": "vector",
							"description":  "向量类型",
							"ref_property": "embedding",
							"is_default":   true,
							"is_native":    true,
							"config": map[string]any{
								"dimension": 768,
								"method": map[string]any{
									"name":   "hnsw",
									"engine": "lucene",
									"parameters": map[string]any{
										"ef_construction": 256,
									},
								},
							},
						},
					}},
				},
			}

			// 创建 dataset 资源
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			resourceIDs = append(resourceIDs, createResp.Body["id"].(string))
		})

		// ========== 反向测试（DS121-DS127） ==========

		Convey("DS121: 重复的resource名称", func() {
			fixedName := generateUniqueName("duplicate-dataset")
			payload1 := buildDatasetResourcePayloadWithName(fixedName)
			payload1["catalog_id"] = catalogID

			// 第一次创建
			resp1 := client.POST("/api/vega-backend/v1/resources", payload1)
			So(resp1.StatusCode, ShouldEqual, http.StatusCreated)
			resourceIDs = append(resourceIDs, resp1.Body["id"].(string))

			// 第二次创建相同名称
			payload2 := buildDatasetResourcePayloadWithName(fixedName)
			payload2["catalog_id"] = catalogID
			resp2 := client.POST("/api/vega-backend/v1/resources", payload2)
			So(resp2.StatusCode, ShouldEqual, http.StatusConflict)
		})

		Convey("DS122: 缺少必填字段 - name", func() {
			payload := map[string]any{
				"catalog_id":     catalogID,
				"category":       "dataset",
				"connector_type": "mariadb",
			}
			resp := client.POST("/api/vega-backend/v1/resources", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		// 清理资源
		deleteTestResourceAndCatalog(client, t, resourceIDs, catalogID)
	})
}

// TestDatasetResourceUpdate Dataset资源更新AT测试
// 测试编号前缀: DS3xx
func TestDatasetResourceUpdate(t *testing.T) {
	Convey("Dataset资源更新AT测试 - 初始化", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 创建测试用的catalog
		catalogID := createTestCatalog(client, t)
		resourceIDs := []string{}

		// ========== 更新测试（DS301-DS310） ==========

		Convey("DS301: 更新dataset资源名称", func() {
			// 创建
			payload := buildDatasetResourcePayload()
			payload["catalog_id"] = catalogID
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			resourceID := createResp.Body["id"].(string)
			resourceIDs = append(resourceIDs, resourceID)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			resourceData := extractFromEntriesResponse(getResp)

			// 基于原数据构建更新payload
			newName := generateUniqueName("updated-dataset")
			updatePayload := buildUpdatePayload(resourceData, map[string]any{
				"name": newName,
			})
			updateResp := client.PUT("/api/vega-backend/v1/resources/"+resourceID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证
			verifyResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			resource := extractFromEntriesResponse(verifyResp)
			So(resource["name"], ShouldEqual, newName)
		})

		Convey("DS302: 更新dataset资源schema", func() {
			// 创建
			payload := buildDatasetResourcePayload()
			payload["catalog_id"] = catalogID
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			resourceID := createResp.Body["id"].(string)
			resourceIDs = append(resourceIDs, resourceID)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			resourceData := extractFromEntriesResponse(getResp)

			// 更新schema
			newSchema := []map[string]any{
				{"name": "address", "type": "text"},
			}
			updatePayload := buildUpdatePayload(resourceData, map[string]any{
				"schema_definition": newSchema,
			})
			updateResp := client.PUT("/api/vega-backend/v1/resources/"+resourceID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})

		// 清理资源
		deleteTestResourceAndCatalog(client, t, resourceIDs, catalogID)
	})
}

// TestDatasetResourceDelete Dataset资源删除AT测试
// 测试编号前缀: DS4xx
func TestDatasetResourceDelete(t *testing.T) {
	Convey("Dataset资源删除AT测试 - 初始化", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 创建测试用的catalog
		catalogID := createTestCatalog(client, t)
		resourceIDs := []string{}

		// ========== 删除测试（DS401-DS410） ==========

		Convey("DS401: 删除存在的dataset资源", func() {
			// 创建
			payload := buildDatasetResourcePayload()
			payload["catalog_id"] = catalogID
			client.SetHeader("Content-Type", "application/json")
			createResp := client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			resourceID := createResp.Body["id"].(string)
			resourceIDs = append(resourceIDs, resourceID)

			// 删除
			deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证已删除
			getResp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(getResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("DS402: 删除不存在的resource", func() {
			resp := client.DELETE("/api/vega-backend/v1/resources/non-existent-id-12345")
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		// 清理资源
		deleteTestResourceAndCatalog(client, t, resourceIDs, catalogID)
	})
}

// TestDatasetDocumentsCreate 测试批量创建dataset文档
func TestDatasetDocumentsCreate(t *testing.T) {
	Convey("DD101: 批量创建dataset文档", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 创建测试用的catalog
		catalogID := createTestCatalog(client, t)
		resourceIDs := []string{}

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		payload["catalog_id"] = catalogID
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)
		resourceIDs = append(resourceIDs, resourceID)

		// 构建批量创建文档的payload
		documentsPayload := []map[string]any{
			{
				"@timestamp": time.Now().UnixMilli(),
				"name":       "User 1",
				"age":        25,
				"content":    generateVector(768),
			},
			{
				"@timestamp": time.Now().UnixMilli(),
				"name":       "User 2",
				"age":        35,
				"content":    generateVector(768),
			},
			{
				"@timestamp": time.Now().UnixMilli(),
				"name":       "User 3",
				"age":        40,
				"content":    generateVector(768),
			},
		}
		resp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", documentsPayload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["ids"], ShouldNotBeEmpty)
		ids, ok := resp.Body["ids"].([]interface{})
		So(ok, ShouldBeTrue)
		So(len(ids), ShouldEqual, 3)

		// 清理资源
		deleteTestResourceAndCatalog(client, t, resourceIDs, catalogID)
	})
}

// TestDatasetDocumentsList 测试列出dataset文档
func TestDatasetDocumentsList(t *testing.T) {
	Convey("DD102: 列出dataset文档", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 创建测试用的catalog
		catalogID := createTestCatalog(client, t)
		resourceIDs := []string{}

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		payload["catalog_id"] = catalogID
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)
		resourceIDs = append(resourceIDs, resourceID)

		// 先创建一些文档
		documentsPayload := []map[string]any{
			{
				"@timestamp": time.Now().UnixMilli(),
				"id":         "doc1",
				"name":       "User 1",
				"age":        25,
				"email":      "user1@example.com",
				"active":     true,
				"tags":       []string{"tag1", "tag2"},
				"content":    generateVector(768),
			},
			{
				"@timestamp": time.Now().UnixMilli() + 1000,
				"id":         "doc2",
				"name":       "User 2",
				"age":        30,
				"email":      "user2@example.com",
				"active":     false,
				"tags":       []string{"tag2", "tag3"},
				"content":    generateVector(768),
			},
			{
				"@timestamp": time.Now().UnixMilli() + 2000,
				"id":         "doc3",
				"name":       "Admin User",
				"age":        35,
				"email":      "admin@example.com",
				"active":     true,
				"content":    generateVector(768),
			},
		}
		createDocsResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", documentsPayload)
		So(createDocsResp.StatusCode, ShouldEqual, http.StatusCreated)

		// 构建基础查询条件
		baseQuery := map[string]any{
			"start": time.Now().UnixMilli() - (24 * 3600 * 1000),
			"end":   time.Now().UnixMilli() + (24 * 3600 * 1000),
			"sort": []map[string]any{
				{
					"field":     "@timestamp",
					"direction": "asc",
				},
			},
			"offset":           0,
			"limit":            10,
			"need_total":       true,
			"use_search_after": false,
		}

		// 测试函数
		testFilterQuery := func(filterCondition map[string]any, expectedMinResults int) {
			query := make(map[string]any)
			for k, v := range baseQuery {
				query[k] = v
			}
			query["filter_condition"] = filterCondition

			// 使用POST请求到/data端点（method override GET）
			client.SetHeader("X-HTTP-Method-Override", "GET")
			resp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", query)
			client.RemoveHeader("X-HTTP-Method-Override")
			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			if resp.Body != nil && resp.Body["entries"] != nil {
				entries := resp.Body["entries"].([]any)
				So(len(entries), ShouldBeGreaterThanOrEqualTo, expectedMinResults)
			}
		}

		// 测试等于查询
		Convey("DD102.1: 等于查询 (eq)", func() {
			testFilterQuery(map[string]any{
				"operation":  "eq",
				"field":      "name",
				"value":      "User 1",
				"value_from": "const",
			}, 1)
		})

		// 测试不等于查询
		Convey("DD102.2: 不等于查询 (not_eq)", func() {
			testFilterQuery(map[string]any{
				"operation":  "!=",
				"field":      "name",
				"value":      "User 1",
				"value_from": "const",
			}, 2)
		})

		// 测试大于查询
		Convey("DD102.3: 大于查询 (gt)", func() {
			testFilterQuery(map[string]any{
				"operation":  "gt",
				"field":      "age",
				"value":      25,
				"value_from": "const",
			}, 2)
		})

		// 测试大于等于查询
		Convey("DD102.4: 大于等于查询 (gte)", func() {
			testFilterQuery(map[string]any{
				"operation":  "gte",
				"field":      "age",
				"value":      30,
				"value_from": "const",
			}, 2)
		})

		// 测试小于查询
		Convey("DD102.5: 小于查询 (lt)", func() {
			testFilterQuery(map[string]any{
				"operation":  "lt",
				"field":      "age",
				"value":      30,
				"value_from": "const",
			}, 1)
		})

		// 测试小于等于查询
		Convey("DD102.6: 小于等于查询 (lte)", func() {
			testFilterQuery(map[string]any{
				"operation":  "lte",
				"field":      "age",
				"value":      30,
				"value_from": "const",
			}, 2)
		})

		// 测试在集合中查询
		Convey("DD102.7: 在集合中查询 (in)", func() {
			testFilterQuery(map[string]any{
				"operation":  "in",
				"field":      "id",
				"value":      []string{"doc1", "doc2"},
				"value_from": "const",
			}, 2)
		})

		// 测试不在集合中查询
		Convey("DD102.8: 不在集合中查询 (not_in)", func() {
			testFilterQuery(map[string]any{
				"operation":  "not_in",
				"field":      "id",
				"value":      []string{"doc1"},
				"value_from": "const",
			}, 2)
		})

		// 测试模糊匹配查询
		Convey("DD102.9: 模糊匹配查询 (like)", func() {
			testFilterQuery(map[string]any{
				"operation":  "like",
				"field":      "name",
				"value":      "User%",
				"value_from": "const",
			}, 2)
		})

		// 测试模糊不匹配查询
		Convey("DD102.10: 不模糊匹配查询 (not_like)", func() {
			testFilterQuery(map[string]any{
				"operation":  "not_like",
				"field":      "name",
				"value":      "%User",
				"value_from": "const",
			}, 2)
		})

		// 测试包含查询
		Convey("DD102.11: 包含查询 (contain)", func() {
			testFilterQuery(map[string]any{
				"operation":  "contain",
				"field":      "tags",
				"value":      []any{"tag2"},
				"value_from": "const",
			}, 2)
		})

		// 测试不包含查询
		Convey("DD102.12: 不包含查询 (not_contain)", func() {
			testFilterQuery(map[string]any{
				"operation":  "not_contain",
				"field":      "tags",
				"value":      []any{"tag3"},
				"value_from": "const",
			}, 2)
		})

		// 测试前缀匹配查询
		Convey("DD102.13: 前缀匹配查询 (prefix)", func() {
			testFilterQuery(map[string]any{
				"operation":  "prefix",
				"field":      "email",
				"value":      "user",
				"value_from": "const",
			}, 2)
		})

		// 测试不前缀匹配查询
		Convey("DD102.14: 不前缀匹配查询 (not_prefix)", func() {
			testFilterQuery(map[string]any{
				"operation":  "not_prefix",
				"field":      "email",
				"value":      "user",
				"value_from": "const",
			}, 1)
		})

		// 测试字段存在查询
		Convey("DD102.15: 字段存在查询 (exist)", func() {
			testFilterQuery(map[string]any{
				"operation":  "exist",
				"field":      "tags",
				"value_from": "const",
			}, 2)
		})

		// 测试字段不存在查询
		Convey("DD102.16: 字段不存在查询 (not_exist)", func() {
			testFilterQuery(map[string]any{
				"operation":  "not_exist",
				"field":      "tags",
				"value_from": "const",
			}, 1)
		})

		// 测试字段为null查询
		Convey("DD102.17: 字段为null查询 (null)", func() {
			testFilterQuery(map[string]any{
				"operation":  "null",
				"field":      "tags",
				"value_from": "const",
			}, 1)
		})

		// 测试字段不为null查询
		Convey("DD102.18: 字段不为null查询 (not_null)", func() {
			testFilterQuery(map[string]any{
				"operation":  "not_null",
				"field":      "tags",
				"value_from": "const",
			}, 2)
		})

		// 测试布尔值true查询
		Convey("DD102.19: 布尔值true查询 (true)", func() {
			testFilterQuery(map[string]any{
				"operation":  "true",
				"field":      "active",
				"value_from": "const",
			}, 2)
		})

		// 测试布尔值false查询
		Convey("DD102.20: 布尔值false查询 (false)", func() {
			testFilterQuery(map[string]any{
				"operation":  "false",
				"field":      "active",
				"value_from": "const",
			}, 1)
		})

		// 测试范围查询
		Convey("DD102.21: 范围查询 (range)", func() {
			testFilterQuery(map[string]any{
				"operation":  "range",
				"field":      "age",
				"value":      []int{25, 35},
				"value_from": "const",
			}, 2)
		})

		// 测试between查询
		Convey("DD102.22: between查询 (between)", func() {
			testFilterQuery(map[string]any{
				"operation":  "between",
				"field":      "age",
				"value":      []int{25, 35},
				"value_from": "const",
			}, 3)
		})

		// 测试逻辑AND查询
		Convey("DD102.23: 逻辑AND查询 (and)", func() {
			testFilterQuery(map[string]any{
				"operation": "and",
				"sub_conditions": []map[string]any{
					{
						"operation":  "eq",
						"field":      "active",
						"value":      true,
						"value_from": "const",
					},
					{
						"operation":  "gt",
						"field":      "age",
						"value":      30,
						"value_from": "const",
					},
				},
			}, 1)
		})

		// 测试逻辑OR查询
		Convey("DD102.24: 逻辑OR查询 (or)", func() {
			testFilterQuery(map[string]any{
				"operation": "or",
				"sub_conditions": []map[string]any{
					{
						"operation":  "eq",
						"field":      "name",
						"value":      "User 1",
						"value_from": "const",
					},
					{
						"operation":  "eq",
						"field":      "name",
						"value":      "User 2",
						"value_from": "const",
					},
				},
			}, 2)
		})

		// 测试match查询
		Convey("DD102.25: match查询 (match field)", func() {
			testFilterQuery(map[string]any{
				"operation":  "match",
				"field":      "name",
				"value":      "User",
				"value_from": "const",
			}, 2)
		})

		// 测试match查询
		Convey("DD102.25.2: match查询 (match fields)", func() {
			testFilterQuery(map[string]any{
				"operation":  "match",
				"fields":     []string{"name"},
				"value":      "User",
				"value_from": "const",
			}, 2)
		})

		// 测试match_phrase查询
		Convey("DD102.26: match_phrase查询 (match_phrase field)", func() {
			testFilterQuery(map[string]any{
				"operation":  "match_phrase",
				"field":      "name",
				"value":      "Admin User",
				"value_from": "const",
			}, 1)
		})

		// 测试match_phrase查询
		Convey("DD102.26.2: match_phrase查询 (match_phrase fields)", func() {
			testFilterQuery(map[string]any{
				"operation":  "match_phrase",
				"fields":     []string{"name"},
				"value":      "Admin User",
				"value_from": "const",
			}, 1)
		})

		// 测试multi_match查询
		Convey("DD102.27: multi_match查询 (multi_match)", func() {
			testFilterQuery(map[string]any{
				"operation":  "multi_match",
				"match_type": "best_fields",
				"fields":     []string{"name", "email"},
				"value":      "User",
				"value_from": "const",
			}, 2)
		})

		// 测试regex查询
		Convey("DD102.28: regex查询 (regex)", func() {
			testFilterQuery(map[string]any{
				"operation":  "regex",
				"field":      "email",
				"value":      ".*@example\\.com",
				"value_from": "const",
			}, 3)
		})

		// 测试empty查询
		Convey("DD102.29: empty查询 (empty)", func() {
			testFilterQuery(map[string]any{
				"operation":  "empty",
				"field":      "tags",
				"value_from": "const",
			}, 1)
		})

		// 测试not_empty查询
		Convey("DD102.30: not_empty查询 (not_empty)", func() {
			testFilterQuery(map[string]any{
				"operation":  "not_empty",
				"field":      "tags",
				"value_from": "const",
			}, 2)
		})

		// 测试out_range查询
		Convey("DD102.31: out_range查询 (out_range)", func() {
			testFilterQuery(map[string]any{
				"operation":  "out_range",
				"field":      "age",
				"value":      []int{30, 35},
				"value_from": "const",
			}, 1)
		})

		// 测试vector查询
		Convey("DD102.32: vector查询 (vector)", func() {
			testFilterQuery(map[string]any{
				"operation":   "knn_vector",
				"field":       "content",
				"value":       generateVector(768),
				"value_from":  "const",
				"limit_key":   "k",
				"limit_value": 3,
			}, 1)
		})

		// 测试AND里面嵌套OR查询
		Convey("DD102.33: AND里面嵌套OR查询 (and with nested or)", func() {
			testFilterQuery(map[string]any{
				"operation": "and",
				"sub_conditions": []map[string]any{
					{
						"operation":  "eq",
						"field":      "active",
						"value":      true,
						"value_from": "const",
					},
					{
						"operation": "or",
						"sub_conditions": []map[string]any{
							{
								"operation":  "eq",
								"field":      "name",
								"value":      "User 1",
								"value_from": "const",
							},
							{
								"operation":  "eq",
								"field":      "name",
								"value":      "Admin User",
								"value_from": "const",
							},
						},
					},
				},
			}, 2)
		})

		// 测试in查询text字段
		Convey("DD102.34: in查询text字段 (in text field)", func() {
			testFilterQuery(map[string]any{
				"operation":  "in",
				"field":      "name",
				"value":      []string{"User 1", "Admin User"},
				"value_from": "const",
			}, 2)
		})

		// 测试not_in查询text字段
		Convey("DD102.35: not_in查询text字段 (not_in text field)", func() {
			testFilterQuery(map[string]any{
				"operation":  "not_in",
				"field":      "name",
				"value":      []string{"User 2"},
				"value_from": "const",
			}, 2)
		})

		// 清理资源
		deleteTestResourceAndCatalog(client, t, resourceIDs, catalogID)
	})
}

// TestDatasetDocumentGet 测试获取dataset文档
func TestDatasetDocumentGet(t *testing.T) {
	Convey("DD103: 获取dataset文档", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 创建测试用的catalog
		catalogID := createTestCatalog(client, t)
		resourceIDs := []string{}

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		payload["catalog_id"] = catalogID
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)
		resourceIDs = append(resourceIDs, resourceID)

		// 先创建一个文档（使用批量创建接口）
		docPayload := buildDatasetDocumentPayload()
		createDocResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", []map[string]any{docPayload})
		So(createDocResp.StatusCode, ShouldEqual, http.StatusCreated)
		So(createDocResp.Body["ids"], ShouldNotBeEmpty)
		ids, ok := createDocResp.Body["ids"].([]interface{})
		So(ok, ShouldBeTrue)
		So(len(ids), ShouldBeGreaterThan, 0)

		// 获取文档（使用POST /:id/data端点，method override GET）
		queryPayload := map[string]any{
			"start":      time.Now().UnixMilli() - (24 * 3600 * 1000),
			"end":        time.Now().UnixMilli(),
			"offset":     0,
			"limit":      10,
			"need_total": true,
		}
		client.SetHeader("X-HTTP-Method-Override", "GET")
		resp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", queryPayload)
		client.RemoveHeader("X-HTTP-Method-Override")
		So(resp.StatusCode, ShouldEqual, http.StatusOK)
		So(resp.Body["entries"], ShouldNotBeEmpty)

		// 清理资源
		deleteTestResourceAndCatalog(client, t, resourceIDs, catalogID)
	})
}

// TestDatasetDocumentUpdate 测试更新dataset文档
func TestDatasetDocumentUpdate(t *testing.T) {
	Convey("DD104: 更新dataset文档", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 创建测试用的catalog
		catalogID := createTestCatalog(client, t)
		resourceIDs := []string{}

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		payload["catalog_id"] = catalogID
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)
		resourceIDs = append(resourceIDs, resourceID)

		// 先创建一个文档（使用批量创建接口）
		docPayload := buildDatasetDocumentPayload()
		createDocResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", []map[string]any{docPayload})
		So(createDocResp.StatusCode, ShouldEqual, http.StatusCreated)
		So(createDocResp.Body["ids"], ShouldNotBeEmpty)
		ids, ok := createDocResp.Body["ids"].([]interface{})
		So(ok, ShouldBeTrue)
		So(len(ids), ShouldBeGreaterThan, 0)
		docID := ids[0].(string)

		// 更新文档（使用批量更新接口）
		updatePayload := []map[string]any{
			{
				"id": docID,
				"document": map[string]any{
					"title":   "Updated Test Document",
					"content": generateVector(768),
					"metadata": map[string]any{
						"author":     "Updated Test User",
						"updated_at": "2024-01-02T00:00:00Z",
					},
				},
			},
		}

		resp := client.PUT("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", updatePayload)
		So(resp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 清理资源
		deleteTestResourceAndCatalog(client, t, resourceIDs, catalogID)
	})
}

// TestDatasetDocumentDelete 测试删除dataset文档
func TestDatasetDocumentDelete(t *testing.T) {
	Convey("DD105: 删除dataset文档", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 创建测试用的catalog
		catalogID := createTestCatalog(client, t)
		resourceIDs := []string{}

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		payload["catalog_id"] = catalogID
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)
		resourceIDs = append(resourceIDs, resourceID)

		// 先创建一个文档（使用批量创建接口）
		docPayload := buildDatasetDocumentPayload()
		createDocResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", []map[string]any{docPayload})
		So(createDocResp.StatusCode, ShouldEqual, http.StatusCreated)
		So(createDocResp.Body["ids"], ShouldNotBeEmpty)
		ids, ok := createDocResp.Body["ids"].([]interface{})
		So(ok, ShouldBeTrue)
		So(len(ids), ShouldBeGreaterThan, 0)
		docID := ids[0].(string)

		// 删除文档（使用批量删除接口）
		resp := client.DELETE("/api/vega-backend/v1/resources/dataset/" + resourceID + "/docs/" + docID)
		So(resp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 清理资源
		deleteTestResourceAndCatalog(client, t, resourceIDs, catalogID)
	})
}

// TestDatasetDocumentDeleteByQuery 测试通过查询条件删除dataset文档
func TestDatasetDocumentDeleteByQuery(t *testing.T) {
	Convey("DD106: 通过查询条件删除dataset文档", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 创建测试用的catalog
		catalogID := createTestCatalog(client, t)
		resourceIDs := []string{}

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		payload["catalog_id"] = catalogID
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)
		resourceIDs = append(resourceIDs, resourceID)

		// 先创建多个文档（使用批量创建接口）
		documentsPayload := []map[string]any{
			{
				"@timestamp": time.Now().UnixMilli(),
				"name":       "User 1",
				"age":        25,
				"content":    generateVector(768),
			},
			{
				"@timestamp": time.Now().UnixMilli(),
				"name":       "User 2",
				"age":        35,
				"content":    generateVector(768),
			},
			{
				"@timestamp": time.Now().UnixMilli(),
				"name":       "User 3",
				"age":        40,
				"content":    generateVector(768),
			},
		}
		createDocResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", documentsPayload)
		So(createDocResp.StatusCode, ShouldEqual, http.StatusCreated)
		So(createDocResp.Body["ids"], ShouldNotBeEmpty)

		// 构建查询条件（例如删除age大于30的文档）
		// 直接使用OpenSearch查询DSL格式
		queryPayload := map[string]any{
			"filter_condition": map[string]any{
				"operation":  "range",
				"field":      "age",
				"value":      []int{25, 35},
				"value_from": "const",
			},
		}

		// 通过查询条件删除文档（使用POST请求，method override DELETE）
		client.SetHeader("X-HTTP-Method-Override", "DELETE")
		resp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs/query", queryPayload)
		client.RemoveHeader("X-HTTP-Method-Override")
		So(resp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 验证是否删除了符合条件的文档
		// 使用POST请求，method override GET
		client.SetHeader("X-HTTP-Method-Override", "GET")
		getResp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", map[string]any{
			"offset":     0,
			"limit":      10,
			"need_total": false,
		})
		client.RemoveHeader("X-HTTP-Method-Override")
		So(getResp.StatusCode, ShouldEqual, http.StatusOK)
		if getResp.Body != nil && getResp.Body["entries"] != nil {
			entries := getResp.Body["entries"].([]any)
			So(len(entries), ShouldEqual, 1) // 只有User 1符合条件（age=25）
		}

		// 清理资源
		deleteTestResourceAndCatalog(client, t, resourceIDs, catalogID)
	})
}

// TestDatasetDocumentsSearchAfter 测试使用search after进行分页查询
func TestDatasetDocumentsSearchAfter(t *testing.T) {
	Convey("DD107: 使用search after进行分页查询", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 创建测试用的catalog
		catalogID := createTestCatalog(client, t)
		resourceIDs := []string{}

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		payload["catalog_id"] = catalogID
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)
		resourceIDs = append(resourceIDs, resourceID)

		// 先创建多个文档（使用批量创建接口）
		documentsPayload := []map[string]any{}
		for i := 1; i <= 5; i++ {
			documentsPayload = append(documentsPayload, map[string]any{
				"@timestamp": time.Now().UnixMilli() + int64(i),
				"name":       fmt.Sprintf("User %d", i),
				"age":        20 + i,
				"content":    generateVector(768),
			})
		}
		createDocResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", documentsPayload)
		So(createDocResp.StatusCode, ShouldEqual, http.StatusCreated)
		So(createDocResp.Body["ids"], ShouldNotBeEmpty)

		// 第一次查询，使用sort和limit=2
		client.SetHeader("X-HTTP-Method-Override", "GET")
		firstResp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", map[string]any{
			"limit": 2,
			"sort": []map[string]any{
				{
					"field":     "age",
					"direction": "asc",
				},
			},
			"need_total": true,
		})
		client.RemoveHeader("X-HTTP-Method-Override")
		So(firstResp.StatusCode, ShouldEqual, http.StatusOK)
		So(firstResp.Body["entries"], ShouldNotBeEmpty)
		entries := firstResp.Body["entries"].([]any)
		So(len(entries), ShouldEqual, 2)

		// 提取最后一个文档的sort值作为search_after参数
		var searchAfter []any
		if len(entries) > 0 {
			lastEntry := entries[len(entries)-1].(map[string]any)
			// 假设age是排序字段，使用age值作为search_after
			if age, ok := lastEntry["age"].(float64); ok {
				searchAfter = []any{age}
			}
		}

		// 第二次查询，使用search_after进行分页
		client.SetHeader("X-HTTP-Method-Override", "GET")
		secondResp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", map[string]any{
			"limit": 2,
			"sort": []map[string]any{
				{
					"field":     "age",
					"direction": "asc",
				},
			},
			"search_after": searchAfter,
			"need_total":   true,
		})
		client.RemoveHeader("X-HTTP-Method-Override")
		So(secondResp.StatusCode, ShouldEqual, http.StatusOK)
		So(secondResp.Body["entries"], ShouldNotBeEmpty)
		secondEntries := secondResp.Body["entries"].([]any)
		So(len(secondEntries), ShouldEqual, 2)

		// 清理资源
		deleteTestResourceAndCatalog(client, t, resourceIDs, catalogID)
	})
}

// TestDatasetDocumentsSourceFilter 测试指定_source搜索
func TestDatasetDocumentsSourceFilter(t *testing.T) {
	Convey("DD108: 指定_source搜索", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		// 创建测试用的catalog
		catalogID := createTestCatalog(client, t)
		resourceIDs := []string{}

		// 创建测试用的dataset resource
		payload := buildDatasetResourcePayload()
		payload["catalog_id"] = catalogID
		createResp := client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)
		resourceIDs = append(resourceIDs, resourceID)

		// 先创建文档（使用批量创建接口）
		documentsPayload := []map[string]any{
			{
				"@timestamp": time.Now().UnixMilli(),
				"name":       "Test User",
				"age":        30,
				"content":    generateVector(768),
			},
		}
		createDocResp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", documentsPayload)
		So(createDocResp.StatusCode, ShouldEqual, http.StatusCreated)
		So(createDocResp.Body["ids"], ShouldNotBeEmpty)

		// 测试指定输出字段（只返回name和age字段）
		client.SetHeader("X-HTTP-Method-Override", "GET")
		resp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", map[string]any{
			"offset":        0,
			"limit":         10,
			"output_fields": []string{"name", "age", "_score"},
			"need_total":    false,
		})
		client.RemoveHeader("X-HTTP-Method-Override")
		So(resp.StatusCode, ShouldEqual, http.StatusOK)
		So(resp.Body["entries"], ShouldNotBeEmpty)

		// 验证返回的文档只包含指定的字段
		entries := resp.Body["entries"].([]any)
		So(len(entries), ShouldEqual, 1)

		doc := entries[0].(map[string]any)
		// 检查是否包含name和age字段
		So(doc["name"], ShouldNotBeEmpty)
		So(doc["age"], ShouldNotBeEmpty)
		So(doc["_score"], ShouldNotBeEmpty)
		// 检查是否不包含content字段
		_, hasContent := doc["content"]
		So(hasContent, ShouldBeFalse)

		// 清理资源
		deleteTestResourceAndCatalog(client, t, resourceIDs, catalogID)
	})
}
