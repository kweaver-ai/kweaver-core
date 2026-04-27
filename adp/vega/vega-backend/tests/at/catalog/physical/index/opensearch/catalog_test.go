// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package opensearch

import (
	"context"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend-tests/at/catalog/helpers"
	"vega-backend-tests/at/setup"
	"vega-backend-tests/testutil"
)

// TestOpenSearchCatalogCreate OpenSearch Catalog创建AT测试
// 编号规则：OS1xx
func TestOpenSearchCatalogCreate(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *OpenSearchPayloadBuilder
	)

	Convey("OpenSearch Catalog创建AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
		So(config.TargetOpenSearch.Host, ShouldNotBeEmpty)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		builder = NewOpenSearchPayloadBuilder(config.TargetOpenSearch)
		builder.SetTestConfig(config)

		helpers.CleanupCatalogs(client, t)

		// ========== 正向测试（OS101-OS108） ==========

		Convey("OS101: 创建 OpenSearch catalog - 基本场景", func() {
			payload := builder.BuildCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("OS102: 创建后验证 connector_type 为 opensearch", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := helpers.ExtractFromEntriesResponse(getResp)
			So(catalog["connector_type"], ShouldEqual, "opensearch")
		})

		Convey("OS103: 创建后验证 type 为 physical", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := helpers.ExtractFromEntriesResponse(getResp)
			So(catalog["type"], ShouldEqual, helpers.CatalogTypePhysical)
		})

		Convey("OS104: 创建 OpenSearch catalog - 完整字段", func() {
			payload := builder.BuildFullCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)

			catalogID := resp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := helpers.ExtractFromEntriesResponse(getResp)
			So(catalog["description"], ShouldNotBeEmpty)
			tags, ok := catalog["tags"].([]any)
			So(ok, ShouldBeTrue)
			So(len(tags), ShouldBeGreaterThan, 0)
		})

		Convey("OS105: 创建带 SSL 配置的 catalog（SSL 禁用）", func() {
			payload := builder.BuildCreatePayloadWithSSL(false)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("OS106: 创建后立即查询", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := helpers.ExtractFromEntriesResponse(getResp)
			So(catalog["id"], ShouldEqual, catalogID)
			So(catalog["name"], ShouldEqual, payload["name"])
		})

		Convey("OS107: OpenSearch 连接测试成功", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("OS108: 获取 OpenSearch catalog 健康状态", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			statusResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
			So(statusResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		// ========== connector_config 负向测试（OS121-OS129） ==========

		Convey("OS121: 缺少 host 字段", func() {
			payload := map[string]any{
				"name":           helpers.GenerateUniqueName("missing-host"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"port":     config.TargetOpenSearch.Port,
					"username": config.TargetOpenSearch.Username,
					"password": config.TargetOpenSearch.Password,
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS122: 缺少 port 字段", func() {
			payload := map[string]any{
				"name":           helpers.GenerateUniqueName("missing-port"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     config.TargetOpenSearch.Host,
					"username": config.TargetOpenSearch.Username,
					"password": config.TargetOpenSearch.Password,
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS123: 缺少认证信息（无 username）", func() {
			payload := builder.BuildCreatePayloadWithMissingAuth()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS124: 空用户名", func() {
			payload := map[string]any{
				"name":           helpers.GenerateUniqueName("empty-username"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     config.TargetOpenSearch.Host,
					"port":     config.TargetOpenSearch.Port,
					"username": "",
					"password": config.TargetOpenSearch.Password,
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS125: 错误凭证", func() {
			payload := builder.BuildCreatePayloadWithWrongCredentials()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS126: 无效端口（非数字）", func() {
			payload := builder.BuildCreatePayloadWithInvalidPort()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS127: 超出范围端口（65536）", func() {
			payload := builder.BuildCreatePayloadWithOutOfRangePort()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS128: 负数端口", func() {
			payload := map[string]any{
				"name":           helpers.GenerateUniqueName("negative-port"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     config.TargetOpenSearch.Host,
					"port":     -1,
					"username": config.TargetOpenSearch.Username,
					"password": config.TargetOpenSearch.Password,
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS129: 无效 host", func() {
			payload := builder.BuildCreatePayloadWithInvalidHost()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		// ========== 边界测试（OS131-OS137） ==========

		Convey("OS131: port 边界值（1）", func() {
			payload := map[string]any{
				"name":           helpers.GenerateUniqueName("port-1"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     config.TargetOpenSearch.Host,
					"port":     1,
					"username": config.TargetOpenSearch.Username,
					"password": builder.GetEncryptedPassword(),
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("OS132: port 边界值（65535）", func() {
			payload := map[string]any{
				"name":           helpers.GenerateUniqueName("port-65535"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     config.TargetOpenSearch.Host,
					"port":     65535,
					"username": config.TargetOpenSearch.Username,
					"password": builder.GetEncryptedPassword(),
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("OS133: host 为 IP 地址", func() {
			payload := builder.BuildCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("OS134: host 为域名", func() {
			payload := map[string]any{
				"name":           helpers.GenerateUniqueName("domain-host"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     "localhost",
					"port":     config.TargetOpenSearch.Port,
					"username": config.TargetOpenSearch.Username,
					"password": builder.GetEncryptedPassword(),
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("OS135: password 为空", func() {
			payload := map[string]any{
				"name":           helpers.GenerateUniqueName("empty-password"),
				"connector_type": "opensearch",
				"connector_config": map[string]any{
					"host":     config.TargetOpenSearch.Host,
					"port":     config.TargetOpenSearch.Port,
					"username": config.TargetOpenSearch.Username,
					"password": "",
					"use_ssl":  config.TargetOpenSearch.UseSSL,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS136: 使用 HTTPS 协议", func() {
			payload := builder.BuildCreatePayloadWithSSL(true)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("OS137: 使用 HTTP 协议", func() {
			payload := builder.BuildCreatePayloadWithSSL(false)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})
	})

	_ = ctx
}

// TestOpenSearchCatalogRead OpenSearch Catalog读取AT测试
// 编号规则：OS2xx
func TestOpenSearchCatalogRead(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *OpenSearchPayloadBuilder
	)

	Convey("OpenSearch Catalog读取AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = NewOpenSearchPayloadBuilder(config.TargetOpenSearch)
		builder.SetTestConfig(config)

		helpers.CleanupCatalogs(client, t)

		// ========== 读取测试（OS201-OS205） ==========

		Convey("OS201: 获取存在的 OpenSearch catalog", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := helpers.ExtractFromEntriesResponse(getResp)
			So(catalog["id"], ShouldEqual, catalogID)
		})

		Convey("OS202: 列表查询 - 按 type 过滤 physical", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			listResp := client.GET("/api/vega-backend/v1/catalogs?type=physical&offset=0&limit=100")
			So(listResp.StatusCode, ShouldEqual, http.StatusOK)

			if entries, ok := listResp.Body["entries"].([]any); ok {
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)
				for _, entry := range entries {
					So(entry.(map[string]any)["type"], ShouldEqual, "physical")
				}
			}
		})

		Convey("OS203: 列表查询 - 按 connector_type 过滤 opensearch", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			listResp := client.GET("/api/vega-backend/v1/catalogs?connector_type=opensearch&offset=0&limit=100")
			So(listResp.StatusCode, ShouldEqual, http.StatusOK)

			if entries, ok := listResp.Body["entries"].([]any); ok {
				So(len(entries), ShouldBeGreaterThanOrEqualTo, 1)
				for _, entry := range entries {
					So(entry.(map[string]any)["connector_type"], ShouldEqual, "opensearch")
				}
			}
		})

		Convey("OS204: 查询 catalog - 验证所有字段返回", func() {
			payload := builder.BuildFullCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := helpers.ExtractFromEntriesResponse(getResp)
			So(catalog["id"], ShouldNotBeEmpty)
			So(catalog["name"], ShouldNotBeEmpty)
			So(catalog["type"], ShouldEqual, helpers.CatalogTypePhysical)
			So(catalog["connector_type"], ShouldEqual, "opensearch")
			So(catalog["create_time"], ShouldNotBeZeroValue)
			So(catalog["update_time"], ShouldNotBeZeroValue)
		})

		Convey("OS205: 验证 connector_config.password 不返回", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := helpers.ExtractFromEntriesResponse(getResp)
			if connCfg, ok := catalog["connector_config"].(map[string]any); ok {
				_, hasPassword := connCfg["password"]
				So(hasPassword, ShouldBeFalse)
			}
		})
	})

	_ = ctx
}

// TestOpenSearchCatalogUpdate OpenSearch Catalog更新AT测试
// 编号规则：OS3xx
func TestOpenSearchCatalogUpdate(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *OpenSearchPayloadBuilder
	)

	Convey("OpenSearch Catalog更新AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = NewOpenSearchPayloadBuilder(config.TargetOpenSearch)
		builder.SetTestConfig(config)

		helpers.CleanupCatalogs(client, t)

		// ========== 更新测试（OS301-OS305） ==========

		Convey("OS301: 整体更新 connector_config", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := helpers.ExtractFromEntriesResponse(getResp)

			osConfig := builder.GetConfig()
			updatePayload := helpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":     osConfig.Host,
					"port":     osConfig.Port,
					"username": osConfig.Username,
					"password": builder.GetEncryptedPassword(),
					"use_ssl":  osConfig.UseSSL,
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("OS302: 更新 connector_config 后连接测试", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := helpers.ExtractFromEntriesResponse(getResp)

			osConfig := builder.GetConfig()
			updatePayload := helpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":     osConfig.Host,
					"port":     osConfig.Port,
					"username": osConfig.Username,
					"password": builder.GetEncryptedPassword(),
					"use_ssl":  osConfig.UseSSL,
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("OS303: 更新 host 为无效地址", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := helpers.ExtractFromEntriesResponse(getResp)

			osConfig := builder.GetConfig()
			updatePayload := helpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":     "invalid-host-12345.example.com",
					"port":     osConfig.Port,
					"username": osConfig.Username,
					"password": builder.GetEncryptedPassword(),
					"use_ssl":  osConfig.UseSSL,
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS304: 更新 port 为无效值", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := helpers.ExtractFromEntriesResponse(getResp)

			osConfig := builder.GetConfig()
			updatePayload := helpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":     osConfig.Host,
					"port":     65536,
					"username": osConfig.Username,
					"password": builder.GetEncryptedPassword(),
					"use_ssl":  osConfig.UseSSL,
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("OS305: 更新 password", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := helpers.ExtractFromEntriesResponse(getResp)

			osConfig := builder.GetConfig()
			updatePayload := helpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":     osConfig.Host,
					"port":     osConfig.Port,
					"username": osConfig.Username,
					"password": builder.GetEncryptedPassword(),
					"use_ssl":  osConfig.UseSSL,
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})

	_ = ctx
}

// TestOpenSearchCatalogDelete OpenSearch Catalog删除AT测试
// 编号规则：OS4xx
func TestOpenSearchCatalogDelete(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *OpenSearchPayloadBuilder
	)

	Convey("OpenSearch Catalog删除AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = NewOpenSearchPayloadBuilder(config.TargetOpenSearch)
		builder.SetTestConfig(config)

		helpers.CleanupCatalogs(client, t)

		// ========== 删除测试（OS401-OS402） ==========

		Convey("OS401: 删除 OpenSearch catalog 后健康状态不可查", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			statusResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
			So(statusResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("OS402: 删除 OpenSearch catalog 后不能测试连接", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)

			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})
	})

	_ = ctx
}

// TestOpenSearchSpecificOptions OpenSearch特有选项测试
// 编号规则：OS5xx
func TestOpenSearchSpecificOptions(t *testing.T) {
	var (
		ctx     context.Context
		config  *setup.TestConfig
		client  *testutil.HTTPClient
		builder *OpenSearchPayloadBuilder
	)

	Convey("OpenSearch特有选项AT测试 - 初始化", t, func() {
		ctx = context.Background()

		var err error
		config, err = setup.LoadTestConfig()
		So(err, ShouldBeNil)

		client = testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		builder = NewOpenSearchPayloadBuilder(config.TargetOpenSearch)
		builder.SetTestConfig(config)

		helpers.CleanupCatalogs(client, t)

		// ========== OpenSearch特有选项测试（OS501-OS504） ==========

		Convey("OS501: SSL/TLS 连接测试（跳过验证）", func() {
			options := map[string]any{"insecure_skip_verify": true}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("OS502: 自定义 index pattern 选项", func() {
			options := map[string]any{"index_pattern": "*"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("OS503: 连接超时选项测试", func() {
			options := map[string]any{"timeout": "30s"}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("OS504: 多节点配置测试（如支持）", func() {
			options := map[string]any{"max_retries": 3}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})
	})

	_ = ctx
}
