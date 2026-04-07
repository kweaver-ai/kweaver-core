// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Util_GenerateUniqueKey(t *testing.T) {
	Convey("Test GenerateUniqueKey", t, func() {
		Convey("Success with single label\n", func() {
			id := "test-id"
			label := map[string]string{
				"key1": "value1",
			}
			result := GenerateUniqueKey(id, label)
			So(result, ShouldContainSubstring, id)
			So(result, ShouldContainSubstring, "key1:value1")
		})

		Convey("Success with multiple labels\n", func() {
			id := "test-id"
			label := map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			}
			result := GenerateUniqueKey(id, label)
			So(result, ShouldContainSubstring, id)
			So(result, ShouldContainSubstring, "key1:value1")
			So(result, ShouldContainSubstring, "key2:value2")
			So(result, ShouldContainSubstring, "key3:value3")
		})

		Convey("Success with empty label\n", func() {
			id := "test-id"
			label := map[string]string{}
			result := GenerateUniqueKey(id, label)
			So(result, ShouldEqual, id)
		})
	})
}

func Test_Util_IsSlice(t *testing.T) {
	Convey("Test IsSlice", t, func() {
		Convey("Success with slice\n", func() {
			result := IsSlice([]string{"a", "b"})
			So(result, ShouldBeTrue)
		})

		Convey("Success with array\n", func() {
			result := IsSlice([3]int{1, 2, 3})
			So(result, ShouldBeTrue)
		})

		Convey("Failed with string\n", func() {
			result := IsSlice("test")
			So(result, ShouldBeFalse)
		})

		Convey("Failed with int\n", func() {
			result := IsSlice(123)
			So(result, ShouldBeFalse)
		})

		Convey("Failed with map\n", func() {
			result := IsSlice(map[string]string{"key": "value"})
			So(result, ShouldBeFalse)
		})
	})
}

func Test_Util_IsSameType(t *testing.T) {
	Convey("Test IsSameType", t, func() {
		Convey("Success with same type\n", func() {
			arr := []any{1, 2, 3}
			result := IsSameType(arr)
			So(result, ShouldBeTrue)
		})

		Convey("Success with empty array\n", func() {
			arr := []any{}
			result := IsSameType(arr)
			So(result, ShouldBeTrue)
		})

		Convey("Failed with different types\n", func() {
			arr := []any{1, "test", 3.14}
			result := IsSameType(arr)
			So(result, ShouldBeFalse)
		})

		Convey("Success with strings\n", func() {
			arr := []any{"a", "b", "c"}
			result := IsSameType(arr)
			So(result, ShouldBeTrue)
		})
	})
}

func Test_Util_SplitString2InterfaceArray(t *testing.T) {
	Convey("Test SplitString2InterfaceArray", t, func() {
		Convey("Success with comma separator\n", func() {
			result := SplitString2InterfaceArray("a,b,c", ",")
			So(len(result), ShouldEqual, 3)
			So(result[0], ShouldEqual, "a")
			So(result[1], ShouldEqual, "b")
			So(result[2], ShouldEqual, "c")
		})

		Convey("Success with custom separator\n", func() {
			result := SplitString2InterfaceArray("a|b|c", "|")
			So(len(result), ShouldEqual, 3)
			So(result[0], ShouldEqual, "a")
			So(result[1], ShouldEqual, "b")
			So(result[2], ShouldEqual, "c")
		})

		Convey("Success with empty string\n", func() {
			result := SplitString2InterfaceArray("", ",")
			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, "")
		})

		Convey("Success with no separator\n", func() {
			result := SplitString2InterfaceArray("abc", ",")
			So(len(result), ShouldEqual, 1)
			So(result[0], ShouldEqual, "abc")
		})

		Convey("Success with multiple separators\n", func() {
			result := SplitString2InterfaceArray("a,,b", ",")
			So(len(result), ShouldEqual, 3)
			So(result[0], ShouldEqual, "a")
			So(result[1], ShouldEqual, "")
			So(result[2], ShouldEqual, "b")
		})
	})
}

func Test_Util_Any2String(t *testing.T) {
	Convey("Test Any2String", t, func() {
		Convey("Success with string\n", func() {
			result := Any2String("test")
			So(result, ShouldEqual, "test")
		})

		Convey("Success with int\n", func() {
			result := Any2String(123)
			So(result, ShouldEqual, "123")
		})

		Convey("Success with int8\n", func() {
			result := Any2String(int8(123))
			So(result, ShouldEqual, "123")
		})

		Convey("Success with int16\n", func() {
			result := Any2String(int16(123))
			So(result, ShouldEqual, "123")
		})

		Convey("Success with int32\n", func() {
			result := Any2String(int32(123))
			So(result, ShouldEqual, "123")
		})

		Convey("Success with int64\n", func() {
			result := Any2String(int64(123))
			So(result, ShouldEqual, "123")
		})

		Convey("Success with uint\n", func() {
			result := Any2String(uint(123))
			So(result, ShouldEqual, "123")
		})

		Convey("Success with uint8\n", func() {
			result := Any2String(uint8(123))
			So(result, ShouldEqual, "123")
		})

		Convey("Success with uint16\n", func() {
			result := Any2String(uint16(123))
			So(result, ShouldEqual, "123")
		})

		Convey("Success with uint32\n", func() {
			result := Any2String(uint32(123))
			So(result, ShouldEqual, "123")
		})

		Convey("Success with uint64\n", func() {
			result := Any2String(uint64(123))
			So(result, ShouldEqual, "123")
		})

		Convey("Success with float32\n", func() {
			result := Any2String(float32(123.45))
			// float32 has precision issues, so check it contains the expected value
			So(result, ShouldContainSubstring, "123.4")
		})

		Convey("Success with float64\n", func() {
			result := Any2String(float64(123.45))
			So(result, ShouldEqual, "123.45")
		})

		Convey("Success with []byte\n", func() {
			result := Any2String([]byte("test"))
			So(result, ShouldEqual, "test")
		})

		Convey("Success with unsupported type\n", func() {
			result := Any2String(map[string]string{"key": "value"})
			So(result, ShouldEqual, "")
		})
	})
}

func Test_Util_CloneStringMap(t *testing.T) {
	Convey("Test CloneStringMap", t, func() {
		Convey("Success with map\n", func() {
			original := map[string]string{
				"key1": "value1",
				"key2": "value2",
			}
			result := CloneStringMap(original)
			So(len(result), ShouldEqual, 2)
			So(result["key1"], ShouldEqual, "value1")
			So(result["key2"], ShouldEqual, "value2")
			// Modify original to ensure it's a clone
			original["key3"] = "value3"
			So(len(result), ShouldEqual, 2)
		})

		Convey("Success with empty map\n", func() {
			original := map[string]string{}
			result := CloneStringMap(original)
			So(len(result), ShouldEqual, 0)
			So(result, ShouldNotBeNil)
		})
	})
}

func Test_Util_RandStringRunes(t *testing.T) {
	Convey("Test RandStringRunes", t, func() {
		Convey("Success with positive length\n", func() {
			result := RandStringRunes(10)
			So(len(result), ShouldEqual, 10)
		})

		Convey("Success with zero length\n", func() {
			result := RandStringRunes(0)
			So(len(result), ShouldEqual, 0)
		})

		Convey("Success with different lengths\n", func() {
			result1 := RandStringRunes(5)
			result2 := RandStringRunes(5)
			So(len(result1), ShouldEqual, 5)
			So(len(result2), ShouldEqual, 5)
			// Results should be different (random)
			So(result1, ShouldNotEqual, result2)
		})
	})
}
