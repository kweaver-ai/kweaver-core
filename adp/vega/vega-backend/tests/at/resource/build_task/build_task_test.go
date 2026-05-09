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

// cleanupResources 清理现有资源
func cleanupResources(client *testutil.HTTPClient, t *testing.T) {
	// 先删除所有 build 任务
	resp := client.GET("/api/vega-backend/v1/build-tasks?offset=0&limit=100")
	if resp.StatusCode == http.StatusOK {
		if entries, ok := resp.Body["entries"].([]any); ok {
			for _, task := range entries {
				if taskMap, ok := task.(map[string]any); ok {
					if taskID, ok := taskMap["id"].(string); ok {
						status := taskMap["status"].(string)
						if status == "running" {
							// 先尝试停止构建任务
							stopResp := client.POST("/api/vega-backend/v1/build-tasks/"+taskID+"/stop", nil)
							if stopResp.StatusCode != http.StatusOK {
								t.Logf("停止 build 任务失败 %s: %d，响应体：%v", taskID, stopResp.StatusCode, stopResp.Body)
							}
							time.Sleep(5 * time.Second)
						}
						// 然后删除构建任务
						deleteTaskResp := client.DELETE("/api/vega-backend/v1/build-tasks/" + taskID)
						if deleteTaskResp.StatusCode != http.StatusNoContent {
							t.Logf("删除 build 任务失败 %s: %d，响应体：%v", taskID, deleteTaskResp.StatusCode, deleteTaskResp.Body)
						}
					}
				}
			}
		}
	}

	// 再删除所有 dataset 资源
	resp = client.GET("/api/vega-backend/v1/resources?category=dataset&offset=0&limit=100")
	if resp.StatusCode == http.StatusOK {
		if entries, ok := resp.Body["entries"].([]any); ok {
			for _, entry := range entries {
				if entryMap, ok := entry.(map[string]any); ok {
					if id, ok := entryMap["id"].(string); ok {
						// 删除资源（使用正确的 API 路径格式）
						deleteResp := client.DELETE("/api/vega-backend/v1/resources/" + id)
						if deleteResp.StatusCode != http.StatusOK && deleteResp.StatusCode != http.StatusNoContent {
							t.Logf("清理资源失败 %s: %d，响应体：%v", id, deleteResp.StatusCode, deleteResp.Body)
						}
					}
				}
			}
		}
	}
}

// TestResourceBatchBuildForMySQL 测试resource批量构建 - 先创建catalog，再创建resource，最后构建
// 测试编号前缀: DS2xx (Resource Build)
func TestResourceBatchBuildForMySQL(t *testing.T) {
	Convey("Resource批量构建For MySQL AT测试 - 初始化", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		// 清理现有资源
		cleanupResources(client, t)
		// 清理现有catalog
		cleanupCatalogs(client, t)

		// ========== 构建测试（DS201-DS210） ==========

		Convey("DS201: 先创建mysql catalog，再创建resource，最后构建", func() {
			// 1. 创建mysql catalog
			catalogPayload := map[string]any{
				"name":           generateUniqueName("test-mysql-catalog"),
				"description":    "测试mysql catalog",
				"tags":           []string{"test", "mysql", "catalog"},
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
			catalogID := catalogResp.Body["id"].(string)

			// 2. 创建resource，使用刚创建的catalog
			resourcePayload := map[string]any{
				"catalog_id":        catalogID,
				"name":              generateUniqueName("test-resource-build"),
				"tags":              []string{"test", "resource"},
				"description":       "测试资源构建",
				"category":          "table",
				"status":            "active",
				"database":          "test",
				"source_identifier": "test.users",
				"schema_definition": []map[string]any{
					{"name": "id", "type": "keyword", "display_name": "ID", "original_name": "id", "description": "唯一标识符"},
					{"name": "name", "type": "keyword", "display_name": "名称", "original_name": "name", "description": "用户名称"},
					{"name": "address", "type": "text", "display_name": "地址", "original_name": "address", "description": "用户地址"},
					{"name": "hobby", "type": "text", "display_name": "爱好", "original_name": "hobby", "description": "用户爱好"},
				},
			}
			resourceResp := client.POST("/api/vega-backend/v1/resources", resourcePayload)
			So(resourceResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resourceResp.Body["id"], ShouldNotBeEmpty)
			resourceID := resourceResp.Body["id"].(string)

			// 3. 创建构建任务
			buildResp := client.POST("/api/vega-backend/v1/build-tasks", map[string]any{"resource_id": resourceID, "mode": "batch", "embedding_fields": "hobby", "embedding_model": "", "build_key_fields": "id"})
			So(buildResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 验证构建成功
			So(buildResp.Body, ShouldNotBeNil)

			// 4. 获取构建任务详情
			buildTaskResp := client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			So(buildTaskResp.Body, ShouldNotBeNil)
			So(buildTaskResp.Body["status"], ShouldEqual, "init")

			// 5. 启动任务
			startResp := client.POST("/api/vega-backend/v1/build-tasks/"+buildResp.Body["id"].(string)+"/start", nil)
			So(startResp.StatusCode, ShouldEqual, http.StatusOK)

			time.Sleep(5 * time.Second)

			// 6. 验证任务状态为running
			buildTaskResp = client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			So(buildTaskResp.Body, ShouldNotBeNil)
			So(buildTaskResp.Body["status"], ShouldEqual, "running")

			time.Sleep(10 * time.Second)

			// // 7. 停止任务，由于测试数据量小，构建任务会很快完成，不需要停止
			// stopResp := client.POST("/api/vega-backend/v1/build-tasks/"+buildResp.Body["id"].(string)+"/stop", nil)
			// So(stopResp.StatusCode, ShouldEqual, http.StatusOK)

			// time.Sleep(10 * time.Second)

			// // 8. 验证任务状态为stopped或者stopping
			// buildTaskResp = client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			// So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			// So(buildTaskResp.Body, ShouldNotBeNil)
			// status := buildTaskResp.Body["status"].(string)
			// So(status == "stopped" || status == "stopping", ShouldBeTrue)
			// time.Sleep(5 * time.Second)

			// 8. 验证任务状态为completed
			buildTaskResp = client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			So(buildTaskResp.Body, ShouldNotBeNil)
			So(buildTaskResp.Body["status"], ShouldEqual, "completed")

			// 9. 删除任务
			deleteResp := client.DELETE("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

// TestResourceStreamingBuildForMySQL 测试resource流式构建 - 先创建catalog，再创建resource，最后构建
// 测试编号前缀: DS2xx (Resource Build)
func TestResourceStreamingBuildForMySQL(t *testing.T) {
	Convey("Resource流式构建For MySQL AT测试 - 初始化", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		// 清理现有资源
		cleanupResources(client, t)
		// 清理现有catalog
		cleanupCatalogs(client, t)

		// ========== 构建测试（DS201-DS210） ==========

		Convey("DS201: 先创建mysql catalog，再创建resource，最后构建", func() {
			// 1. 创建mysql catalog
			catalogPayload := map[string]any{
				"name":           generateUniqueName("test-mysql-catalog"),
				"description":    "测试mysql catalog",
				"tags":           []string{"test", "mysql", "catalog"},
				"connector_type": "mysql",
				"connector_config": map[string]any{
					"host":     "192.168.36.54",
					"port":     3306,
					"username": "root",
					"password": "Password123",
					"database": "test",
				},
			}
			catalogResp := client.POST("/api/vega-backend/v1/catalogs", catalogPayload)
			So(catalogResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(catalogResp.Body["id"], ShouldNotBeEmpty)
			catalogID := catalogResp.Body["id"].(string)

			// 2. 创建resource，使用刚创建的catalog
			resourcePayload := map[string]any{
				"catalog_id":        catalogID,
				"name":              generateUniqueName("test-resource-build"),
				"tags":              []string{"test", "resource"},
				"description":       "测试资源构建",
				"category":          "table",
				"status":            "active",
				"database":          "test",
				"source_identifier": "test.users",
				"source_metadata": map[string]any{
					"primary_keys": []string{"id"},
				},
				"schema_definition": []map[string]any{
					{"name": "id", "type": "keyword", "display_name": "ID", "original_name": "id", "description": "唯一标识符"},
					{"name": "name", "type": "keyword", "display_name": "名称", "original_name": "name", "description": "用户名称"},
					{"name": "address", "type": "text", "display_name": "地址", "original_name": "address", "description": "用户地址"},
					{"name": "hobby", "type": "text", "display_name": "爱好", "original_name": "hobby", "description": "用户爱好"},
				},
			}
			resourceResp := client.POST("/api/vega-backend/v1/resources", resourcePayload)
			So(resourceResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resourceResp.Body["id"], ShouldNotBeEmpty)
			resourceID := resourceResp.Body["id"].(string)

			// 3. 创建构建任务
			buildResp := client.POST("/api/vega-backend/v1/build-tasks", map[string]any{"resource_id": resourceID, "mode": "streaming", "embedding_fields": "hobby", "embedding_model": "", "build_key_fields": "id"})
			So(buildResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 验证构建成功
			So(buildResp.Body, ShouldNotBeNil)

			// 4. 获取构建任务详情
			buildTaskResp := client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			So(buildTaskResp.Body, ShouldNotBeNil)
			So(buildTaskResp.Body["status"], ShouldEqual, "init")

			// 5. 启动任务
			startResp := client.POST("/api/vega-backend/v1/build-tasks/"+buildResp.Body["id"].(string)+"/start", nil)
			So(startResp.StatusCode, ShouldEqual, http.StatusOK)

			time.Sleep(10 * time.Second)

			// 6. 验证任务状态为running
			buildTaskResp = client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			So(buildTaskResp.Body, ShouldNotBeNil)
			So(buildTaskResp.Body["status"], ShouldEqual, "running")

			// 7. 停止任务
			stopResp := client.POST("/api/vega-backend/v1/build-tasks/"+buildResp.Body["id"].(string)+"/stop", nil)
			So(stopResp.StatusCode, ShouldEqual, http.StatusOK)

			time.Sleep(10 * time.Second)

			// 8. 验证任务状态为stopped或者stopping
			buildTaskResp = client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			So(buildTaskResp.Body, ShouldNotBeNil)
			status := buildTaskResp.Body["status"].(string)
			So(status == "stopped" || status == "stopping", ShouldBeTrue)

			time.Sleep(5 * time.Second)

			// 9. 删除任务
			deleteResp := client.DELETE("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

// TestResourceBatchBuildForPG 测试resource批量构建 - 先创建catalog，再创建resource，最后构建
// 测试编号前缀: DS2xx (Resource Build)
func TestResourceBatchBuildForPG(t *testing.T) {
	Convey("Resource批量构建For PG AT测试 - 初始化", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		// 清理现有资源
		cleanupResources(client, t)
		// 清理现有catalog
		cleanupCatalogs(client, t)

		// ========== 构建测试（DS201-DS210） ==========

		Convey("DS201: 先创建PG catalog，再创建resource，最后构建", func() {
			// 1. 创建PG catalog
			catalogPayload := map[string]any{
				"name":           generateUniqueName("test-pg-catalog"),
				"description":    "测试PG catalog",
				"tags":           []string{"test", "pg", "catalog"},
				"connector_type": "postgresql",
				"connector_config": map[string]any{
					"host":     "192.168.36.54",
					"port":     5432,
					"username": "postgres",
					"password": "Password123",
					"database": "test",
				},
			}
			catalogResp := client.POST("/api/vega-backend/v1/catalogs", catalogPayload)
			So(catalogResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(catalogResp.Body["id"], ShouldNotBeEmpty)
			catalogID := catalogResp.Body["id"].(string)

			// 2. 创建resource，使用刚创建的catalog
			resourcePayload := map[string]any{
				"catalog_id":        catalogID,
				"name":              generateUniqueName("test-resource-build"),
				"tags":              []string{"test", "resource"},
				"description":       "测试资源构建",
				"category":          "table",
				"status":            "active",
				"database":          "test",
				"source_identifier": "public.users",
				"schema_definition": []map[string]any{
					{"name": "id", "type": "keyword", "display_name": "ID", "original_name": "id", "description": "唯一标识符"},
					{"name": "name", "type": "keyword", "display_name": "名称", "original_name": "name", "description": "用户名称"},
					{"name": "address", "type": "text", "display_name": "地址", "original_name": "address", "description": "用户地址"},
					{"name": "hobby", "type": "text", "display_name": "爱好", "original_name": "hobby", "description": "用户爱好"},
				},
			}
			resourceResp := client.POST("/api/vega-backend/v1/resources", resourcePayload)
			So(resourceResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resourceResp.Body["id"], ShouldNotBeEmpty)
			resourceID := resourceResp.Body["id"].(string)

			// 3. 创建构建任务
			buildResp := client.POST("/api/vega-backend/v1/build-tasks", map[string]any{"resource_id": resourceID, "mode": "batch", "embedding_fields": "hobby", "embedding_model": "", "build_key_fields": "id"})
			So(buildResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 验证构建成功
			So(buildResp.Body, ShouldNotBeNil)

			// 4. 获取构建任务详情
			buildTaskResp := client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			So(buildTaskResp.Body, ShouldNotBeNil)
			So(buildTaskResp.Body["status"], ShouldEqual, "init")

			// 5. 启动任务
			startResp := client.POST("/api/vega-backend/v1/build-tasks/"+buildResp.Body["id"].(string)+"/start", nil)
			So(startResp.StatusCode, ShouldEqual, http.StatusOK)

			time.Sleep(10 * time.Second)

			// 6. 验证任务状态为running或者completed
			buildTaskResp = client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			So(buildTaskResp.Body, ShouldNotBeNil)
			status := buildTaskResp.Body["status"].(string)
			So(status == "running" || status == "completed", ShouldBeTrue)
			if status == "running" {
				time.Sleep(10 * time.Second)
				// 验证任务状态为completed
				buildTaskResp = client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
				So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
				So(buildTaskResp.Body, ShouldNotBeNil)
				So(buildTaskResp.Body["status"], ShouldEqual, "completed")
			}

			time.Sleep(10 * time.Second)

			// // 7. 停止任务，由于测试数据量小，构建任务会很快完成，不需要停止
			// stopResp := client.POST("/api/vega-backend/v1/build-tasks/"+buildResp.Body["id"].(string)+"/stop", nil)
			// So(stopResp.StatusCode, ShouldEqual, http.StatusOK)

			// time.Sleep(10 * time.Second)

			// // 8. 验证任务状态为stopped或者stopping
			// buildTaskResp = client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			// So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			// So(buildTaskResp.Body, ShouldNotBeNil)
			// status := buildTaskResp.Body["status"].(string)
			// So(status == "stopped" || status == "stopping", ShouldBeTrue)
			// time.Sleep(5 * time.Second)

			// 9. 删除任务
			deleteResp := client.DELETE("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

// TestResourceStreamingBuildForPG 测试resource流式构建 - 先创建catalog，再创建resource，最后构建
// 测试编号前缀: DS2xx (Resource Build)
func TestResourceStreamingBuildForPG(t *testing.T) {
	Convey("Resource流式构建For PG AT测试 - 初始化", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		// 清理现有资源
		cleanupResources(client, t)
		// 清理现有catalog
		cleanupCatalogs(client, t)

		// ========== 构建测试（DS201-DS210） ==========

		Convey("DS201: 先创建PG catalog，再创建resource，最后构建", func() {
			// 1. 创建PG catalog
			catalogPayload := map[string]any{
				"name":           generateUniqueName("test-pg-catalog"),
				"description":    "测试PG catalog",
				"tags":           []string{"test", "pg", "catalog"},
				"connector_type": "postgresql",
				"connector_config": map[string]any{
					"host":     "192.168.36.54",
					"port":     5432,
					"username": "postgres",
					"password": "Password123",
					"database": "test",
				},
			}
			catalogResp := client.POST("/api/vega-backend/v1/catalogs", catalogPayload)
			So(catalogResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(catalogResp.Body["id"], ShouldNotBeEmpty)
			catalogID := catalogResp.Body["id"].(string)

			// 2. 创建resource，使用刚创建的catalog
			resourcePayload := map[string]any{
				"catalog_id":        catalogID,
				"name":              generateUniqueName("test-resource-build"),
				"tags":              []string{"test", "resource"},
				"description":       "测试资源构建",
				"category":          "table",
				"status":            "active",
				"database":          "test",
				"source_identifier": "public.users",
				"source_metadata": map[string]any{
					"primary_keys": []string{"id"},
				},
				"schema_definition": []map[string]any{
					{"name": "id", "type": "keyword", "display_name": "ID", "original_name": "id", "description": "唯一标识符"},
					{"name": "name", "type": "keyword", "display_name": "名称", "original_name": "name", "description": "用户名称"},
					{"name": "address", "type": "text", "display_name": "地址", "original_name": "address", "description": "用户地址"},
					{"name": "hobby", "type": "text", "display_name": "爱好", "original_name": "hobby", "description": "用户爱好"},
				},
			}
			resourceResp := client.POST("/api/vega-backend/v1/resources", resourcePayload)
			So(resourceResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resourceResp.Body["id"], ShouldNotBeEmpty)
			resourceID := resourceResp.Body["id"].(string)

			// 3. 创建构建任务
			buildResp := client.POST("/api/vega-backend/v1/build-tasks", map[string]any{"resource_id": resourceID, "mode": "streaming", "embedding_fields": "hobby", "embedding_model": "", "build_key_fields": "id"})
			So(buildResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 验证构建成功
			So(buildResp.Body, ShouldNotBeNil)

			// 4. 获取构建任务详情
			buildTaskResp := client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			So(buildTaskResp.Body, ShouldNotBeNil)
			So(buildTaskResp.Body["status"], ShouldEqual, "init")

			// 5. 启动任务
			startResp := client.POST("/api/vega-backend/v1/build-tasks/"+buildResp.Body["id"].(string)+"/start", nil)
			So(startResp.StatusCode, ShouldEqual, http.StatusOK)

			time.Sleep(10 * time.Second)

			// 6. 验证任务状态为running
			buildTaskResp = client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			So(buildTaskResp.Body, ShouldNotBeNil)
			So(buildTaskResp.Body["status"], ShouldEqual, "running")

			time.Sleep(10 * time.Second)

			// 7. 停止任务
			stopResp := client.POST("/api/vega-backend/v1/build-tasks/"+buildResp.Body["id"].(string)+"/stop", nil)
			So(stopResp.StatusCode, ShouldEqual, http.StatusOK)

			time.Sleep(10 * time.Second)

			// 8. 验证任务状态为stopped或者stopping
			buildTaskResp = client.GET("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(buildTaskResp.StatusCode, ShouldEqual, http.StatusOK)
			So(buildTaskResp.Body, ShouldNotBeNil)
			status := buildTaskResp.Body["status"].(string)
			So(status == "stopped" || status == "stopping", ShouldBeTrue)

			time.Sleep(5 * time.Second)

			// 9. 删除任务
			deleteResp := client.DELETE("/api/vega-backend/v1/build-tasks/" + buildResp.Body["id"].(string))
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

// cleanupCatalogs 清理现有catalog
func cleanupCatalogs(client *testutil.HTTPClient, t *testing.T) {
	resp := client.GET("/api/vega-backend/v1/catalogs?offset=0&limit=100")
	if resp.StatusCode == http.StatusOK {
		if entries, ok := resp.Body["entries"].([]any); ok {
			for _, entry := range entries {
				if entryMap, ok := entry.(map[string]any); ok {
					if id, ok := entryMap["id"].(string); ok {
						deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + id)
						if deleteResp.StatusCode != http.StatusNoContent {
							t.Logf("清理catalog失败 %s: %d", id, deleteResp.StatusCode)
						}
					}
				}
			}
		}
	}
}

// TestDeleteResourceWithBuildTask 测试删除有buildtask的resource时失败
// 测试编号前缀: DS2xx (Resource Build)
func TestDeleteResourceWithBuildTask(t *testing.T) {
	Convey("删除有buildtask的resource时失败 AT测试", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		cleanupResources(client, t)
		cleanupCatalogs(client, t)

		Convey("DS211: 删除有buildtask的resource时应该失败", func() {
			catalogPayload := map[string]any{
				"name":           generateUniqueName("test-mysql-catalog"),
				"description":    "测试mysql catalog",
				"tags":           []string{"test", "mysql", "catalog"},
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
			catalogID := catalogResp.Body["id"].(string)

			resourcePayload := map[string]any{
				"catalog_id":        catalogID,
				"name":              generateUniqueName("test-resource-delete"),
				"tags":              []string{"test", "resource"},
				"description":       "测试删除有buildtask的resource",
				"category":          "table",
				"status":            "active",
				"database":          "test",
				"source_identifier": "test.users",
				"schema_definition": []map[string]any{
					{"name": "id", "type": "keyword", "display_name": "ID", "original_name": "id", "description": "唯一标识符"},
					{"name": "name", "type": "keyword", "display_name": "名称", "original_name": "name", "description": "用户名称"},
					{"name": "address", "type": "text", "display_name": "地址", "original_name": "address", "description": "用户地址"},
					{"name": "hobby", "type": "text", "display_name": "爱好", "original_name": "hobby", "description": "用户爱好"},
				},
			}
			resourceResp := client.POST("/api/vega-backend/v1/resources", resourcePayload)
			So(resourceResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resourceResp.Body["id"], ShouldNotBeEmpty)
			resourceID := resourceResp.Body["id"].(string)

			buildResp := client.POST("/api/vega-backend/v1/build-tasks", map[string]any{"resource_id": resourceID, "mode": "batch", "embedding_fields": "", "embedding_model": "", "build_key_fields": "id"})
			So(buildResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(buildResp.Body["id"], ShouldNotBeEmpty)

			deleteResourceResp := client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
			So(deleteResourceResp.StatusCode, ShouldEqual, http.StatusBadRequest)

			taskID := buildResp.Body["id"].(string)
			deleteTaskResp := client.DELETE("/api/vega-backend/v1/build-tasks/" + taskID)
			So(deleteTaskResp.StatusCode, ShouldEqual, http.StatusNoContent)

			deleteResourceResp = client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
			So(deleteResourceResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

// TestDeleteRunningBuildTask 测试buildtask任务启动后删除失败
// 测试编号前缀: DS2xx (Resource Build)
func TestDeleteRunningBuildTask(t *testing.T) {
	Convey("buildtask任务启动后删除失败 AT测试", t, func() {
		var err error
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		cleanupResources(client, t)
		cleanupCatalogs(client, t)

		Convey("DS212: buildtask任务启动后删除应该失败", func() {
			catalogPayload := map[string]any{
				"name":           generateUniqueName("test-mysql-catalog"),
				"description":    "测试mysql catalog",
				"tags":           []string{"test", "mysql", "catalog"},
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
			catalogID := catalogResp.Body["id"].(string)

			resourcePayload := map[string]any{
				"catalog_id":        catalogID,
				"name":              generateUniqueName("test-resource-build-delete"),
				"tags":              []string{"test", "resource"},
				"description":       "测试删除running状态的buildtask",
				"category":          "table",
				"status":            "active",
				"database":          "test",
				"source_identifier": "test.users",
				"schema_definition": []map[string]any{
					{"name": "id", "type": "keyword", "display_name": "ID", "original_name": "id", "description": "唯一标识符"},
					{"name": "name", "type": "keyword", "display_name": "名称", "original_name": "name", "description": "用户名称"},
					{"name": "address", "type": "text", "display_name": "地址", "original_name": "address", "description": "用户地址"},
					{"name": "hobby", "type": "text", "display_name": "爱好", "original_name": "hobby", "description": "用户爱好"},
				},
			}
			resourceResp := client.POST("/api/vega-backend/v1/resources", resourcePayload)
			So(resourceResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resourceResp.Body["id"], ShouldNotBeEmpty)
			resourceID := resourceResp.Body["id"].(string)

			buildResp := client.POST("/api/vega-backend/v1/build-tasks", map[string]any{"resource_id": resourceID, "mode": "batch", "embedding_fields": "", "embedding_model": "", "build_key_fields": "id"})
			So(buildResp.StatusCode, ShouldEqual, http.StatusCreated)
			So(buildResp.Body["id"], ShouldNotBeEmpty)
			taskID := buildResp.Body["id"].(string)

			startResp := client.POST("/api/vega-backend/v1/build-tasks/"+taskID+"/start", nil)
			So(startResp.StatusCode, ShouldEqual, http.StatusOK)

			time.Sleep(5 * time.Second)

			deleteTaskResp := client.DELETE("/api/vega-backend/v1/build-tasks/" + taskID)
			So(deleteTaskResp.StatusCode, ShouldEqual, http.StatusBadRequest)

			stopResp := client.POST("/api/vega-backend/v1/build-tasks/"+taskID+"/stop", nil)
			So(stopResp.StatusCode, ShouldEqual, http.StatusOK)

			time.Sleep(10 * time.Second)

			deleteTaskResp = client.DELETE("/api/vega-backend/v1/build-tasks/" + taskID)
			So(deleteTaskResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}
