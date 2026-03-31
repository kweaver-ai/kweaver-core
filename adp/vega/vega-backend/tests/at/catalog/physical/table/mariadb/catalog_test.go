// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mariadb

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	cataloghelpers "vega-backend-tests/at/catalog/helpers"
	"vega-backend-tests/at/setup"
	"vega-backend-tests/testutil"
)

// TestMariaDBCatalogCreate MariaDB Catalog创建AT测试
// 编号规则：MD1xx
func TestMariaDBCatalogCreate(t *testing.T) {

	Convey("MariaDB Catalog创建AT测试 - 初始化", t, func() {

		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
		So(config.TargetMariaDB.Host, ShouldNotBeEmpty)

		builder := NewMariaDBPayloadBuilder(config.TargetMariaDB)
		builder.SetTestConfig(config)

		tableSize := 1
		recordSize := 0
		tableSuffix := fmt.Sprintf("%d", time.Now().Unix())
		testTableName, records, err := builder.ResetTestDatabase(tableSuffix, tableSize, recordSize)
		So(err, ShouldBeNil)
		So(testTableName, ShouldNotBeEmpty)
		So(len(records), ShouldEqual, recordSize)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ AT测试环境就绪，VEGA Manager: %s", config.VegaBackend.BaseURL)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 正向测试（MD101-MD110） ==========
		Convey("MD101: 创建MariaDB catalog - 基本场景", func() {
			payload := builder.BuildCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
			So(resp.Body["id"], ShouldNotBeEmpty)
		})

		Convey("MD102: 创建后验证connector_type为mariadb", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["connector_type"], ShouldEqual, "mariadb")
		})

		Convey("MD103: 创建后验证type为physical", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["type"], ShouldEqual, cataloghelpers.CatalogTypePhysical)
		})

		Convey("MD104: 创建MariaDB catalog - 完整字段", func() {
			payload := builder.BuildFullCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := resp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["description"], ShouldNotBeEmpty)
			So(len(catalog["tags"].([]any)), ShouldBeGreaterThan, 0)
		})

		Convey("MD105: 创建带MariaDB特定options（charset/timeout）", func() {
			options := map[string]any{
				"charset": "utf8mb4",
				"timeout": "10s",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD106: 创建后立即查询", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["id"], ShouldEqual, catalogID)
			So(catalog["name"], ShouldEqual, payload["name"])
		})

		Convey("MD107: MariaDB连接测试成功", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("MD108: 获取MariaDB catalog健康状态", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			statusResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
			So(statusResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("MD109: 创建实例级MariaDB catalog（不指定database）", func() {
			payload := builder.BuildCreatePayloadWithoutDatabase()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD110: 实例级MariaDB catalog连接测试成功", func() {
			payload := builder.BuildCreatePayloadWithoutDatabase()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		// ========== connector_config负向测试（MD121-MD129） ==========

		Convey("MD121: 缺少host字段", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("missing-host"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD122: 缺少port字段", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("missing-port"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD123: 缺少user字段", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("missing-user"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD124: 空用户名", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("empty-user"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  "",
					"password":  mariadbConfig.Password,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD125: 错误密码", func() {
			payload := builder.BuildCreatePayloadWithWrongCredentials()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD126: 不存在的数据库", func() {
			payload := builder.BuildCreatePayloadWithNonExistentDB()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD127: 无效端口（非数字）", func() {
			payload := builder.BuildCreatePayloadWithInvalidPort()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD128: 超出范围端口（65536）", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("overflow-port"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      65536,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  mariadbConfig.Password,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD129: 负数端口", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("negative-port"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      -1,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  mariadbConfig.Password,
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		// ========== 边界测试（MD131-MD138） ==========

		Convey("MD131: port边界值（1）", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("port-1"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      1,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MD132: port边界值（65535）", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("port-65535"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      65535,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MD133: database名称最大长度（64字符）", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("long-db"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{strings.Repeat("d", 64)},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MD134: database名称超过最大长度", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("too-long-db"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{strings.Repeat("d", 65)},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD135: host为IP地址", func() {
			payload := builder.BuildCreatePayload()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD136: host为域名", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("domain-host"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      "localhost",
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MD137: 不指定database（实例级连接）", func() {
			payload := builder.BuildCreatePayloadWithoutDatabase()
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD138: password为空（无密码连接）", func() {
			mariadbConfig := builder.GetConfig()
			payload := map[string]any{
				"name":           cataloghelpers.GenerateUniqueName("no-password"),
				"connector_type": "mariadb",
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  "",
				},
			}
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})
	})
}

// TestMariaDBCatalogRead MariaDB Catalog读取AT测试
// 编号规则：MD2xx
func TestMariaDBCatalogRead(t *testing.T) {

	Convey("MariaDB Catalog读取AT测试 - 初始化", t, func() {

		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
		So(config.TargetMariaDB.Host, ShouldNotBeEmpty)

		builder := NewMariaDBPayloadBuilder(config.TargetMariaDB)
		builder.SetTestConfig(config)

		tableSize := 1
		recordSize := 0
		tableSuffix := fmt.Sprintf("%d", time.Now().Unix())
		testTableName, records, err := builder.ResetTestDatabase(tableSuffix, tableSize, recordSize)
		So(err, ShouldBeNil)
		So(testTableName, ShouldNotBeEmpty)
		So(len(records), ShouldEqual, recordSize)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 读取测试（MD201-MD205） ==========
		Convey("MD201: 获取存在的MariaDB catalog", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["id"], ShouldEqual, catalogID)
		})

		Convey("MD202: 列表查询 - 按type过滤physical", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			listResp := client.GET("/api/vega-backend/v1/catalogs?type=physical&offset=0&limit=100")
			So(listResp.StatusCode, ShouldEqual, http.StatusOK)

			entries, ok := listResp.Body["entries"].([]any)
			So(ok, ShouldBeTrue)
			So(len(entries), ShouldEqual, 1)
			for _, entry := range entries {
				So(entry.(map[string]any)["type"], ShouldEqual, "physical")
			}
		})

		Convey("MD203: 列表查询 - 按connector_type过滤mariadb", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			listResp := client.GET("/api/vega-backend/v1/catalogs?connector_type=mariadb&offset=0&limit=100")
			So(listResp.StatusCode, ShouldEqual, http.StatusOK)

			entries, ok := listResp.Body["entries"].([]any)
			So(ok, ShouldBeTrue)
			So(len(entries), ShouldEqual, 1)
			for _, entry := range entries {
				So(entry.(map[string]any)["connector_type"], ShouldEqual, "mariadb")
			}
		})

		Convey("MD204: 查询catalog - 验证所有字段返回", func() {
			payload := builder.BuildFullCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			So(catalog["id"], ShouldNotBeEmpty)
			So(catalog["name"], ShouldNotBeEmpty)
			So(catalog["type"], ShouldEqual, cataloghelpers.CatalogTypePhysical)
			So(catalog["connector_type"], ShouldEqual, "mariadb")
			So(catalog["create_time"], ShouldNotBeZeroValue)
			So(catalog["update_time"], ShouldNotBeZeroValue)
		})

		Convey("MD205: 验证connector_config.password不返回", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			So(getResp.StatusCode, ShouldEqual, http.StatusOK)

			catalog := cataloghelpers.ExtractFromEntriesResponse(getResp)
			connCfg, ok := catalog["connector_config"].(map[string]any)
			So(ok, ShouldBeTrue)
			_, hasPassword := connCfg["password"]
			So(hasPassword, ShouldBeFalse)
		})
	})
}

// TestMariaDBCatalogUpdate MariaDB Catalog更新AT测试
// 编号规则：MD3xx
func TestMariaDBCatalogUpdate(t *testing.T) {

	Convey("MariaDB Catalog更新AT测试 - 初始化", t, func() {

		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
		So(config.TargetMariaDB.Host, ShouldNotBeEmpty)

		builder := NewMariaDBPayloadBuilder(config.TargetMariaDB)
		builder.SetTestConfig(config)

		tableSize := 1
		recordSize := 0
		tableSuffix := fmt.Sprintf("%d", time.Now().Unix())
		testTableName, records, err := builder.ResetTestDatabase(tableSuffix, tableSize, recordSize)
		So(err, ShouldBeNil)
		So(testTableName, ShouldNotBeEmpty)
		So(len(records), ShouldEqual, recordSize)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 更新测试（MD301-MD305） ==========
		Convey("MD301: 整体更新connector_config", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mariadbConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
					"options": map[string]any{
						"charset": "utf8mb4",
					},
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})

		Convey("MD302: 更新connector_config后连接测试", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mariadbConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)

			testResp := client.POST("/api/vega-backend/v1/catalogs/"+catalogID+"/test-connection", nil)
			So(testResp.StatusCode, ShouldEqual, http.StatusOK)
		})

		Convey("MD303: 更新host为无效地址", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mariadbConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      "invalid-host-12345.example.com",
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD304: 更新port为无效值", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mariadbConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      65536,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("MD305: 更新password", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			getResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID)
			originalData := cataloghelpers.ExtractFromEntriesResponse(getResp)

			mariadbConfig := builder.GetConfig()
			updatePayload := cataloghelpers.BuildUpdatePayload(originalData, map[string]any{
				"connector_config": map[string]any{
					"host":      mariadbConfig.Host,
					"port":      mariadbConfig.Port,
					"databases": []string{mariadbConfig.Database},
					"username":  mariadbConfig.Username,
					"password":  builder.GetEncryptedPassword(),
				},
			})
			updateResp := client.PUT("/api/vega-backend/v1/catalogs/"+catalogID, updatePayload)
			So(updateResp.StatusCode, ShouldEqual, http.StatusNoContent)
		})
	})
}

// TestMariaDBCatalogDelete MariaDB Catalog删除AT测试
// 编号规则：MD4xx
func TestMariaDBCatalogDelete(t *testing.T) {

	Convey("MariaDB Catalog删除AT测试 - 初始化", t, func() {

		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
		So(config.TargetMariaDB.Host, ShouldNotBeEmpty)

		builder := NewMariaDBPayloadBuilder(config.TargetMariaDB)
		builder.SetTestConfig(config)

		tableSize := 1
		recordSize := 0
		tableSuffix := fmt.Sprintf("%d", time.Now().Unix())
		testTableName, records, err := builder.ResetTestDatabase(tableSuffix, tableSize, recordSize)
		So(err, ShouldBeNil)
		So(testTableName, ShouldNotBeEmpty)
		So(len(records), ShouldEqual, recordSize)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== 删除测试（MD401-MD402） ==========
		Convey("MD401: 删除MariaDB catalog后健康状态不可查", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			deleteResp := client.DELETE("/api/vega-backend/v1/catalogs/" + catalogID)
			So(deleteResp.StatusCode, ShouldEqual, http.StatusNoContent)

			statusResp := client.GET("/api/vega-backend/v1/catalogs/" + catalogID + "/health-status")
			So(statusResp.StatusCode, ShouldEqual, http.StatusNotFound)
		})

		Convey("MD402: 删除MariaDB catalog后不能测试连接", func() {
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
}

// TestMariaDBCatalogDiscover MariaDB Catalog Discover AT测试
// 编号规则：MD5xx
// 注意：Discover测试需要先初始化待发现的库表，然后对比Discover结果与实际表结构
func TestMariaDBCatalogDiscover(t *testing.T) {

	Convey("MariaDB Catalog Discover AT测试 - 初始化", t, func() {

		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
		So(config.TargetMariaDB.Host, ShouldNotBeEmpty)

		builder := NewMariaDBPayloadBuilder(config.TargetMariaDB)
		builder.SetTestConfig(config)

		tableSize := 1
		recordSize := 0
		tableSuffix := fmt.Sprintf("%d", time.Now().Unix())
		testTableName, records, err := builder.ResetTestDatabase(tableSuffix, tableSize, recordSize)
		So(err, ShouldBeNil)
		So(testTableName, ShouldNotBeEmpty)
		So(len(records), ShouldEqual, recordSize)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cataloghelpers.CleanupCatalogs(client, t)
		cataloghelpers.CleanupResources(client, t)

		tableSchema := []any{
			map[string]any{"description": "", "display_name": "c_int", "name": "c_int", "original_name": "c_int", "type": "integer"},
			map[string]any{"description": "", "display_name": "c_int2", "name": "c_int2", "original_name": "c_int2", "type": "integer"},
			map[string]any{"description": "", "display_name": "c_bigint", "name": "c_bigint", "original_name": "c_bigint", "type": "integer"},
			map[string]any{"description": "", "display_name": "c_bigint2", "name": "c_bigint2", "original_name": "c_bigint2", "type": "integer"},
			map[string]any{"description": "", "display_name": "c_smallint", "name": "c_smallint", "original_name": "c_smallint", "type": "integer"}, map[string]any{"description": "", "display_name": "c_smallint2", "name": "c_smallint2", "original_name": "c_smallint2", "type": "integer"},
			map[string]any{"description": "", "display_name": "c_tinyint", "name": "c_tinyint", "original_name": "c_tinyint", "type": "integer"},
			map[string]any{"description": "", "display_name": "c_tinyint2", "name": "c_tinyint2", "original_name": "c_tinyint2", "type": "integer"},
			map[string]any{"description": "", "display_name": "c_mediumint", "name": "c_mediumint", "original_name": "c_mediumint", "type": "integer"},
			map[string]any{"description": "", "display_name": "c_mediumint2", "name": "c_mediumint2", "original_name": "c_mediumint2", "type": "integer"},
			map[string]any{"description": "", "display_name": "c_decimal", "name": "c_decimal", "original_name": "c_decimal", "type": "decimal"},
			map[string]any{"description": "", "display_name": "c_float", "name": "c_float", "original_name": "c_float", "type": "float"},
			map[string]any{"description": "", "display_name": "c_double", "name": "c_double", "original_name": "c_double", "type": "float"},
			map[string]any{"description": "", "display_name": "c_char", "name": "c_char", "original_name": "c_char", "type": "string"},
			map[string]any{"description": "", "display_name": "c_varchar", "name": "c_varchar", "original_name": "c_varchar", "type": "string"},
			map[string]any{"description": "", "display_name": "c_text", "name": "c_text", "original_name": "c_text", "type": "text"},
			map[string]any{"description": "", "display_name": "c_mediumtext", "name": "c_mediumtext", "original_name": "c_mediumtext", "type": "text"},
			map[string]any{"description": "", "display_name": "c_longtext", "name": "c_longtext", "original_name": "c_longtext", "type": "text"},
			map[string]any{"description": "", "display_name": "c_date", "name": "c_date", "original_name": "c_date", "type": "date"},
			map[string]any{"description": "", "display_name": "c_time", "name": "c_time", "original_name": "c_time", "type": "time"},
			map[string]any{"description": "", "display_name": "c_datetime", "name": "c_datetime", "original_name": "c_datetime", "type": "datetime"},
			map[string]any{"description": "", "display_name": "c_datetime2", "name": "c_datetime2", "original_name": "c_datetime2", "type": "datetime"},
			map[string]any{"description": "", "display_name": "c_timestamp", "name": "c_timestamp", "original_name": "c_timestamp", "type": "datetime"},
			map[string]any{"description": "", "display_name": "c_timestamp2", "name": "c_timestamp2", "original_name": "c_timestamp2", "type": "datetime"},
			map[string]any{"description": "", "display_name": "c_year", "name": "c_year", "original_name": "c_year", "type": "integer"},
			map[string]any{"description": "", "display_name": "c_binary", "name": "c_binary", "original_name": "c_binary", "type": "binary"},
			map[string]any{"description": "", "display_name": "c_varbinary", "name": "c_varbinary", "original_name": "c_varbinary", "type": "binary"},
			map[string]any{"description": "", "display_name": "c_blob", "name": "c_blob", "original_name": "c_blob", "type": "binary"},
			map[string]any{"description": "", "display_name": "c_longblob", "name": "c_longblob", "original_name": "c_longblob", "type": "binary"},
			map[string]any{"description": "", "display_name": "c_bit", "name": "c_bit", "original_name": "c_bit", "type": "boolean"},
			map[string]any{"description": "", "display_name": "c_bool", "name": "c_bool", "original_name": "c_bool", "type": "integer"},
			map[string]any{"description": "", "display_name": "c_boolean", "name": "c_boolean", "original_name": "c_boolean", "type": "integer"},
			map[string]any{"description": "", "display_name": "c_null", "name": "c_null", "original_name": "c_null", "type": "string"},
			map[string]any{"description": "", "display_name": "c_not_null", "name": "c_not_null", "original_name": "c_not_null", "type": "string"},
			map[string]any{"description": "", "display_name": "c_default", "name": "c_default", "original_name": "c_default", "type": "string"},
			map[string]any{"description": "这是注释", "display_name": "c_comment", "name": "c_comment", "original_name": "c_comment", "type": "string"},
			map[string]any{"description": "", "display_name": "c_collate", "name": "c_collate", "original_name": "c_collate", "type": "string"},
		}

		tableSourceMeta := map[string]any{
			"columns": []any{
				map[string]any{"name": "c_int", "type": "integer", "orig_type": "int(11)", "nullable": false, "description": "", "num_precision": 10, "ordinal_position": 1, "column_key": "PRI"},
				map[string]any{"name": "c_int2", "type": "integer", "orig_type": "int(11)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 10, "ordinal_position": 2, "column_key": "UNI"},
				map[string]any{"name": "c_bigint", "type": "integer", "orig_type": "bigint(20)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 19, "ordinal_position": 3, "column_key": "MUL"},
				map[string]any{"name": "c_bigint2", "type": "integer", "orig_type": "bigint(20)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 19, "ordinal_position": 4, "column_key": ""},
				map[string]any{"name": "c_smallint", "type": "integer", "orig_type": "smallint(6)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 5, "ordinal_position": 5, "column_key": "MUL"},
				map[string]any{"name": "c_smallint2", "type": "integer", "orig_type": "smallint(6)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 5, "ordinal_position": 6, "column_key": ""},
				map[string]any{"name": "c_tinyint", "type": "integer", "orig_type": "tinyint(4)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 3, "ordinal_position": 7, "column_key": ""},
				map[string]any{"name": "c_tinyint2", "type": "integer", "orig_type": "tinyint(4)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 3, "ordinal_position": 8, "column_key": ""},
				map[string]any{"name": "c_mediumint", "type": "integer", "orig_type": "mediumint(9)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 7, "ordinal_position": 9, "column_key": ""},
				map[string]any{"name": "c_mediumint2", "type": "integer", "orig_type": "mediumint(8)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 7, "ordinal_position": 10, "column_key": ""},
				map[string]any{"name": "c_decimal", "type": "decimal", "orig_type": "decimal(10,2)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 10, "num_scale": 2, "ordinal_position": 11, "column_key": ""},
				map[string]any{"name": "c_float", "type": "float", "orig_type": "float", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 12, "ordinal_position": 12, "column_key": ""},
				map[string]any{"name": "c_double", "type": "float", "orig_type": "double", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 22, "ordinal_position": 13, "column_key": ""},
				map[string]any{"name": "c_char", "type": "string", "orig_type": "char(10)", "nullable": true, "default_value": "NULL", "description": "", "char_max_len": 10, "charset": "utf8mb4", "collation": "utf8mb4_unicode_ci", "ordinal_position": 14, "column_key": ""},
				map[string]any{"name": "c_varchar", "type": "string", "orig_type": "varchar(255)", "nullable": true, "default_value": "NULL", "description": "", "char_max_len": 255, "charset": "utf8mb4", "collation": "utf8mb4_unicode_ci", "ordinal_position": 15, "column_key": ""},
				map[string]any{"name": "c_text", "type": "text", "orig_type": "text", "nullable": true, "default_value": "NULL", "description": "", "char_max_len": 65535, "charset": "utf8mb4", "collation": "utf8mb4_unicode_ci", "ordinal_position": 16, "column_key": ""},
				map[string]any{"name": "c_mediumtext", "type": "text", "orig_type": "mediumtext", "nullable": true, "default_value": "NULL", "description": "", "char_max_len": 16777215, "charset": "utf8mb4", "collation": "utf8mb4_unicode_ci", "ordinal_position": 17, "column_key": ""},
				map[string]any{"name": "c_longtext", "type": "text", "orig_type": "longtext", "nullable": true, "default_value": "NULL", "description": "", "char_max_len": 4294967295, "charset": "utf8mb4", "collation": "utf8mb4_unicode_ci", "ordinal_position": 18, "column_key": ""},
				map[string]any{"name": "c_date", "type": "date", "orig_type": "date", "nullable": true, "default_value": "NULL", "description": "", "ordinal_position": 19, "column_key": ""},
				map[string]any{"name": "c_time", "type": "time", "orig_type": "time", "nullable": true, "default_value": "NULL", "description": "", "ordinal_position": 20, "column_key": ""},
				map[string]any{"name": "c_datetime", "type": "datetime", "orig_type": "datetime", "nullable": true, "default_value": "NULL", "description": "", "ordinal_position": 21, "column_key": ""},
				map[string]any{"name": "c_datetime2", "type": "datetime", "orig_type": "datetime(6)", "nullable": true, "default_value": "NULL", "description": "", "datetime_precision": 6, "ordinal_position": 22, "column_key": ""},
				map[string]any{"name": "c_timestamp", "type": "datetime", "orig_type": "timestamp", "nullable": true, "default_value": "current_timestamp()", "description": "", "ordinal_position": 23, "column_key": ""},
				map[string]any{"name": "c_timestamp2", "type": "datetime", "orig_type": "timestamp(6)", "nullable": true, "default_value": "current_timestamp(6)", "description": "", "datetime_precision": 6, "ordinal_position": 24, "column_key": ""},
				map[string]any{"name": "c_year", "type": "integer", "orig_type": "year(4)", "nullable": true, "default_value": "NULL", "description": "", "ordinal_position": 25, "column_key": ""},
				map[string]any{"name": "c_binary", "type": "binary", "orig_type": "binary(16)", "nullable": true, "default_value": "NULL", "description": "", "char_max_len": 16, "ordinal_position": 26, "column_key": ""},
				map[string]any{"name": "c_varbinary", "type": "binary", "orig_type": "varbinary(255)", "nullable": true, "default_value": "NULL", "description": "", "char_max_len": 255, "ordinal_position": 27, "column_key": ""},
				map[string]any{"name": "c_blob", "type": "binary", "orig_type": "blob", "nullable": true, "default_value": "NULL", "description": "", "char_max_len": 65535, "ordinal_position": 28, "column_key": ""},
				map[string]any{"name": "c_longblob", "type": "binary", "orig_type": "longblob", "nullable": true, "default_value": "NULL", "description": "", "char_max_len": 4294967295, "ordinal_position": 29, "column_key": ""},
				map[string]any{"name": "c_bit", "type": "boolean", "orig_type": "bit(8)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 8, "ordinal_position": 30, "column_key": ""},
				map[string]any{"name": "c_bool", "type": "integer", "orig_type": "tinyint(1)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 3, "ordinal_position": 31, "column_key": ""},
				map[string]any{"name": "c_boolean", "type": "integer", "orig_type": "tinyint(1)", "nullable": true, "default_value": "NULL", "description": "", "num_precision": 3, "ordinal_position": 32, "column_key": ""},
				map[string]any{"name": "c_null", "type": "string", "orig_type": "varchar(20)", "nullable": true, "default_value": "NULL", "description": "", "char_max_len": 20, "charset": "utf8mb4", "collation": "utf8mb4_unicode_ci", "ordinal_position": 33, "column_key": ""},
				map[string]any{"name": "c_not_null", "type": "string", "orig_type": "varchar(20)", "nullable": false, "description": "", "char_max_len": 20, "charset": "utf8mb4", "collation": "utf8mb4_unicode_ci", "ordinal_position": 34, "column_key": ""},
				map[string]any{"name": "c_default", "type": "string", "orig_type": "varchar(20)", "nullable": true, "default_value": "'default_value'", "description": "", "char_max_len": 20, "charset": "utf8mb4", "collation": "utf8mb4_unicode_ci", "ordinal_position": 35, "column_key": ""},
				map[string]any{"name": "c_comment", "type": "string", "orig_type": "varchar(20)", "nullable": true, "default_value": "NULL", "description": "这是注释", "char_max_len": 20, "charset": "utf8mb4", "collation": "utf8mb4_unicode_ci", "ordinal_position": 36, "column_key": ""},
				map[string]any{"name": "c_collate", "type": "string", "orig_type": "varchar(20)", "nullable": true, "default_value": "NULL", "description": "", "char_max_len": 20, "charset": "utf8mb4", "collation": "utf8mb4_unicode_ci", "ordinal_position": 37, "column_key": ""},
			},
			"indices": []map[string]any{
				{"name": "idx_c_bigint", "columns": []any{"c_bigint"}, "unique": false, "primary": false},
				{"name": "idx_multi", "columns": []any{"c_smallint", "c_smallint2"}, "unique": false, "primary": false},
				{"name": "PRIMARY", "columns": []any{"c_int"}, "unique": true, "primary": true},
				{"name": "uk_c_int2", "columns": []any{"c_int2"}, "unique": true, "primary": false},
			},
			"primary_keys": []any{"c_int"},
			"properties": map[string]any{
				"charset":      "utf8mb4",
				"collation":    "utf8mb4_unicode_ci",
				"create_time":  1771056294000,
				"data_length":  16384,
				"engine":       "InnoDB",
				"index_length": 49152,
				"row_count":    0,
			},
			"table_type": "table",
		}
		_ = tableSourceMeta

		// ========== Discover 正向测试（MD501-MD510） ==========
		Convey("MD501: 触发Discover - 基本场景", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			err = builder.RunDiscoverTask(client, catalogID)
			So(err, ShouldBeNil)
		})

		Convey("MD502: Discover后验证Resource存在", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			err = builder.RunDiscoverTask(client, catalogID)
			So(err, ShouldBeNil)

			resourceResp := client.GET("/api/vega-backend/v1/resources?catalog_id=" + catalogID + "&offset=0&limit=100")
			So(resourceResp.StatusCode, ShouldEqual, http.StatusOK)
			entries, ok := resourceResp.Body["entries"].([]any)
			So(ok, ShouldBeTrue)
			So(len(entries), ShouldEqual, tableSize)
		})

		Convey("MD503: 验证发现的Resource的category", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			err = builder.RunDiscoverTask(client, catalogID)
			So(err, ShouldBeNil)

			resourceResp := client.GET("/api/vega-backend/v1/resources?catalog_id=" + catalogID + "&offset=0&limit=100")
			So(resourceResp.StatusCode, ShouldEqual, http.StatusOK)
			entries, ok := resourceResp.Body["entries"].([]any)
			So(ok, ShouldBeTrue)
			So(len(entries), ShouldEqual, tableSize)

			for _, entry := range entries {
				resource := entry.(map[string]any)
				So(resource["category"], ShouldEqual, "table")
			}
		})

		Convey("MD504: Discover后验证Resource与Catalog关联", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			err = builder.RunDiscoverTask(client, catalogID)
			So(err, ShouldBeNil)

			resourceResp := client.GET("/api/vega-backend/v1/resources?catalog_id=" + catalogID + "&offset=0&limit=100")
			So(resourceResp.StatusCode, ShouldEqual, http.StatusOK)
			entries, ok := resourceResp.Body["entries"].([]any)
			So(ok, ShouldBeTrue)
			So(len(entries), ShouldEqual, tableSize)

			for _, entry := range entries {
				resource := entry.(map[string]any)
				So(resource["catalog_id"], ShouldEqual, catalogID)
			}
		})

		Convey("MD505: 验证发现的Resource的schema_definition", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			err = builder.RunDiscoverTask(client, catalogID)
			So(err, ShouldBeNil)

			resourceResp := client.GET("/api/vega-backend/v1/resources?catalog_id=" + catalogID + "&offset=0&limit=100")
			So(resourceResp.StatusCode, ShouldEqual, http.StatusOK)
			entries, ok := resourceResp.Body["entries"].([]any)
			So(ok, ShouldBeTrue)
			So(len(entries), ShouldEqual, tableSize)

			for _, entry := range entries {
				resource, ok := entry.(map[string]any)
				So(ok, ShouldBeTrue)
				schema_definition, ok := resource["schema_definition"].([]any)
				So(ok, ShouldBeTrue)
				So(schema_definition, ShouldResemble, tableSchema)
			}
		})

		Convey("MD506: 验证发现的Resource的source_metadata", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			err = builder.RunDiscoverTask(client, catalogID)
			So(err, ShouldBeNil)

			resourceResp := client.GET("/api/vega-backend/v1/resources?catalog_id=" + catalogID + "&offset=0&limit=100")
			So(resourceResp.StatusCode, ShouldEqual, http.StatusOK)
			entries, ok := resourceResp.Body["entries"].([]any)
			So(ok, ShouldBeTrue)
			So(len(entries), ShouldEqual, tableSize)

			for _, entry := range entries {
				resource, ok := entry.(map[string]any)
				So(ok, ShouldBeTrue)
				source_metadata, ok := resource["source_metadata"].(map[string]any)
				So(ok, ShouldBeTrue)
				So(source_metadata["columns"], ShouldNotBeEmpty)
				So(source_metadata["indices"], ShouldNotBeEmpty)
				So(source_metadata["primary_keys"], ShouldNotBeEmpty)
				So(source_metadata["properties"], ShouldNotBeEmpty)
				So(source_metadata["table_type"], ShouldEqual, "table")
			}
		})

		Convey("MD507: Discover后重复执行", func() {
			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			err = builder.RunDiscoverTask(client, catalogID)
			So(err, ShouldBeNil)

			resourceResp1 := client.GET("/api/vega-backend/v1/resources?catalog_id=" + catalogID + "&offset=0&limit=100")
			entries1, ok := resourceResp1.Body["entries"].([]any)
			So(ok, ShouldBeTrue)
			So(len(entries1), ShouldEqual, tableSize)

			err = builder.RunDiscoverTask(client, catalogID)
			So(err, ShouldBeNil)

			resourceResp2 := client.GET("/api/vega-backend/v1/resources?catalog_id=" + catalogID + "&offset=0&limit=100")
			entries2, ok := resourceResp2.Body["entries"].([]any)
			So(ok, ShouldBeTrue)
			So(len(entries2), ShouldEqual, tableSize)
		})

		Convey("MD508: 实例级Catalog Discover（不指定database）", func() {
			payload := builder.BuildCreatePayloadWithoutDatabase()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			err = builder.RunDiscoverTask(client, catalogID)
			So(err, ShouldBeNil)
		})

		// ========== Discover 负向测试（MD521-MD528） ==========

		// Convey("MD521: Discover不存在的Catalog", func() {
		// 	discoverResp := client.POST("/api/vega-backend/v1/catalogs/non-existent-catalog-id/discover", nil)
		// 	So(discoverResp.StatusCode, ShouldEqual, http.StatusNotFound)
		// })

		// Convey("MD522: Discover连接失败的Catalog", func() {
		// 	payload := builder.BuildCreatePayloadWithWrongCredentials()
		// 	createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
		// 	if createResp.StatusCode == http.StatusCreated {
		// 		catalogID := createResp.Body["id"].(string)
		// 		err = builder.RunDiscoverTask(client, catalogID)
		// 		So(err, ShouldNotBeNil)
		// 	} else {
		// 		So(createResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		// 	}
		// })

		// Convey("MD523: Discover权限不足的数据库", func() {
		// 	payload := builder.BuildCreatePayload()
		// 	createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
		// 	So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

		// 	catalogID := createResp.Body["id"].(string)
		// 	err = builder.RunDiscoverTask(client, catalogID)
		// 	So(err, ShouldNotBeNil)
		// })

		// Convey("MD524: Discover空数据库", func() {
		// 	payload := builder.BuildCreatePayloadWithNonExistentDB()
		// 	createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
		// 	if createResp.StatusCode == http.StatusCreated {
		// 		catalogID := createResp.Body["id"].(string)
		// 		err = builder.RunDiscoverTask(client, catalogID)
		// 		So(err, ShouldNotBeNil)
		// 	} else {
		// 		So(createResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		// 	}
		// })

		// Convey("MD525: Discover时数据库连接超时", func() {
		// 	mariadbConfig := builder.GetConfig()
		// 	payload := map[string]any{
		// 		"name":           cataloghelpers.GenerateUniqueName("timeout-catalog"),
		// 		"connector_type": "mariadb",
		// 		"connector_config": map[string]any{
		// 			"host":      "192.0.2.1",
		// 			"port":      mariadbConfig.Port,
		// 			"databases": []string{mariadbConfig.Database},
		// 			"username":  mariadbConfig.Username,
		// 			"password":  builder.GetEncryptedPassword(),
		// 			"options": map[string]any{
		// 				"timeout": "1s",
		// 			},
		// 		},
		// 	}
		// 	createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
		// 	if createResp.StatusCode == http.StatusCreated {
		// 		catalogID := createResp.Body["id"].(string)
		// 		err = builder.RunDiscoverTask(client, catalogID)
		// 		So(err, ShouldNotBeNil)
		// 	}
		// })

		// Convey("MD526: Discover时数据库认证失败", func() {
		// 	payload := builder.BuildCreatePayloadWithWrongCredentials()
		// 	createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
		// 	if createResp.StatusCode == http.StatusCreated {
		// 		catalogID := createResp.Body["id"].(string)
		// 		err = builder.RunDiscoverTask(client, catalogID)
		// 		So(err, ShouldNotBeNil)
		// 	} else {
		// 		So(createResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		// 	}
		// })

		// Convey("MD527: Discover时数据库不存在", func() {
		// 	payload := builder.BuildCreatePayloadWithNonExistentDB()
		// 	createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
		// 	if createResp.StatusCode == http.StatusCreated {
		// 		catalogID := createResp.Body["id"].(string)
		// 		err = builder.RunDiscoverTask(client, catalogID)
		// 		So(err, ShouldNotBeNil)
		// 	} else {
		// 		So(createResp.StatusCode, ShouldEqual, http.StatusBadRequest)
		// 	}
		// })

		// ========== Discover 边界测试（MD531-MD536） ==========

		Convey("MD531: Discover表数量边界 - 少量表", func() {
			tableSize := 10
			recordSize := 0
			tableSuffix := fmt.Sprintf("%d", time.Now().Unix())
			_, _, err = builder.ResetTestDatabase(tableSuffix, tableSize, recordSize)
			So(err, ShouldBeNil)

			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			err = builder.RunDiscoverTask(client, catalogID)
			So(err, ShouldBeNil)

			resourceResp := client.GET("/api/vega-backend/v1/resources?catalog_id=" + catalogID + "&offset=0&limit=100")
			So(resourceResp.StatusCode, ShouldEqual, http.StatusOK)
			entries, ok := resourceResp.Body["entries"].([]any)
			So(ok, ShouldBeTrue)
			So(len(entries), ShouldEqual, tableSize)
		})

		Convey("MD532: Discover表数量边界 - 大量表（100+）", func() {
			tableSize := 110
			recordSize := 0
			tableSuffix := fmt.Sprintf("%d", time.Now().Unix())
			_, _, err = builder.ResetTestDatabase(tableSuffix, tableSize, recordSize)
			So(err, ShouldBeNil)

			payload := builder.BuildCreatePayload()
			createResp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(createResp.StatusCode, ShouldEqual, http.StatusCreated)

			catalogID := createResp.Body["id"].(string)
			err = builder.RunDiscoverTask(client, catalogID)
			So(err, ShouldBeNil)

			resourceResp := client.GET("/api/vega-backend/v1/resources?catalog_id=" + catalogID + "&offset=0&limit=200")
			So(resourceResp.StatusCode, ShouldEqual, http.StatusOK)
			entries, ok := resourceResp.Body["entries"].([]any)
			So(ok, ShouldBeTrue)
			So(len(entries), ShouldEqual, tableSize)
		})
	})
}

// TestMariaDBSpecificOptions MariaDB特有选项AT测试
// 编号规则：MD6xx
func TestMariaDBSpecificOptions(t *testing.T) {

	Convey("MariaDB特有选项AT测试 - 初始化", t, func() {

		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)
		So(config.TargetMariaDB.Host, ShouldNotBeEmpty)

		builder := NewMariaDBPayloadBuilder(config.TargetMariaDB)
		builder.SetTestConfig(config)

		tableSize := 1
		recordSize := 0
		tableSuffix := fmt.Sprintf("%d", time.Now().Unix())
		testTableName, records, err := builder.ResetTestDatabase(tableSuffix, tableSize, recordSize)
		So(err, ShouldBeNil)
		So(testTableName, ShouldNotBeEmpty)
		So(len(records), ShouldEqual, recordSize)

		client := testutil.NewHTTPClient(config.VegaBackend.BaseURL)
		err = client.CheckHealth()
		So(err, ShouldBeNil)

		cataloghelpers.CleanupCatalogs(client, t)

		// ========== MariaDB特有测试（MD601-MD606） ==========
		Convey("MD601: MariaDB charset选项测试（utf8mb4）", func() {
			options := map[string]any{
				"charset": "utf8mb4",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD602: MariaDB parseTime选项测试", func() {
			options := map[string]any{
				"parseTime": "true",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD603: MariaDB loc选项测试（时区）", func() {
			options := map[string]any{
				"loc": "Local",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD604: MariaDB timeout选项测试", func() {
			options := map[string]any{
				"timeout": "30s",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})

		Convey("MD605: MariaDB SSL连接测试", func() {
			options := map[string]any{
				"tls": "true",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldBeIn, []int{http.StatusCreated, http.StatusBadRequest})
		})

		Convey("MD606: MariaDB collation选项测试", func() {
			options := map[string]any{
				"collation": "utf8mb4_unicode_ci",
			}
			payload := builder.BuildCreatePayloadWithOptions(options)
			resp := client.POST("/api/vega-backend/v1/catalogs", payload)
			So(resp.StatusCode, ShouldEqual, http.StatusCreated)
		})
	})
}
