// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package opensearch

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend-tests/at/resource/helpers"
)

// TestOpenSearchResourceCommon OpenSearch Resource通用AT测试
// 使用resource/helpers包中的通用测试用例
func TestOpenSearchResourceCommon(t *testing.T) {
	Convey("OpenSearch Resource通用AT测试 - 初始化", t, func() {
		// 创建测试套件
		suite, err := helpers.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)
		So(suite, ShouldNotBeNil)

		// 初始化测试环境
		err = suite.Setup()
		So(err, ShouldBeNil)

		// 测试结束后清理
		defer suite.Cleanup()

		// ========== 创建测试（RM1xx） ==========
		Convey("创建测试（RM1xx）", func() {
			helpers.RunCommonCreateTests(suite)
		})

		// ========== 负向测试（RM1xx 121-140） ==========
		Convey("负向测试（RM1xx 121-140）", func() {
			helpers.RunCommonNegativeTests(suite)
		})

		// ========== 边界测试（RM1xx 141-160） ==========
		Convey("边界测试（RM1xx 141-160）", func() {
			helpers.RunCommonBoundaryTests(suite)
		})

		// ========== 安全测试（RM1xx 161-170） ==========
		Convey("安全测试（RM1xx 161-170）", func() {
			helpers.RunCommonSecurityTests(suite)
		})
	})
}

// TestOpenSearchResourceRead OpenSearch Resource读取AT测试
func TestOpenSearchResourceRead(t *testing.T) {
	Convey("OpenSearch Resource读取AT测试 - 初始化", t, func() {
		suite, err := helpers.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)
		defer suite.Cleanup()

		// ========== 读取测试（RM2xx） ==========
		Convey("读取测试（RM2xx）", func() {
			helpers.RunCommonReadTests(suite)
		})
	})
}

// TestOpenSearchResourceUpdate OpenSearch Resource更新AT测试
func TestOpenSearchResourceUpdate(t *testing.T) {
	Convey("OpenSearch Resource更新AT测试 - 初始化", t, func() {
		suite, err := helpers.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)
		defer suite.Cleanup()

		// ========== 更新测试（RM3xx） ==========
		Convey("更新测试（RM3xx）", func() {
			helpers.RunCommonUpdateTests(suite)
		})
	})
}

// TestOpenSearchResourceDelete OpenSearch Resource删除AT测试
func TestOpenSearchResourceDelete(t *testing.T) {
	Convey("OpenSearch Resource删除AT测试 - 初始化", t, func() {
		suite, err := helpers.NewTestSuite(t, "opensearch")
		So(err, ShouldBeNil)

		err = suite.Setup()
		So(err, ShouldBeNil)
		defer suite.Cleanup()

		// ========== 删除测试（RM4xx） ==========
		Convey("删除测试（RM4xx）", func() {
			helpers.RunCommonDeleteTests(suite)
		})

		// ========== 名称唯一性测试（RM5xx） ==========
		Convey("名称唯一性测试（RM5xx）", func() {
			helpers.RunCommonNameUniquenessTests(suite)
		})
	})
}
