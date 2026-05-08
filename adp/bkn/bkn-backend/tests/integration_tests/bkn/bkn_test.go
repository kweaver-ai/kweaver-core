// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package bkn

import (
	"bytes"
	"net/http"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"bkn-backend-tests/integration_tests/bkn/helpers"
	"bkn-backend-tests/integration_tests/setup"
	"bkn-backend-tests/testutil"
)

// TestBKNImportExport BKN导入导出集成测试
// 测试编号前缀: BKN1xx (Import/Export)
func TestBKNImportExport(t *testing.T) {

	Convey("BKN导入导出集成测试 - 初始化", t, func() {

		// 加载测试配置
		config, err := setup.LoadTestConfig()
		So(err, ShouldBeNil)
		So(config, ShouldNotBeNil)

		// 创建HTTP客户端
		client := testutil.NewHTTPClient(config.BKNBackend.BaseURL)

		// 验证服务可用性
		err = client.CheckHealth()
		So(err, ShouldBeNil)
		t.Logf("✓ 集成测试环境就绪，BKN Backend: %s", config.BKNBackend.BaseURL)

		// 清理现有测试知识网络
		helpers.CleanupKNs(client, t)

		// ========== BKN 导入测试（BKN101-BKN103） ==========

		Convey("BKN101: 导入BKN - 使用 k8s-network 示例", func() {

			// 从 examples 目录构建 tar 包
			tarData, err := helpers.BuildTarFromExamplesDir("k8s-network")
			So(err, ShouldBeNil)
			So(tarData, ShouldNotBeNil)
			So(len(tarData), ShouldBeGreaterThan, 0)

			// 上传BKN文件
			resp := client.POSTMultipart(
				"/api/bkn-backend/v1/bkns",
				"file",
				tarData,
				"k8s-network.tar",
				nil,
			)

			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(resp.Body, ShouldNotBeNil)
			So(resp.Body["kn_id"], ShouldEqual, "k8s-network")
		})

		Convey("BKN102: 导入BKN后验证对象类型、关系类型、行动类型已创建", func() {

			knID := "k8s-network"
			// 导入BKN（k8s-network 包含 object_types、relation_types、action_types）
			tarData, _ := helpers.BuildTarFromExamplesDir("k8s-network")
			resp := client.POSTMultipart(
				"/api/bkn-backend/v1/bkns",
				"file",
				tarData,
				"k8s-network.tar",
				nil,
			)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			// 验证对象类型已创建
			otEntries := helpers.VerifyObjectTypesExist(client, knID, t)
			So(len(otEntries), ShouldBeGreaterThan, 0)

			// 验证关系类型已创建
			rtEntries := helpers.VerifyRelationTypesExist(client, knID, t)
			So(len(rtEntries), ShouldBeGreaterThan, 0)

			// 验证行动类型已创建
			atEntries := helpers.VerifyActionTypesExist(client, knID, t)
			So(len(atEntries), ShouldBeGreaterThan, 0)

			// 验证概念分组已创建
			cgEntries := helpers.VerifyConceptGroupsExist(client, knID, t)
			So(len(cgEntries), ShouldBeGreaterThan, 0)
		})

		Convey("BKN103: 导入含 metrics 的 k8s-network 并校验指标条数", func() {
			knID := "k8s-network"
			tarData, err := helpers.BuildTarFromExamplesDir("k8s-network")
			So(err, ShouldBeNil)
			resp := client.POSTMultipart(
				"/api/bkn-backend/v1/bkns",
				"file",
				tarData,
				"k8s-network.tar",
				nil,
			)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			n := helpers.VerifyMetricsCountAtLeast(client, knID, t, 5)
			So(n, ShouldBeGreaterThanOrEqualTo, 5)
		})

		// ========== BKN 导出测试（BKN121-BKN122） ==========

		Convey("BKN121: 导出BKN - 基本场景", func() {
			knID := "k8s-network"

			// 先导入一些数据
			tarData, _ := helpers.BuildTarFromExamplesDir("k8s-network")
			client.POSTMultipart(
				"/api/bkn-backend/v1/bkns",
				"file",
				tarData,
				"k8s-network.tar",
				nil,
			)

			// 导出BKN
			resp := client.GET("/api/bkn-backend/v1/bkns/" + knID)

			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(resp.RawBody, ShouldNotBeNil)
			So(len(resp.RawBody), ShouldBeGreaterThan, 0)
			So(helpers.IsValidTar(resp.RawBody), ShouldBeTrue)
		})

		Convey("BKN122: 导出BKN - 验证Content-Disposition包含kn_id", func() {
			knID := "k8s-network"

			// 先导入数据
			tarData, _ := helpers.BuildTarFromExamplesDir("k8s-network")
			client.POSTMultipart(
				"/api/bkn-backend/v1/bkns",
				"file",
				tarData,
				"k8s-network.tar",
				nil,
			)

			resp := client.GET("/api/bkn-backend/v1/bkns/" + knID)

			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			// 验证 Content-Disposition 头包含文件下载信息
			contentDisposition := resp.Headers.Get("Content-Disposition")
			So(contentDisposition, ShouldNotBeEmpty)
			So(strings.Contains(contentDisposition, knID), ShouldBeTrue)
		})

		Convey("BKN124: 导出 tar 含 metrics 条目（内容与示例一致）", func() {
			knID := "k8s-network"
			tarData, _ := helpers.BuildTarFromExamplesDir("k8s-network")
			client.POSTMultipart(
				"/api/bkn-backend/v1/bkns",
				"file",
				tarData,
				"k8s-network.tar",
				nil,
			)

			resp := client.GET("/api/bkn-backend/v1/bkns/" + knID)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)
			So(bytes.Contains(resp.RawBody, []byte("metrics/")), ShouldBeTrue)
			So(bytes.Contains(resp.RawBody, []byte("pod_running_count")), ShouldBeTrue)
		})

		// ========== 负向测试（BKN201-BKN220） ==========

		Convey("BKN201: 导入无效文件格式", func() {

			// 上传非tar文件
			invalidData := []byte("this is not a tar file")
			resp := client.POSTMultipart(
				"/api/bkn-backend/v1/bkns",
				"file",
				invalidData,
				"invalid.txt",
				nil,
			)

			// 应该返回错误
			So(resp.StatusCode, ShouldBeGreaterThanOrEqualTo, 400)
		})

		Convey("BKN202: 导出不存在的知识网络", func() {
			// 尝试导出不存在的KN
			resp := client.GET("/api/bkn-backend/v1/bkns/non-existent-kn-id")

			// 应该返回错误
			So(resp.StatusCode, ShouldBeGreaterThanOrEqualTo, 400)
		})

		Convey("BKN203: 导入空文件", func() {

			// 上传空文件
			resp := client.POSTMultipart(
				"/api/bkn-backend/v1/bkns",
				"file",
				[]byte{},
				"empty.tar",
				nil,
			)

			// 应该返回错误
			So(resp.StatusCode, ShouldBeGreaterThanOrEqualTo, 400)
		})

		Convey("BKN204: 导入缺少network.bkn的tar包", func() {

			// 构建缺少 network.bkn 的 tar 包
			tarData, err := helpers.BuildTarWithoutNetworkBKN()
			So(err, ShouldBeNil)

			resp := client.POSTMultipart(
				"/api/bkn-backend/v1/bkns",
				"file",
				tarData,
				"incomplete.tar",
				nil,
			)

			// 应该返回错误
			So(resp.StatusCode, ShouldBeGreaterThanOrEqualTo, 400)
		})

		Convey("BKN206: strict_mode=true 时 data_view 对象类上的指标严格校验失败", func() {
			tarData, err := helpers.BuildTarFromExamplesDir("k8s-network")
			So(err, ShouldBeNil)
			resp := client.POSTMultipart(
				"/api/bkn-backend/v1/bkns?strict_mode=true",
				"file",
				tarData,
				"k8s-network.tar",
				nil,
			)
			So(resp.StatusCode, ShouldBeGreaterThanOrEqualTo, 400)
		})

		// ========== 复杂数据测试（BKN221） ==========

		Convey("BKN221: 导出包含复杂结构的BKN", func() {
			knID := "k8s-network"

			// 先导入复杂结构（k8s-network 包含对象、关系、行动等）
			tarData, _ := helpers.BuildTarFromExamplesDir("k8s-network")
			client.POSTMultipart(
				"/api/bkn-backend/v1/bkns",
				"file",
				tarData,
				"k8s-network.tar",
				nil,
			)

			// 导出
			resp := client.GET("/api/bkn-backend/v1/bkns/" + knID)
			So(resp.StatusCode, ShouldEqual, http.StatusOK)

			// 验证导出内容包含所有类型
			So(bytes.Contains(resp.RawBody, []byte("object_types")), ShouldBeTrue)
			So(bytes.Contains(resp.RawBody, []byte("relation_types")), ShouldBeTrue)
			So(bytes.Contains(resp.RawBody, []byte("action_types")), ShouldBeTrue)
		})
	})
}
