// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logical

import (
	"net/http"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	cataloghelpers "vega-backend-tests/at/catalog/helpers"
	logicalhelpers "vega-backend-tests/at/catalog/logical/helpers"
	"vega-backend-tests/at/setup"
	"vega-backend-tests/testutil"
)

// TestLogicalCatalogCreate Logical Catalog创建AT测试
// 测试编号前缀: LG1xx (Logical Create)
func TestLogicalCatalogCreate(t *testing.T) {

	Convey("Logical Catalog创建AT测试 - 初始化", t, func() {

		// 加载测试配置
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		// 创建HTTP客户端
		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)

		// 验证服务可用性
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		// 清理现有logical catalog
		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 正向测试（LG101-LG120） ==========

		Convey("LG101: 创建logical catalog - 基本场景", func() {
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("LG102: 创建logical catalog - 完整字段", func() {
			payload := logicalhelpers.BuildFullLogicalCatalogPayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("LG103: 创建后验证type为logical", func() {
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 查询验证
			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog, ShouldNotBeNil)
			So(catalog["type"], ShouldEqual, cataloghelpers.CatalogTypeLogical)
		})

		Convey("LG104: 创建后立即查询", func() {
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 立即查询
			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)
			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["id"], ShouldEqual, catalogID)
			So(catalog["name"], ShouldEqual, payload["name"])
		})

		Convey("LG105: 创建多个logical catalog，列表查询", func() {
			// 创建3个logical catalog
			for i := 0; i < 3; i++ {
				payload := logicalhelpers.BuildLogicalCatalogPayload()
				resp := client.POST("/api/vega-backend/v1/catalogs", payload)
				So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			}

			// 列表查询（按type过滤）
			listResp := client.GET("/api/vega-backend/v1/catalogs?type=logical&offset=0&limit=10")
			So(listResp.StatusCode, ShouldEqual, http.StatusOK)

			entries, ok := listResp.Body["entries"].([]any)
			So(ok, ShouldBeTrue)
			So(len(entries), ShouldEqual, 3)
		})

		Convey("LG106: logical catalog无connector_type", func() {
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 查询验证
			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog, ShouldNotBeNil)
			// logical catalog没有connector_type
			So(catalog["connector_type"], ShouldEqual, "")
		})

		// ========== 负向测试（LG121-LG127） ==========

		Convey("LG121: 重复的catalog名称", func() {
			fixedName := cataloghelpers.GenerateUniqueName("duplicate-logical")
			payload1 := logicalhelpers.BuildLogicalCatalogPayloadWithName(fixedName)

			// 第一次创建
			resp1 := client.POST("/api/vega-backend/v1/catalogs", payload1)
			So(resp1.StatusCode, ShouldEqual, http.StatusCreated)

			// 第二次创建相同名称
			payload2 := logicalhelpers.BuildLogicalCatalogPayloadWithName(fixedName)
			resp2 := client.POST("/api/vega-backend/v1/catalogs", payload2)
			So(resp2.StatusCode, ShouldEqual, http.StatusConflict)
		})

		Convey("LG122: 缺少必填字段 - name", func() {
			payload := map[string]any{
				"type":        "logical",
				"description": "缺少name字段",
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("LG123: 空字符串name", func() {
			payload := map[string]any{
				"name":        "",
				"type":        "logical",
				"description": "空字符串name",
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("LG124: name超过最大长度（255字符）", func() {
			payload := map[string]any{
				"name":        cataloghelpers.BuildStringWithLength("a", 256),
				"type":        "logical",
				"description": "超长name",
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("LG125: description超过最大长度（1000字符）", func() {
			payload := map[string]any{
				"name":        cataloghelpers.GenerateUniqueName("long-desc"),
				"type":        "logical",
				"description": cataloghelpers.BuildStringWithLength("d", 1001),
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("LG126: 单个tag超过最大长度（40字符）", func() {
			payload := map[string]any{
				"name":        cataloghelpers.GenerateUniqueName("long-tag"),
				"type":        "logical",
				"description": "单个tag超长",
				"tags":        []string{cataloghelpers.BuildStringWithLength("t", 41)},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("LG127: tags数量超过限制（最大5个）", func() {
			// 最多 5 个 tags，这里创建 6 个
			var tags []string
			for i := 0; i < 6; i++ {
				tags = append(tags, cataloghelpers.GenerateUniqueName("tag"))
			}
			payload := map[string]any{
				"name":        cataloghelpers.GenerateUniqueName("many-tags"),
				"type":        "logical",
				"description": "tags数量超限",
				"tags":        tags,
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("LG128: tag包含无效字符", func() {
			// 无效字符: /:?\"<>|：？''""！《》,#[]{}%&*$^!=.'
			invalidTags := []string{
				"tag/with/slash",
				"tag:with:colon",
				"tag?question",
				"tag<angle>",
				"tag#hash",
				"tag[bracket]",
				"tag{brace}",
				"tag%percent",
				"tag&ampersand",
				"tag*asterisk",
				"tag$dollar",
				"tag=equals",
				"tag.dot",
			}
			for _, invalidTag := range invalidTags {
				payload := map[string]any{
					"name":        cataloghelpers.GenerateUniqueName("invalid-tag"),
					"type":        "logical",
					"description": "tag包含无效字符",
					"tags":        []string{invalidTag},
				}
				resp := client.POST("/api/vega-backend/v1/catalogs", payload)
				So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
			}
		})

		// ========== 边界测试（LG131-LG139） ==========

		Convey("LG131: name最大长度（255字符）", func() {
			payload := map[string]any{
				"name":        cataloghelpers.BuildUniqueStringWithLength(255),
				"type":        "logical",
				"description": "name达到最大长度",
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("LG132: description最大长度（1000字符）", func() {
			payload := map[string]any{
				"name":        cataloghelpers.GenerateUniqueName("max-desc"),
				"type":        "logical",
				"description": cataloghelpers.BuildStringWithLength("d", 1000),
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("LG133: name包含中文", func() {
			payload := map[string]any{
				"name":        cataloghelpers.GenerateUniqueName("中文测试"),
				"type":        "logical",
				"description": "中文name测试",
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("LG134: name包含特殊字符（下划线、连字符）", func() {
			payload := map[string]any{
				"name":        cataloghelpers.GenerateUniqueName("test_name-special"),
				"type":        "logical",
				"description": "特殊字符name测试",
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("LG135: tags为空数组", func() {
			payload := map[string]any{
				"name":        cataloghelpers.GenerateUniqueName("empty-tags"),
				"type":        "logical",
				"description": "空tags数组",
				"tags":        []string{},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("LG136: tags包含最大数量（5个）", func() {
			// 最多 5 个 tags
			tags := []string{"tag1", "tag2", "tag3", "tag4", "tag5"}
			payload := map[string]any{
				"name":        cataloghelpers.GenerateUniqueName("max-tags"),
				"type":        "logical",
				"description": "达到tags最大数量",
				"tags":        tags,
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("LG137: 单个tag最大长度（40字符）", func() {
			payload := map[string]any{
				"name":        cataloghelpers.GenerateUniqueName("max-tag-len"),
				"type":        "logical",
				"description": "单个tag达到最大长度",
				"tags":        []string{cataloghelpers.BuildStringWithLength("t", 40)},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("LG138: description为空字符串", func() {
			payload := map[string]any{
				"name":        cataloghelpers.GenerateUniqueName("empty-desc"),
				"type":        "logical",
				"description": "",
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("LG139: logical catalog不需要connector_config", func() {
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}

// TestLogicalCatalogRead Logical Catalog读取AT测试
// 测试编号前缀: LG2xx
func TestLogicalCatalogRead(t *testing.T) {

	Convey("Logical Catalog读取AT测试 - 初始化", t, func() {

		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 读取测试（LG201-LG210） ==========

		Convey("LG201: 获取存在的logical catalog", func() {
			// 先创建
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 查询
			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog, ShouldNotBeNil)
			So(catalog["id"], ShouldEqual, catalogID)
		})

		Convey("LG202: 获取不存在的catalog", func() {
			resp := client.GET("/api/vega-backend/v1/catalogs/non-existent-id-12345")
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("LG203: 列表查询 - 按type过滤logical", func() {
			// 创建1个logical catalog
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 查询logical类型
			logicalResp := client.GET("/api/vega-backend/v1/catalogs?type=logical&offset=0&limit=100")
			So(logicalResp.StatusCode, ShouldEqual, http.StatusOK)

			entries := logicalResp.Body["entries"].([]any)
			So(len(entries), ShouldEqual, 1)

			// 验证都是logical类型
			for _, entry := range entries {
				So(entry.(map[string]any)["type"], ShouldEqual, "logical")
			}
		})

		Convey("LG204: 列表分页测试", func() {
			// 创建5个logical catalog
			for i := 0; i < 5; i++ {
				payload := logicalhelpers.BuildLogicalCatalogPayload()
				resp := client.POST("/api/vega-backend/v1/catalogs", payload)
				So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			}

			// 分页查询
			listResp := client.GET("/api/vega-backend/v1/catalogs?type=logical&offset=0&limit=2")
			So(listResp.StatusCode, ShouldEqual, http.StatusOK)
			So(len(listResp.Body["entries"].([]any)), ShouldEqual, 2)
		})

		Convey("LG205: 列表查询 - 按name模糊搜索", func() {
			// 创建带特定前缀的logical catalog
			searchPrefix := cataloghelpers.GenerateUniqueName("search-test")
			payload := logicalhelpers.BuildLogicalCatalogPayloadWithName(searchPrefix)
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 模糊搜索
			searchResp := client.GET("/api/vega-backend/v1/catalogs?name=" + searchPrefix[:20] + "&offset=0&limit=10")
			So(searchResp.StatusCode, ShouldEqual, http.StatusOK)
			So(len(searchResp.Body["entries"].([]any)), ShouldEqual, 1)
		})

		Convey("LG206: 列表查询 - 空结果", func() {
			// 使用不存在的type查询
			listResp := client.GET("/api/vega-backend/v1/catalogs?name=non-existent-catalog-xyz-12345&offset=0&limit=10")
			So(listResp.StatusCode, ShouldEqual, http.StatusOK)
			So(len(listResp.Body["entries"].([]any)), ShouldEqual, 0)
		})
	})
}

// TestLogicalCatalogUpdate Logical Catalog更新AT测试
// 测试编号前缀: LG3xx
func TestLogicalCatalogUpdate(t *testing.T) {

	Convey("Logical Catalog更新AT测试 - 初始化", t, func() {

		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 更新测试（LG301-LG310） ==========

		Convey("LG301: 更新logical catalog名称", func() {
			// 创建
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 获取原始数据
			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalogData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			// 基于原数据构建更新payload
			newName := cataloghelpers.GenerateUniqueName("updated-logical")
			updatePayload := cataloghelpers.BuildUpdatePayload(catalogData, map[string]any{
				"name": newName,
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证
			verifyResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := cataloghelpers.ExtractFromEntriesResponse(verifyResp)
			So(catalog["name"], ShouldEqual, newName)
		})

		Convey("LG302: 更新logical catalog描述", func() {
			// 创建
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			catalogID := createResp.Body["id"].(string)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalogData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			// 基于原数据构建更新payload
			newDescription := "更新后的逻辑Catalog描述"
			updatePayload := cataloghelpers.BuildUpdatePayload(catalogData, map[string]any{
				"description": newDescription,
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证
			verifyResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := cataloghelpers.ExtractFromEntriesResponse(verifyResp)
			So(catalog["description"], ShouldEqual, newDescription)
		})

		Convey("LG303: 更新不存在的catalog", func() {
			updatePayload := map[string]any{
				"name": "new-name",
			}
			resp := client.PUT("/api/vega-backend/v1/catalogs/non-existent-id-12345", updatePayload)
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("LG304: 验证update_time更新", func() {
			// 创建
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			catalogID := createResp.Body["id"].(string)

			// 等待1秒确保时间戳不同
			time.Sleep(1 * time.Second)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)
			originalUpdateTime := originalData["update_time"].(float64)

			// 更新
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"description": "验证update_time更新",
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证update_time已更新
			verifyResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			newData := cataloghelpers.ExtractFromEntriesResponse(verifyResp)
			newUpdateTime := newData["update_time"].(float64)
			So(newUpdateTime, ShouldBeGreaterThan, originalUpdateTime)
		})

		Convey("LG305: 更新tags", func() {
			// 创建
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			catalogID := createResp.Body["id"].(string)

			// 获取原始数据
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalogData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			// 更新tags
			newTags := []string{"updated-tag-1", "updated-tag-2", "updated-tag-3"}
			updatePayload := cataloghelpers.BuildUpdatePayload(catalogData, map[string]any{
				"tags": newTags,
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证
			verifyResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := cataloghelpers.ExtractFromEntriesResponse(verifyResp)
			So(len(catalog["tags"].([]any)), ShouldEqual, 3)
		})

		Convey("LG306: 更新为已存在的name", func() {
			// 创建两个catalog
			payload1 := logicalhelpers.BuildLogicalCatalogPayload()
			createResp1 := client.POST("/api/vega-backend/v1/catalogs", payload1)
			So(createResp1.StatusCode, ShouldEqual, http.StatusCreated)
			existingName := payload1["name"].(string)

			payload2 := logicalhelpers.BuildLogicalCatalogPayload()
			createResp2 := client.POST("/api/vega-backend/v1/catalogs", payload2)
			So(createResp2.StatusCode, ShouldEqual, http.StatusCreated)
			catalogID2 := createResp2.Body["id"].(string)

			// 获取第二个catalog的数据
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID2)
			catalogData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			// 尝试更新为已存在的name
			updatePayload := cataloghelpers.BuildUpdatePayload(catalogData, map[string]any{
				"name": existingName,
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID2, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusConflict)
		})
	})
}

// TestLogicalCatalogDelete Logical Catalog删除AT测试
// 测试编号前缀: LG4xx
func TestLogicalCatalogDelete(t *testing.T) {

	Convey("Logical Catalog删除AT测试 - 初始化", t, func() {

		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 删除测试（LG401-LG410） ==========

		Convey("LG401: 删除存在的logical catalog", func() {
			// 创建
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// 删除
			catalogID := createResp.Body["id"].(string)
			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 验证已删除
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("LG402: 删除不存在的catalog", func() {
			resp := client.DELETE("/api/vega-backend/v1/catalogs/non-existent-id-12345")
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("LG403: 重复删除同一catalog", func() {
			// 创建
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			catalogID := createResp.Body["id"].(string)

			// 第一次删除
			deleteResp1 := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp1.StatusCode, ShouldEqual, http.StatusNoContent)

			// 第二次删除
			deleteResp2 := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp2.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("LG404: 删除后可以创建同名catalog", func() {
			// 创建
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			catalogName := payload["name"]
			createResp1 := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp1.StatusCode, ShouldEqual, http.StatusCreated)

			// 删除
			catalogID1 := createResp1.Body["id"].(string)
			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID1)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			// 创建同名catalog
			payload2 := logicalhelpers.BuildLogicalCatalogPayloadWithName(catalogName.(string))
			createResp2 := client.POST("/api/vega-backend/v1/catalogs", payload2)
			So(createResp2.StatusCode, ShouldEqual, http.StatusCreated)

			// 新创建的catalog应该有不同的ID
			catalogID2 := createResp2.Body["id"].(string)
			So(catalogID2, ShouldNotEqual, catalogID1)
		})
	})
}

// TestLogicalCatalogSpecific Logical Catalog特有测试
// 测试编号前缀: LG5xx
func TestLogicalCatalogSpecific(t *testing.T) {

	Convey("Logical Catalog特有AT测试 - 初始化", t, func() {

		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== Logical特有测试（LG501-LG502） ==========

		Convey("LG501: logical catalog测试连接", func() {
			// 创建logical catalog
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// logical catalog测试连接永远返回200 OK
			catalogID := createResp.Body["id"].(string)
			testConnResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testConnResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("LG502: logical catalog健康检查", func() {
			// 创建logical catalog
			payload := logicalhelpers.BuildLogicalCatalogPayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			// logical catalog健康检查永远返回200 OK
			catalogID := createResp.Body["id"].(string)
			healthResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
			So(healthResp.StatusCode, ShouldEqual, http.StatusOK)
		})
	})
}
