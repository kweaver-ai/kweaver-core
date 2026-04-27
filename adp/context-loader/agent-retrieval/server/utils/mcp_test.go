// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// TestGenerateMCPKey 测试 GenerateMCPKey 函数
func TestGenerateMCPKey(t *testing.T) {
	convey.Convey("TestGenerateMCPKey", t, func() {
		convey.Convey("正常参数", func() {
			result := GenerateMCPKey("mcp-001", 1)
			convey.So(result, convey.ShouldEqual, "mcp-001-1")
		})

		convey.Convey("版本号为 0", func() {
			result := GenerateMCPKey("mcp-001", 0)
			convey.So(result, convey.ShouldEqual, "mcp-001-0")
		})

		convey.Convey("空 mcpID", func() {
			result := GenerateMCPKey("", 1)
			convey.So(result, convey.ShouldEqual, "-1")
		})
	})
}

// TestGenerateMCPServerVersion 测试 GenerateMCPServerVersion 函数
func TestGenerateMCPServerVersion(t *testing.T) {
	convey.Convey("TestGenerateMCPServerVersion", t, func() {
		convey.Convey("版本号 1", func() {
			result := GenerateMCPServerVersion(1)
			convey.So(result, convey.ShouldEqual, "1.0.0")
		})

		convey.Convey("版本号 10", func() {
			result := GenerateMCPServerVersion(10)
			convey.So(result, convey.ShouldEqual, "10.0.0")
		})

		convey.Convey("版本号 0", func() {
			result := GenerateMCPServerVersion(0)
			convey.So(result, convey.ShouldEqual, "0.0.0")
		})
	})
}
