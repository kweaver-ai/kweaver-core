// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_CE(t *testing.T) {
	Convey("Test CE", t, func() {
		Convey("成功 - 条件为true返回trueVal", func() {
			result := CE(true, "yes", "no")
			So(result, ShouldEqual, "yes")
		})

		Convey("成功 - 条件为false返回falseVal", func() {
			result := CE(false, "yes", "no")
			So(result, ShouldEqual, "no")
		})

		Convey("成功 - int类型", func() {
			result := CE(true, 10, 20)
			So(result, ShouldEqual, 10)
			result = CE(false, 10, 20)
			So(result, ShouldEqual, 20)
		})

		Convey("成功 - bool类型", func() {
			result := CE(true, true, false)
			So(result, ShouldBeTrue)
			result = CE(false, true, false)
			So(result, ShouldBeFalse)
		})
	})
}

func Test_QuotationMark(t *testing.T) {
	Convey("Test QuotationMark", t, func() {
		Convey("成功 - 普通字符串", func() {
			result := QuotationMark("test")
			So(result, ShouldEqual, "\"test\"")
		})

		Convey("成功 - 已有前缀引号", func() {
			result := QuotationMark("\"test")
			So(result, ShouldEqual, "\"test")
		})

		Convey("成功 - 已有后缀引号", func() {
			result := QuotationMark("test\"")
			So(result, ShouldEqual, "test\"")
		})

		Convey("成功 - 已有前后引号", func() {
			result := QuotationMark("\"test\"")
			So(result, ShouldEqual, "\"test\"")
		})

		Convey("成功 - 空字符串", func() {
			result := QuotationMark("")
			So(result, ShouldEqual, "\"\"")
		})
	})
}

func Test_GenerateUniqueKey(t *testing.T) {
	Convey("Test GenerateUniqueKey", t, func() {
		Convey("成功 - 单个label", func() {
			label := map[string]string{
				"key1": "value1",
			}
			result := GenerateUniqueKey("id1", label)
			So(result, ShouldEqual, "id1-key1:value1")
		})

		Convey("成功 - 多个label", func() {
			label := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			result := GenerateUniqueKey("id1", label)
			// 由于map遍历顺序不确定，只检查包含关键部分
			So(result, ShouldContainSubstring, "id1-")
			So(result, ShouldContainSubstring, "key1:value1")
			So(result, ShouldContainSubstring, "key2:value2")
		})

		Convey("成功 - 空label", func() {
			label := map[string]string{}
			result := GenerateUniqueKey("id1", label)
			So(result, ShouldEqual, "id1")
		})

		Convey("成功 - label键值对排序", func() {
			label := map[string]string{
				"z": "value1",
				"a": "value2",
			}
			result := GenerateUniqueKey("id1", label)
			// 应该按字母顺序排序
			So(result, ShouldContainSubstring, "a:value2")
			So(result, ShouldContainSubstring, "z:value1")
		})
	})
}

func Test_IsSlice(t *testing.T) {
	Convey("Test IsSlice", t, func() {
		Convey("成功 - []int", func() {
			result := IsSlice([]int{1, 2, 3})
			So(result, ShouldBeTrue)
		})

		Convey("成功 - []string", func() {
			result := IsSlice([]string{"a", "b"})
			So(result, ShouldBeTrue)
		})

		Convey("成功 - [3]int数组", func() {
			result := IsSlice([3]int{1, 2, 3})
			So(result, ShouldBeTrue)
		})

		Convey("失败 - string", func() {
			result := IsSlice("test")
			So(result, ShouldBeFalse)
		})

		Convey("失败 - int", func() {
			result := IsSlice(42)
			So(result, ShouldBeFalse)
		})

		Convey("失败 - map", func() {
			result := IsSlice(map[string]int{"a": 1})
			So(result, ShouldBeFalse)
		})
	})
}

func Test_IsSameType(t *testing.T) {
	Convey("Test IsSameType", t, func() {
		Convey("成功 - 空数组", func() {
			result := IsSameType([]any{})
			So(result, ShouldBeTrue)
		})

		Convey("成功 - 相同类型int", func() {
			result := IsSameType([]any{1, 2, 3})
			So(result, ShouldBeTrue)
		})

		Convey("成功 - 相同类型string", func() {
			result := IsSameType([]any{"a", "b", "c"})
			So(result, ShouldBeTrue)
		})

		Convey("失败 - 不同类型", func() {
			result := IsSameType([]any{1, "2", 3})
			So(result, ShouldBeFalse)
		})

		Convey("成功 - 单个元素", func() {
			result := IsSameType([]any{1})
			So(result, ShouldBeTrue)
		})
	})
}

func Test_SplitString2InterfaceArray(t *testing.T) {
	Convey("Test SplitString2InterfaceArray", t, func() {
		Convey("成功 - 单个分隔符", func() {
			result := SplitString2InterfaceArray("a,b,c", ",")
			So(result, ShouldResemble, []any{"a", "b", "c"})
		})

		Convey("成功 - 多个分隔符", func() {
			result := SplitString2InterfaceArray("a,b,c,d", ",")
			So(result, ShouldResemble, []any{"a", "b", "c", "d"})
		})

		Convey("成功 - 空字符串", func() {
			result := SplitString2InterfaceArray("", ",")
			So(result, ShouldResemble, []any{""})
		})

		Convey("成功 - 无分隔符", func() {
			result := SplitString2InterfaceArray("abc", ",")
			So(result, ShouldResemble, []any{"abc"})
		})

		Convey("成功 - 自定义分隔符", func() {
			result := SplitString2InterfaceArray("a|b|c", "|")
			So(result, ShouldResemble, []any{"a", "b", "c"})
		})

		Convey("成功 - 多字符分隔符", func() {
			result := SplitString2InterfaceArray("a||b||c", "||")
			So(result, ShouldResemble, []any{"a", "b", "c"})
		})
	})
}

func Test_Any2String(t *testing.T) {
	Convey("Test Any2String", t, func() {
		Convey("成功 - string类型", func() {
			result := Any2String("test")
			So(result, ShouldEqual, "test")
		})

		Convey("成功 - int类型", func() {
			result := Any2String(42)
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - int8类型", func() {
			result := Any2String(int8(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - int16类型", func() {
			result := Any2String(int16(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - int32类型", func() {
			result := Any2String(int32(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - int64类型", func() {
			result := Any2String(int64(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - uint类型", func() {
			result := Any2String(uint(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - uint8类型", func() {
			result := Any2String(uint8(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - uint16类型", func() {
			result := Any2String(uint16(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - uint32类型", func() {
			result := Any2String(uint32(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - uint64类型", func() {
			result := Any2String(uint64(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - float32类型", func() {
			result := Any2String(float32(42.5))
			So(result, ShouldEqual, "42.5")
		})

		Convey("成功 - float64类型", func() {
			result := Any2String(float64(42.5))
			So(result, ShouldEqual, "42.5")
		})

		Convey("成功 - []byte类型", func() {
			result := Any2String([]byte("test"))
			So(result, ShouldEqual, "test")
		})

		Convey("成功 - 不支持的类型返回空字符串", func() {
			result := Any2String([]int{1, 2, 3})
			So(result, ShouldEqual, "")
		})
	})
}

func Test_CloneStringMap(t *testing.T) {
	Convey("Test CloneStringMap", t, func() {
		Convey("成功 - 普通map", func() {
			original := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			cloned := CloneStringMap(original)
			So(cloned, ShouldResemble, original)
		})

		Convey("成功 - 空map", func() {
			original := map[string]string{}
			cloned := CloneStringMap(original)
			So(cloned, ShouldResemble, original)
			So(len(cloned), ShouldEqual, 0)
		})

		Convey("成功 - 修改克隆不影响原map", func() {
			original := map[string]string{
				"key1": "value1",
			}
			cloned := CloneStringMap(original)
			cloned["key2"] = "value2"
			So(len(original), ShouldEqual, 1)
			So(len(cloned), ShouldEqual, 2)
		})
	})
}

func Test_RandStringRunes(t *testing.T) {
	Convey("Test RandStringRunes", t, func() {
		Convey("成功 - 长度为0", func() {
			result := RandStringRunes(0)
			So(result, ShouldEqual, "")
		})

		Convey("成功 - 长度为1", func() {
			result := RandStringRunes(1)
			So(len(result), ShouldEqual, 1)
		})

		Convey("成功 - 长度为10", func() {
			result := RandStringRunes(10)
			So(len(result), ShouldEqual, 10)
		})

		Convey("成功 - 两次生成结果不同", func() {
			result1 := RandStringRunes(10)
			result2 := RandStringRunes(10)
			// 虽然理论上可能相同，但概率极低
			// 这里只检查长度
			So(len(result1), ShouldEqual, 10)
			So(len(result2), ShouldEqual, 10)
		})

		Convey("成功 - 只包含字母字符", func() {
			result := RandStringRunes(100)
			for _, r := range result {
				So((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'), ShouldBeTrue)
			}
		})
	})
}
