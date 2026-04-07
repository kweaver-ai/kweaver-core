// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// TestJSONToObject_Success 测试 JSONToObject 成功场景
func TestJSONToObject_Success(t *testing.T) {
	convey.Convey("TestJSONToObject_Success", t, func() {
		type TestStruct struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		jsonStr := `{"name": "test", "value": 123}`
		result := JSONToObject[TestStruct](jsonStr)
		convey.So(result.Name, convey.ShouldEqual, "test")
		convey.So(result.Value, convey.ShouldEqual, 123)
	})
}

// TestJSONToObject_EmptyString 测试 JSONToObject 空字符串
func TestJSONToObject_EmptyString(t *testing.T) {
	convey.Convey("TestJSONToObject_EmptyString", t, func() {
		type TestStruct struct {
			Name string `json:"name"`
		}

		result := JSONToObject[TestStruct]("")
		convey.So(result.Name, convey.ShouldEqual, "")
	})
}

// TestJSONToObject_InvalidJSON 测试 JSONToObject 无效 JSON
func TestJSONToObject_InvalidJSON(t *testing.T) {
	convey.Convey("TestJSONToObject_InvalidJSON", t, func() {
		type TestStruct struct {
			Name string `json:"name"`
		}

		result := JSONToObject[TestStruct]("invalid json")
		convey.So(result.Name, convey.ShouldEqual, "")
	})
}

// TestJSONToObjectWithError_Success 测试 JSONToObjectWithError 成功场景
func TestJSONToObjectWithError_Success(t *testing.T) {
	convey.Convey("TestJSONToObjectWithError_Success", t, func() {
		type TestStruct struct {
			Name string `json:"name"`
		}

		result, err := JSONToObjectWithError[TestStruct](`{"name": "test"}`)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.Name, convey.ShouldEqual, "test")
	})
}

// TestJSONToObjectWithError_EmptyString 测试 JSONToObjectWithError 空字符串
func TestJSONToObjectWithError_EmptyString(t *testing.T) {
	convey.Convey("TestJSONToObjectWithError_EmptyString", t, func() {
		type TestStruct struct {
			Name string `json:"name"`
		}

		result, err := JSONToObjectWithError[TestStruct]("")
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.Name, convey.ShouldEqual, "")
	})
}

// TestJSONToObjectWithError_InvalidJSON 测试 JSONToObjectWithError 无效 JSON
func TestJSONToObjectWithError_InvalidJSON(t *testing.T) {
	convey.Convey("TestJSONToObjectWithError_InvalidJSON", t, func() {
		type TestStruct struct {
			Name string `json:"name"`
		}

		_, err := JSONToObjectWithError[TestStruct]("invalid json")
		convey.So(err, convey.ShouldNotBeNil)
	})
}

// TestAnyToObject_Success 测试 AnyToObject 成功场景
func TestAnyToObject_Success(t *testing.T) {
	convey.Convey("TestAnyToObject_Success", t, func() {
		type TestStruct struct {
			Name  string `json:"name"`
			Value int    `json:"value"`
		}

		source := map[string]interface{}{
			"name":  "test",
			"value": 123,
		}

		var result TestStruct
		err := AnyToObject(source, &result)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result.Name, convey.ShouldEqual, "test")
		convey.So(result.Value, convey.ShouldEqual, 123)
	})
}

// TestAnyToObject_SliceToStruct 测试 AnyToObject 数组转换
func TestAnyToObject_SliceToStruct(t *testing.T) {
	convey.Convey("TestAnyToObject_SliceToStruct", t, func() {
		source := []map[string]interface{}{
			{"name": "item1"},
			{"name": "item2"},
		}

		type Item struct {
			Name string `json:"name"`
		}

		var result []Item
		err := AnyToObject(source, &result)
		convey.So(err, convey.ShouldBeNil)
		convey.So(len(result), convey.ShouldEqual, 2)
		convey.So(result[0].Name, convey.ShouldEqual, "item1")
	})
}
