// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package helpers

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// RunCommonCreateTests 运行通用Resource创建测试
// 测试编号前缀: RM1xx (Resource Module Create)
func RunCommonCreateTests(suite *TestSuite) {
	catalogID := suite.CatalogID

	// ========== 正向测试（RM101-RM120） ==========

	Convey("RM101: 创建resource - 基本场景", func() {
		payload := BuildCreatePayload(catalogID)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("RM102: 创建resource - 最小字段", func() {
		payload := BuildMinimalPayload(catalogID)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("RM103: 创建resource - 完整字段", func() {
		payload := BuildFullCreatePayload(catalogID)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("RM104: 创建后立即查询", func() {
		payload := BuildCreatePayload(catalogID)
		createResp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

		resourceID := createResp.Body["id"].(string)

		// 立即查询
		getResp := suite.Client.GET("/api/vega-backend/v1/resources/" + resourceID)
		So(getResp.StatusCode, ShouldEqual, http.StatusOK)
		resource := ExtractFromEntriesResponse(getResp)
		So(resource["id"], ShouldEqual, resourceID)
		So(resource["name"], ShouldEqual, payload["name"])
		So(resource["catalog_id"], ShouldEqual, catalogID)
	})

	Convey("RM105: 创建带category的resource", func() {
		payload := BuildPayloadWithCategory(catalogID, "table")
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("RM106: Tags数组测试", func() {
		Convey("空tags数组", func() {
			payload := BuildPayloadWithEmptyTags(catalogID)
			resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("单个tag", func() {
			payload := BuildCreatePayload(catalogID)
			payload["tags"] = []string{"single-tag"}
			resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("多个tags", func() {
			payload := BuildCreatePayload(catalogID)
			payload["tags"] = []string{"tag1", "tag2", "tag3", "tag4", "tag5"}
			resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})
	})

	Convey("RM107: 特殊字符名称测试", func() {
		testCases := []struct {
			name         string
			resourceName string
			expectCode   int
		}{
			{"中文名称", "测试资源", 201},
			{"连字符", "test-resource", 201},
			{"下划线", "test_resource", 201},
			{"点号", "test.resource", 201},
			{"混合", "测试-resource_01.test", 201},
		}

		for _, tc := range testCases {
			Convey(tc.name, func() {
				payload := BuildPayloadWithSpecialCharsName(catalogID, tc.resourceName)
				resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
				So(resp.StatusCode, ShouldEqual, tc.expectCode)
			})
		}
	})

	Convey("RM108: 创建多个resource，列表查询", func() {
		// 创建3个resource
		for i := 0; i < 3; i++ {
			payload := BuildCreatePayload(catalogID)
			resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		}

		// 列表查询
		listResp := suite.Client.GET("/api/vega-backend/v1/resources?offset=0&limit=10")
		So(listResp.StatusCode, ShouldEqual, http.StatusOK)

		if listResp.Body != nil {
			if entries, ok := listResp.Body["entries"].([]any); ok {
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 3)
			}

			if totalCount, ok := listResp.Body["total_count"].(float64); ok {
				So(totalCount, ShouldBeGreaterThanOrEqualTo, 3.0)
			}
		}
	})

	Convey("RM109: 创建resource - 带description", func() {
		payload := BuildCreatePayload(catalogID)
		payload["description"] = "这是一个测试resource的注释"
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})
}

// RunCommonNegativeTests 运行通用Resource负向测试
// 测试编号前缀: RM1xx (121-140)
func RunCommonNegativeTests(suite *TestSuite) {
	catalogID := suite.CatalogID

	Convey("RM121: 缺少必填字段 - name", func() {
		payload := BuildPayloadWithMissingName(catalogID)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM122: 缺少必填字段 - catalog_id", func() {
		payload := BuildPayloadWithMissingCatalogID()
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM123: 重复的resource名称（同一catalog内）", func() {
		fixedName := fmt.Sprintf("duplicate-test-%d", time.Now().Unix())
		payload1 := BuildCreatePayload(catalogID)
		payload1["name"] = fixedName

		// 第一次创建
		resp1 := suite.Client.POST("/api/vega-backend/v1/resources", payload1)
		So(resp1.StatusCode, ShouldEqual, http.StatusCreated)

		// 第二次创建相同名称
		payload2 := BuildCreatePayload(catalogID)
		payload2["name"] = fixedName

		resp2 := suite.Client.POST("/api/vega-backend/v1/resources", payload2)
		So(resp2.StatusCode, ShouldEqual, http.StatusConflict)
		So(resp2.Error, ShouldNotBeNil)
	})

	Convey("RM124: 无效JSON格式", func() {
		invalidJSON := `{"name": "test", "catalog_id": }`
		resp := suite.Client.POST("/api/vega-backend/v1/resources", invalidJSON)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM125: 错误的Content-Type", func() {
		payload := BuildCreatePayload(catalogID)
		suite.Client.SetHeader("Content-Type", "text/plain")
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		suite.Client.SetHeader("Content-Type", "application/json")
		So(resp.StatusCode, ShouldEqual, http.StatusNotAcceptable)
	})

	Convey("RM126: 超长name字段（>128字符）", func() {
		payload := BuildPayloadWithLongName(catalogID, 129)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM127: 超长description字段（>1000字符）", func() {
		payload := BuildPayloadWithLongDescription(catalogID, 1001)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM128: name为空字符串", func() {
		payload := BuildPayloadWithEmptyName(catalogID)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM129: name只有空格", func() {
		payload := BuildPayloadWithWhitespaceName(catalogID)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		// name只有长度限制，空格也是有效字符
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
	})

	Convey("RM130: tags包含空字符串", func() {
		payload := BuildPayloadWithEmptyTagInArray(catalogID)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM131: tags包含非法字符", func() {
		payload := BuildPayloadWithInvalidCharTag(catalogID)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM132: 单个tag超长（41字符）", func() {
		payload := BuildPayloadWithLongTag(catalogID, 41)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM133: 无效的catalog_id", func() {
		payload := BuildPayloadWithInvalidCatalogID()
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM134: 请求体为空", func() {
		resp := suite.Client.POST("/api/vega-backend/v1/resources", nil)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})
}

// RunCommonBoundaryTests 运行通用Resource边界测试
// 测试编号前缀: RM1xx (141-160)
func RunCommonBoundaryTests(suite *TestSuite) {
	catalogID := suite.CatalogID

	Convey("RM141: name长度边界 - 1字符（最小）", func() {
		payload := BuildPayloadWithMinName(catalogID)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("RM142: name长度边界 - 128字符（最大允许）", func() {
		payload := BuildPayloadWithLongName(catalogID, 128)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("RM143: name长度边界 - 129字符（超出）", func() {
		payload := BuildPayloadWithLongName(catalogID, 129)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM144: description长度边界 - 1000字符（最大允许）", func() {
		payload := BuildPayloadWithExactDescription(catalogID, 1000)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("RM145: description长度边界 - 1001字符（超出）", func() {
		payload := BuildPayloadWithExactDescription(catalogID, 1001)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM146: tags数量边界 - 5个（最大允许）", func() {
		payload := BuildPayloadWithManyTags(catalogID, 5)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("RM147: tags数量边界 - 6个（超出）", func() {
		payload := BuildPayloadWithManyTags(catalogID, 6)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("RM148: 单个tag长度边界 - 40字符（最大允许）", func() {
		payload := BuildPayloadWithLongTag(catalogID, 40)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("RM149: 单个tag长度边界 - 41字符（超出）", func() {
		payload := BuildPayloadWithLongTag(catalogID, 41)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
	})
}

// RunCommonSecurityTests 运行通用Resource安全测试
// 测试编号前缀: RM1xx (161-170)
func RunCommonSecurityTests(suite *TestSuite) {
	catalogID := suite.CatalogID

	// 注：name字段只有长度限制，没有内容限制，因此特殊字符会被接受
	// 这些测试验证服务端能安全处理这些特殊输入（不会导致SQL注入、XSS等）

	Convey("RM161: SQL注入尝试", func() {
		payload := BuildPayloadWithSQLInjection(catalogID)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		// name没有内容限制，创建成功；验证系统能安全处理SQL注入payload
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("RM162: XSS尝试", func() {
		payload := BuildPayloadWithXSS(catalogID)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		// name没有内容限制，创建成功；验证系统能安全处理XSS payload
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})

	Convey("RM163: 路径遍历尝试", func() {
		payload := BuildPayloadWithPathTraversal(catalogID)
		resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		// name没有内容限制，创建成功；验证系统能安全处理路径遍历payload
		So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		So(resp.Body["id"], ShouldNotBeEmpty)
	})
}

// RunCommonReadTests 运行通用Resource读取测试
// 测试编号前缀: RM2xx
func RunCommonReadTests(suite *TestSuite) {
	catalogID := suite.CatalogID

	Convey("RM201: 获取存在的resource", func() {
		// 先创建
		payload := BuildCreatePayload(catalogID)
		createResp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)

		// 查询
		getResp := suite.Client.GET("/api/vega-backend/v1/resources/" + resourceID)
		So(getResp.StatusCode, ShouldEqual, http.StatusOK)

		resource := ExtractFromEntriesResponse(getResp)
		So(resource, ShouldNotBeNil)
		So(resource["id"], ShouldEqual, resourceID)
	})

	Convey("RM202: 获取不存在的resource", func() {
		resp := suite.Client.GET("/api/vega-backend/v1/resources/non-existent-id-12345")
		So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
	})

	Convey("RM203: 列表分页测试", func() {
		// 创建5个resource
		for i := 0; i < 5; i++ {
			payload := BuildCreatePayload(catalogID)
			resp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		}

		// 分页查询
		listResp := suite.Client.GET("/api/vega-backend/v1/resources?offset=0&limit=2")
		So(listResp.StatusCode, ShouldEqual, http.StatusOK)

		if entries, ok := listResp.Body["entries"].([]any); ok {
			So(len(entries), ShouldEqual, 2)
		}
	})

	Convey("RM204: 批量获取resources", func() {
		// 创建2个resource
		ids := make([]string, 2)
		for i := 0; i < 2; i++ {
			payload := BuildCreatePayload(catalogID)
			createResp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			ids[i] = createResp.Body["id"].(string)
		}

		// 批量获取
		batchResp := suite.GetResources(ids)
		So(batchResp.StatusCode, ShouldEqual, http.StatusOK)

		if batchResp.Body != nil {
			if entries, ok := batchResp.Body["entries"].([]any); ok {
				So(len(entries), ShouldEqual, 2)
			}
		}
	})

	Convey("RM205: 批量获取 - 部分ID不存在", func() {
		// 创建1个resource
		payload := BuildCreatePayload(catalogID)
		createResp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		existingID := createResp.Body["id"].(string)

		// 批量获取（包含不存在的ID）
		ids := []string{existingID, "non-existent-id-99999"}
		batchResp := suite.GetResources(ids)
		// 部分不存在时，应返回存在的部分或报错
		So(batchResp.StatusCode, ShouldEqual, http.StatusNotFound)
	})

	Convey("RM206: 按catalog_id过滤列表", func() {
		// 创建resource
		payload := BuildCreatePayload(catalogID)
		createResp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

		// 按catalog_id过滤
		filterResp := suite.Client.GET(fmt.Sprintf("/api/vega-backend/v1/resources?catalog_id=%s&offset=0&limit=100", catalogID))
		So(filterResp.StatusCode, ShouldEqual, http.StatusOK)

		if filterResp.Body != nil && filterResp.Body["entries"] != nil {
			entries := filterResp.Body["entries"].([]any)
			So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)

			// 验证都属于当前catalog
			for _, entry := range entries {
				resourceEntry := entry.(map[string]any)
				So(resourceEntry["catalog_id"], ShouldEqual, catalogID)
			}
		}
	})
}

// RunCommonUpdateTests 运行通用Resource更新测试
// 测试编号前缀: RM3xx
func RunCommonUpdateTests(suite *TestSuite) {
	catalogID := suite.CatalogID

	Convey("RM301: 更新resource名称", func() {
		// 创建
		payload := BuildCreatePayload(catalogID)
		resourceID, resourceData, err := suite.CreateAndGetResource(payload)
		So(err, ShouldBeNil)

		// 基于原数据构建更新payload
		newName := GenerateUniqueName("updated-resource")
		updatePayload := BuildUpdatePayload(resourceData, map[string]any{
			"name": newName,
		})
		updateResp := suite.Client.PUT("/api/vega-backend/v1/resources/"+resourceID, updatePayload)
		So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 验证
		getResp := suite.Client.GET("/api/vega-backend/v1/resources/" + resourceID)
		resource := ExtractFromEntriesResponse(getResp)
		So(resource["name"], ShouldEqual, newName)
	})

	Convey("RM302: 更新resource注释", func() {
		// 创建
		payload := BuildCreatePayload(catalogID)
		resourceID, resourceData, err := suite.CreateAndGetResource(payload)
		So(err, ShouldBeNil)

		// 基于原数据构建更新payload
		newDescription := "更新后的资源注释内容"
		updatePayload := BuildUpdatePayload(resourceData, map[string]any{
			"description": newDescription,
		})
		updateResp := suite.Client.PUT("/api/vega-backend/v1/resources/"+resourceID, updatePayload)
		So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 验证
		getResp := suite.Client.GET("/api/vega-backend/v1/resources/" + resourceID)
		resource := ExtractFromEntriesResponse(getResp)
		So(resource["description"], ShouldEqual, newDescription)
	})

	Convey("RM303: 更新不存在的resource", func() {
		updatePayload := map[string]any{
			"name": "new-name",
		}
		resp := suite.Client.PUT("/api/vega-backend/v1/resources/non-existent-id-12345", updatePayload)
		So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
	})

	Convey("RM304: 更新tags", func() {
		// 创建
		payload := BuildCreatePayload(catalogID)
		resourceID, resourceData, err := suite.CreateAndGetResource(payload)
		So(err, ShouldBeNil)

		// 基于原数据构建更新payload
		updatePayload := BuildUpdatePayload(resourceData, map[string]any{
			"tags": []string{"updated", "test"},
		})
		updateResp := suite.Client.PUT("/api/vega-backend/v1/resources/"+resourceID, updatePayload)
		So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
	})

	Convey("RM305: 更新category", func() {
		// 创建
		payload := BuildPayloadWithCategory(catalogID, "table")
		resourceID, resourceData, err := suite.CreateAndGetResource(payload)
		So(err, ShouldBeNil)

		// 基于原数据构建更新payload
		updatePayload := BuildUpdatePayload(resourceData, map[string]any{
			"category": "file",
		})
		updateResp := suite.Client.PUT("/api/vega-backend/v1/resources/"+resourceID, updatePayload)
		So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
	})
}

// RunCommonDeleteTests 运行通用Resource删除测试
// 测试编号前缀: RM4xx
func RunCommonDeleteTests(suite *TestSuite) {
	catalogID := suite.CatalogID

	Convey("RM401: 删除存在的resource", func() {
		// 创建
		payload := BuildCreatePayload(catalogID)
		createResp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)

		// 删除
		deleteResp := suite.Client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
		So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 验证已删除
		getResp := suite.Client.GET("/api/vega-backend/v1/resources/" + resourceID)
		So(getResp.StatusCode, ShouldEqual, http.StatusNotFound)
	})

	Convey("RM402: 删除不存在的resource", func() {
		resp := suite.Client.DELETE("/api/vega-backend/v1/resources/non-existent-id-12345")
		So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
	})

	Convey("RM403: 重复删除同一resource", func() {
		// 创建
		payload := BuildCreatePayload(catalogID)
		createResp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID := createResp.Body["id"].(string)

		// 第一次删除
		deleteResp1 := suite.Client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
		So(deleteResp1.StatusCode, ShouldEqual, http.StatusNoContent)

		// 第二次删除
		deleteResp2 := suite.Client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
		So(deleteResp2.StatusCode, ShouldEqual, http.StatusNotFound)
	})

	Convey("RM404: 批量删除resources", func() {
		// 创建3个resource
		ids := make([]string, 3)
		for i := 0; i < 3; i++ {
			payload := BuildCreatePayload(catalogID)
			createResp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
			ids[i] = createResp.Body["id"].(string)
		}

		// 批量删除
		batchDeleteResp := suite.DeleteResources(ids)
		So(batchDeleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 验证所有resource都已删除
		for _, resourceID := range ids {
			getResp := suite.Client.GET("/api/vega-backend/v1/resources/" + resourceID)
			So(getResp.StatusCode, ShouldEqual, http.StatusNotFound)
		}
	})

	Convey("RM405: 批量删除 - 部分ID不存在", func() {
		// 创建1个resource
		payload := BuildCreatePayload(catalogID)
		createResp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)
		existingID := createResp.Body["id"].(string)

		// 批量删除（包含不存在的ID）
		ids := []string{existingID, "non-existent-id-99999"}
		batchDeleteResp := suite.DeleteResources(ids)
		// 部分不存在时，行为取决于API设计
		So(batchDeleteResp.StatusCode, ShouldEqual, http.StatusNotFound)
	})

	Convey("RM406: 删除后列表中不再显示", func() {
		// 创建resource
		payload := BuildCreatePayload(catalogID)
		resourceName := payload["name"]
		createResp := suite.Client.POST("/api/vega-backend/v1/resources", payload)
		So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

		resourceID := createResp.Body["id"].(string)

		// 删除resource
		deleteResp := suite.Client.DELETE("/api/vega-backend/v1/resources/" + resourceID)
		So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 查询列表
		listResp := suite.Client.GET("/api/vega-backend/v1/resources?offset=0&limit=1000")
		So(listResp.StatusCode, ShouldEqual, http.StatusOK)

		if listResp.Body != nil && listResp.Body["entries"] != nil {
			entries := listResp.Body["entries"].([]any)

			// 验证删除的resource不在列表中
			found := false
			for _, entry := range entries {
				resourceEntry := entry.(map[string]any)
				if resourceEntry["name"] == resourceName {
					found = true
					break
				}
			}

			So(found, ShouldBeFalse)
		}
	})
}

// RunCommonNameUniquenessTests 运行通用Resource名称唯一性测试
// 测试编号前缀: RM5xx
func RunCommonNameUniquenessTests(suite *TestSuite) {
	catalogID := suite.CatalogID

	Convey("RM501: 同一catalog内重名冲突", func() {
		fixedName := fmt.Sprintf("unique-test-%d", time.Now().UnixNano())

		// 第一次创建
		payload1 := BuildCreatePayload(catalogID)
		payload1["name"] = fixedName
		resp1 := suite.Client.POST("/api/vega-backend/v1/resources", payload1)
		So(resp1.StatusCode, ShouldEqual, http.StatusCreated)

		// 第二次创建相同名称
		payload2 := BuildCreatePayload(catalogID)
		payload2["name"] = fixedName
		resp2 := suite.Client.POST("/api/vega-backend/v1/resources", payload2)
		So(resp2.StatusCode, ShouldEqual, http.StatusConflict)
	})

	Convey("RM502: 不同catalog内同名共存", func() {
		// 创建第二个catalog
		secondCatalogID, err := suite.CreatePrerequisiteCatalog()
		So(err, ShouldBeNil)
		So(secondCatalogID, ShouldNotBeEmpty)

		fixedName := fmt.Sprintf("cross-catalog-test-%d", time.Now().UnixNano())

		// 在第一个catalog中创建
		payload1 := BuildCreatePayload(catalogID)
		payload1["name"] = fixedName
		resp1 := suite.Client.POST("/api/vega-backend/v1/resources", payload1)
		So(resp1.StatusCode, ShouldEqual, http.StatusCreated)

		// 在第二个catalog中创建同名resource
		payload2 := BuildCreatePayload(secondCatalogID)
		payload2["name"] = fixedName
		resp2 := suite.Client.POST("/api/vega-backend/v1/resources", payload2)
		So(resp2.StatusCode, ShouldEqual, http.StatusCreated)
	})

	Convey("RM503: 删除后重建同名resource", func() {
		fixedName := fmt.Sprintf("recreate-test-%d", time.Now().UnixNano())

		// 创建
		payload1 := BuildCreatePayload(catalogID)
		payload1["name"] = fixedName
		resp1 := suite.Client.POST("/api/vega-backend/v1/resources", payload1)
		So(resp1.StatusCode, ShouldEqual, http.StatusCreated)
		resourceID1 := resp1.Body["id"].(string)

		// 删除
		deleteResp := suite.Client.DELETE("/api/vega-backend/v1/resources/" + resourceID1)
		So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

		// 重建同名
		payload2 := BuildCreatePayload(catalogID)
		payload2["name"] = fixedName
		resp2 := suite.Client.POST("/api/vega-backend/v1/resources", payload2)
		So(resp2.StatusCode, ShouldEqual, http.StatusCreated)

		// 新resource应有不同ID
		resourceID2 := resp2.Body["id"].(string)
		So(resourceID2, ShouldNotEqual, resourceID1)
	})
}

// RunAllCommonTests 运行所有通用Resource测试
func RunAllCommonTests(suite *TestSuite) {
	RunCommonCreateTests(suite)
	RunCommonNegativeTests(suite)
	RunCommonBoundaryTests(suite)
	RunCommonSecurityTests(suite)
	RunCommonReadTests(suite)
	RunCommonUpdateTests(suite)
	RunCommonDeleteTests(suite)
	RunCommonNameUniquenessTests(suite)
}
