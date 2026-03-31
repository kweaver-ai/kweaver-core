// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mariadb

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend-tests/at/catalog/physical/table/mariadb"
	"vega-backend-tests/at/resource/helpers"
	"vega-backend-tests/at/setup"
	"vega-backend-tests/testutil"
)

// TestMariaDBResourceRead MariaDB Resource读取AT测试
// 编号规则：MR1xx
func TestMariaDBResourceRead(t *testing.T) {
	var (
		ctx        context.Context
		config     *setup.TestConfig
		client     *testutil.HTTPClient
		builder    *mariadb.MariaDBPayloadBuilder
		catalogID  string
		resourceID string
	)

	Convey("MariaDB Resource读取AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
		So(config.TargetMariaDB.Host, ShouldNotBeEmpty)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = mariadb.NewMariaDBPayloadBuilder(config.TargetMariaDB)
		builder.SetTestConfig(config)

		helpers.CleanupCatalogs(client, t)

		// ========== 准备测试数据 ==========
		// 1. 重置测试数据库（创建测试表）
		tableSuffix := fmt.Sprintf("%d", time.Now().Unix())
		testTableNames, _, err := builder.ResetTestDatabase(tableSuffix, 1, 0)
		So(err, ShouldBeNil)
		So(len(testTableNames), ShouldBeGreaterThan, 0)

		// 2. 创建 Catalog
		catalogPayload := builder.BuildCreatePayload()
		createResp := client.POST("/api/vega-backend/v1/catalogs", catalogPayload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		catalogID = createResp.Body["id"].(string)

		// 3. 触发 Discovery
		discoveryResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/discover", nil)
		So(discoveryResp.StatusCode, ShouldEqual, http.StatusOK)

		// 4. 轮询等待 Discovery 完成
		var taskID string
		if discoveryResp.Body != nil {
			if taskIDVal, ok := discoveryResp.Body["task_id"].(string); ok {
				taskID = taskIDVal
			}
		}

		if taskID != "" {
			maxAttempts := 60
			for attempt := 0; attempt < maxAttempts; attempt++ {
				taskResp := client.GET("/api/vega-backend/v1/tasks/" + taskID)
				if taskResp.StatusCode == http.StatusOK && taskResp.Body != nil {
					if status, ok := taskResp.Body["status"].(string); ok {
						if status == "completed" || status == "success" {
							break
						} else if status == "failed" || status == "error" {
							So(false, ShouldBeTrue, "Discovery task failed")
							break
						}
					}
				}
				time.Sleep(3 * time.Second)
			}
		}

		// 5. 获取发现的 Resource
		resourceResp := client.GET("/api/vega-backend/v1/resources?catalog_id=" + catalogID + "&offset=0&limit=100")
		So(resourceResp.StatusCode, ShouldEqual, http.StatusOK)
		So(resourceResp.Body, ShouldNotBeNil)

		entries := resourceResp.Body["entries"].([]any)
		So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)

		firstResource := entries[0].(map[string]any)
		resourceID = firstResource["id"].(string)

		// ========== 读取测试（MR101-MR108） ==========

		Convey("MR101: 获取存在的 MariaDB Resource", func() {
			resp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(resp.Body["id"], ShouldEqual, resourceID)
		})

		Convey("MR102: 获取不存在的 Resource", func() {
			resp := client.GET("/api/vega-backend/v1/resources/non-existent-resource-id")
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("MR103: 列表查询 - 按 catalog_id 过滤", func() {
			resp := client.GET("/api/vega-backend/v1/resources?catalog_id=" + catalogID + "&offset=0&limit=100")
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			if entries, ok := resp.Body["entries"].([]any); ok {
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)
				for _, entry := range entries {
					entryMap := entry.(map[string]any)
					So(entryMap["catalog_id"], ShouldEqual, catalogID)
				}
			}
		})

		Convey("MR104: 列表查询 - 按 category=table 过滤", func() {
			resp := client.GET("/api/vega-backend/v1/resources?category=table&offset=0&limit=100")
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			if entries, ok := resp.Body["entries"].([]any); ok {
				for _, entry := range entries {
					entryMap := entry.(map[string]any)
					So(entryMap["category"], ShouldEqual, "table")
				}
			}
		})

		Convey("MR105: 列表查询 - 按 connector_type=mariadb 过滤", func() {
			resp := client.GET("/api/vega-backend/v1/resources?connector_type=mariadb&offset=0&limit=100")
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			if entries, ok := resp.Body["entries"].([]any); ok {
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)
				for _, entry := range entries {
					entryMap := entry.(map[string]any)
					So(entryMap["connector_type"], ShouldEqual, "mariadb")
				}
			}
		})

		Convey("MR106: 列表分页测试", func() {
			resp := client.GET("/api/vega-backend/v1/resources?offset=0&limit=1")
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			if entries, ok := resp.Body["entries"].([]any); ok {
				So(len(entries), ShouldEqual, 1)
			}
		})

		Convey("MR107: 验证 Resource 包含完整的 schema 信息", func() {
			resp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			resource := resp.Body
			schemaDef, hasSchema := resource["schema_definition"]
			So(hasSchema, ShouldBeTrue)
			if schemaList, ok := schemaDef.([]any); ok {
				So(len(schemaList), ShouldBeGreaterThanOrEqualTo, 1)
			}
		})

		Convey("MR108: 验证 Resource 包含正确的 catalog 关联", func() {
			resp := client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			resource := resp.Body
			So(resource["catalog_id"], ShouldEqual, catalogID)
			So(resource["catalog_id"], ShouldNotBeEmpty)
		})
	})

	_ = ctx
}
