// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package rate

import (
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend-tests/at/setup"
	"vega-backend-tests/testutil"
)

func generateUniqueName(prefix string) string {
	suffix := rand.Intn(10000)
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().Unix(), suffix)
}

func generateVector(dims int) []float64 {
	vector := make([]float64, dims)
	for i := range vector {
		vector[i] = rand.Float64()*2 - 1
	}
	return vector
}

func createTestCatalog(client *testutil.HTTPClient, t *testing.T) string {
	catalogPayload := map[string]any{
		"name":           generateUniqueName("test-concurrency-catalog"),
		"description":    "测试并发查询的catalog",
		"tags":           []string{"test", "concurrency", "catalog"},
		"connector_type": "mysql",
		"connector_config": map[string]any{
			"host":     "localhost",
			"port":     3330,
			"username": "username",
			"password": "password",
			"database": "test",
		},
	}
	catalogResp := client.POST("/api/vega-backend/v1/catalogs", catalogPayload)
	So(catalogResp.StatusCode, ShouldEqual, http.StatusCreated)
	So(catalogResp.Body["id"], ShouldNotBeEmpty)
	return catalogResp.Body["id"].(string)
}

func createTestDatasetResource(client *testutil.HTTPClient, catalogID string) string {
	payload := map[string]any{
		"catalog_id":        catalogID,
		"name":              generateUniqueName("test-concurrency-dataset"),
		"tags":              []string{"test", "concurrency", "dataset"},
		"description":       "测试并发查询的数据集",
		"category":          "dataset",
		"status":            "active",
		"source_identifier": "at_db",
		"schema_definition": []map[string]any{
			{"name": "id", "type": "keyword", "display_name": "ID", "original_name": "id"},
			{"name": "@timestamp", "type": "long", "display_name": "时间戳", "original_name": "@timestamp"},
			{"name": "name", "type": "text", "display_name": "名称", "original_name": "name"},
			{"name": "age", "type": "integer", "display_name": "年龄", "original_name": "age"},
			{"name": "content", "type": "vector", "display_name": "内容向量", "original_name": "content", "features": []map[string]any{
				{
					"name":         "content",
					"display_name": "内容向量",
					"feature_type": "vector",
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
	resp := client.POST("/api/vega-backend/v1/resources", payload)
	So(resp.StatusCode, ShouldEqual, http.StatusCreated)
	So(resp.Body["id"], ShouldNotBeEmpty)
	return resp.Body["id"].(string)
}

func createTestDocuments(client *testutil.HTTPClient, resourceID string, count int) {
	documentsPayload := []map[string]any{}
	for i := 0; i < count; i++ {
		documentsPayload = append(documentsPayload, map[string]any{
			"@timestamp": time.Now().UnixMilli(),
			"id":         fmt.Sprintf("doc-%d", i),
			"name":       fmt.Sprintf("User %d", i),
			"age":        20 + i%30,
			"content":    generateVector(768),
		})
	}
	resp := client.POST("/api/vega-backend/v1/resources/dataset/"+resourceID+"/docs", documentsPayload)
	So(resp.StatusCode, ShouldEqual, http.StatusCreated)
	So(resp.Body["ids"], ShouldNotBeEmpty)
}

func TestConcurrencyQuery(t *testing.T) {
	Convey("并发查询测试 - RC101", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		catalogID := createTestCatalog(client, t)
		resourceID := createTestDatasetResource(client, catalogID)

		t.Logf("✓ 创建测试资源完成，catalogID: %s, resourceID: %s", catalogID, resourceID)

		createTestDocuments(client, resourceID, 10)
		t.Logf("✓ 创建测试文档完成")

		Convey("RC101: 并发查询 /data 接口", func() {
			// 修改 config\vega-backend-config.yaml 中的 max_concurrent_queries 为 10
			requestCount := 10

			var wg sync.WaitGroup
			successCount := 0
			limitExceededCount := 0
			var mu sync.Mutex

			for i := 0; i < requestCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					client.SetHeader("X-HTTP-Method-Override", "GET")
					resp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", map[string]any{
						"offset":     0,
						"limit":      5,
						"need_total": true,
					})

					mu.Lock()
					if resp.StatusCode == http.StatusTooManyRequests {
						limitExceededCount++
					} else if resp.StatusCode == http.StatusOK {
						successCount++
					}
					mu.Unlock()
				}()
			}

			wg.Wait()

			t.Logf("并发测试完成 - 总请求数: %d, 成功: %d, 失败: %d",
				requestCount, successCount, limitExceededCount)

			So(limitExceededCount, ShouldEqual, 0)
			So(successCount, ShouldEqual, requestCount)
		})

		Convey("RC102: 超过全局并发限制时返回 ErrGlobalLimitExceeded", func() {
			concurrency := 10 // 修改 config\vega-backend-config.yaml 中的 max_concurrent_queries 为 10
			requestCount := 30
			var wg sync.WaitGroup
			globalLimitExceededCount := 0
			catalogLimitExceededCount := 0
			successCount := 0
			var mu sync.Mutex

			for i := 0; i < requestCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					client.SetHeader("X-HTTP-Method-Override", "GET")
					resp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", map[string]any{
						"offset":     0,
						"limit":      10,
						"need_total": true,
					})

					mu.Lock()
					if resp.StatusCode == http.StatusTooManyRequests {
						if errorMsg, ok := resp.Body["error_message"].(string); ok {
							if strings.Contains(errorMsg, "global concurrency limit exceeded") {
								globalLimitExceededCount++
							} else if strings.Contains(errorMsg, "catalog concurrency limit exceeded") {
								catalogLimitExceededCount++
							}
						} else {
							globalLimitExceededCount++ // 默认认为是全局限流
						}
					} else if resp.StatusCode == http.StatusOK {
						successCount++
					}
					mu.Unlock()
				}()
			}

			wg.Wait()

			t.Logf("全局并发限流测试完成 - 成功: %d, 全局限流: %d, catalog限流: %d", successCount, globalLimitExceededCount, catalogLimitExceededCount)
			So(successCount+globalLimitExceededCount+catalogLimitExceededCount, ShouldEqual, requestCount)
			So(successCount, ShouldBeGreaterThanOrEqualTo, concurrency)
			So(globalLimitExceededCount, ShouldBeGreaterThan, 0) // 必须有全局限流
			So(catalogLimitExceededCount, ShouldEqual, 0) // catalog限流无效
		})

		Convey("RC103: Catalog级别限流测试 - 全局并发10（与RC101对比），catalog并发2", func() {
			catalogConcurrencyLimit := int64(2)

			updateCatalogPayload := map[string]any{
				"id":             catalogID,
				"name":           generateUniqueName("test-concurrency-catalog"),
				"description":    "测试并发查询的catalog",
				"tags":           []string{"test", "concurrency", "catalog"},
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":       "localhost",
					"port":       3330,
					"username":   "username",
					"password":   "password",
					"database":   "test",
					"concurrent": catalogConcurrencyLimit,
				},
			}
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updateCatalogPayload)
			if updateResp.StatusCode != http.StatusNoContent {
				t.Logf("更新catalog失败，状态码: %d，响应体: %v", updateResp.StatusCode, updateResp.Body)
			}
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 全局并发，修改 config\vega-backend-config.yaml 中的 max_concurrent_queries 为 10
			requestCount := 10

			var wg sync.WaitGroup
			successCount := 0
			globalLimitExceededCount := 0
			catalogLimitExceededCount := 0
			var mu sync.Mutex

			for i := 0; i < requestCount; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					client.SetHeader("X-HTTP-Method-Override", "GET")
					resp := client.POST("/api/vega-backend/v1/resources/"+resourceID+"/data", map[string]any{
						"offset":     0,
						"limit":      5,
						"need_total": true,
					})

					mu.Lock()
					if resp.StatusCode == http.StatusTooManyRequests {
						if errorMsg, ok := resp.Body["error_message"].(string); ok {
							if strings.Contains(errorMsg, "global concurrency limit exceeded") {
								globalLimitExceededCount++
							} else if strings.Contains(errorMsg, "catalog concurrency limit exceeded") {
								catalogLimitExceededCount++
							}
						} else {
							catalogLimitExceededCount++ // 默认认为是catalog限流
						}
					} else if resp.StatusCode == http.StatusOK {
						successCount++
					}
					mu.Unlock()
				}()
			}

			wg.Wait()

			t.Logf("Catalog级别限流测试完成 - 成功: %d, 全局限流: %d, catalog限流: %d", successCount, globalLimitExceededCount, catalogLimitExceededCount)
			So(successCount+globalLimitExceededCount+catalogLimitExceededCount, ShouldEqual, requestCount)
			So(successCount, ShouldBeGreaterThanOrEqualTo, int(catalogConcurrencyLimit))
			So(catalogLimitExceededCount, ShouldBeGreaterThan, 0) // 必须有catalog级限流（全局10并发，但是catalog最大2并发）
			So(globalLimitExceededCount, ShouldEqual, 0) // 全局限流无效
		})

		deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
		So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

		deleteCatalogResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
		So(deleteCatalogResp.StatusCode, ShouldEqual, http.StatusNoContent)
	})
}
