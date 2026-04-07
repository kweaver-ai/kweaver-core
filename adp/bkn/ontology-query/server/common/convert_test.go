// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package common

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Convert_StringToStringSlice(t *testing.T) {
	Convey("Test StringToStringSlice", t, func() {
		Convey("success 3->3", func() {
			result := StringToStringSlice("1, 2, 3")
			So(result, ShouldResemble, []string{"1", "2", "3"})
		})
		Convey("success 3->2", func() {
			result := StringToStringSlice("1, 2, ")
			So(result, ShouldResemble, []string{"1", "2"})
		})
		Convey("empty", func() {
			result := StringToStringSlice("")
			So(result, ShouldResemble, []string{})
		})
	})
}

func Test_Convert_BytesToGiB(t *testing.T) {
	Convey("Test BytesToGiB", t, func() {
		Convey("BytesToGiB", func() {
			actual := BytesToGiB(1024)
			So(actual, ShouldEqual, 0)
		})
		Convey("BytesToGiB2", func() {
			actual := BytesToGiB(1024 * 1024 * 1024)
			So(actual, ShouldEqual, 1)
		})
		Convey("BytesToGiB3", func() {
			actual := BytesToGiB(1024 * 1024 * 1024 * 10)
			So(actual, ShouldEqual, 10)
		})
	})
}

func Test_Convert_GiBToBytes(t *testing.T) {
	Convey("Test GiBToBytes", t, func() {
		Convey("GiBToBytes", func() {
			actual := GiBToBytes(1)
			So(actual, ShouldEqual, 1024*1024*1024)
		})
		Convey("GiBToBytes2", func() {
			actual := GiBToBytes(10)
			So(actual, ShouldEqual, 1024*1024*1024*10)
		})
		Convey("GiBToBytes3", func() {
			actual := GiBToBytes(0)
			So(actual, ShouldEqual, 0)
		})
	})
}

func Test_AnyToInt64(t *testing.T) {
	Convey("Test AnyToInt64", t, func() {
		Convey("成功 - int类型", func() {
			result, err := AnyToInt64(42)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("成功 - int8类型", func() {
			result, err := AnyToInt64(int8(42))
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("成功 - int16类型", func() {
			result, err := AnyToInt64(int16(42))
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("成功 - int32类型", func() {
			result, err := AnyToInt64(int32(42))
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("成功 - int64类型", func() {
			result, err := AnyToInt64(int64(42))
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("成功 - uint类型", func() {
			result, err := AnyToInt64(uint(42))
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("成功 - uint8类型", func() {
			result, err := AnyToInt64(uint8(42))
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("成功 - uint16类型", func() {
			result, err := AnyToInt64(uint16(42))
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("成功 - uint32类型", func() {
			result, err := AnyToInt64(uint32(42))
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("成功 - uint64类型", func() {
			result, err := AnyToInt64(uint64(42))
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("成功 - float32类型", func() {
			result, err := AnyToInt64(float32(42.5))
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("成功 - float64类型", func() {
			result, err := AnyToInt64(float64(42.5))
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("成功 - string类型", func() {
			result, err := AnyToInt64("42")
			So(err, ShouldBeNil)
			So(result, ShouldEqual, 42)
		})

		Convey("失败 - string类型无效", func() {
			_, err := AnyToInt64("invalid")
			So(err, ShouldNotBeNil)
		})

		Convey("失败 - 不支持的类型", func() {
			_, err := AnyToInt64([]int{1, 2, 3})
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_AnyToBool(t *testing.T) {
	Convey("Test AnyToBool", t, func() {
		Convey("成功 - bool类型true", func() {
			result, err := AnyToBool(true)
			So(err, ShouldBeNil)
			So(result, ShouldBeTrue)
		})

		Convey("成功 - bool类型false", func() {
			result, err := AnyToBool(false)
			So(err, ShouldBeNil)
			So(result, ShouldBeFalse)
		})

		Convey("成功 - string类型true", func() {
			result, err := AnyToBool("true")
			So(err, ShouldBeNil)
			So(result, ShouldBeTrue)
		})

		Convey("成功 - string类型false", func() {
			result, err := AnyToBool("false")
			So(err, ShouldBeNil)
			So(result, ShouldBeFalse)
		})

		Convey("成功 - string类型1", func() {
			result, err := AnyToBool("1")
			So(err, ShouldBeNil)
			So(result, ShouldBeTrue)
		})

		Convey("成功 - string类型0", func() {
			result, err := AnyToBool("0")
			So(err, ShouldBeNil)
			So(result, ShouldBeFalse)
		})

		Convey("成功 - string类型TRUE", func() {
			result, err := AnyToBool("TRUE")
			So(err, ShouldBeNil)
			So(result, ShouldBeTrue)
		})

		Convey("成功 - string类型FALSE", func() {
			result, err := AnyToBool("FALSE")
			So(err, ShouldBeNil)
			So(result, ShouldBeFalse)
		})

		Convey("失败 - string类型无效", func() {
			_, err := AnyToBool("invalid")
			So(err, ShouldNotBeNil)
		})

		Convey("失败 - 不支持的类型", func() {
			_, err := AnyToBool(42)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_AnyToString(t *testing.T) {
	Convey("Test AnyToString", t, func() {
		Convey("成功 - string类型", func() {
			result := AnyToString("test")
			So(result, ShouldEqual, "test")
		})

		Convey("成功 - int类型", func() {
			result := AnyToString(42)
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - int8类型", func() {
			result := AnyToString(int8(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - int16类型", func() {
			result := AnyToString(int16(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - int32类型", func() {
			result := AnyToString(int32(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - int64类型", func() {
			result := AnyToString(int64(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - uint类型", func() {
			result := AnyToString(uint(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - uint8类型", func() {
			result := AnyToString(uint8(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - uint16类型", func() {
			result := AnyToString(uint16(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - uint32类型", func() {
			result := AnyToString(uint32(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - uint64类型", func() {
			result := AnyToString(uint64(42))
			So(result, ShouldEqual, "42")
		})

		Convey("成功 - float32类型", func() {
			result := AnyToString(float32(42.5))
			So(result, ShouldEqual, "42.500000")
		})

		Convey("成功 - float64类型", func() {
			result := AnyToString(float64(42.5))
			So(result, ShouldEqual, "42.500000")
		})

		Convey("成功 - bool类型true", func() {
			result := AnyToString(true)
			So(result, ShouldEqual, "true")
		})

		Convey("成功 - bool类型false", func() {
			result := AnyToString(false)
			So(result, ShouldEqual, "false")
		})

		Convey("成功 - []byte类型", func() {
			result := AnyToString([]byte("test"))
			So(result, ShouldEqual, "test")
		})

		Convey("成功 - nil类型", func() {
			result := AnyToString(nil)
			So(result, ShouldEqual, "")
		})

		Convey("成功 - 其他类型", func() {
			result := AnyToString([]int{1, 2, 3})
			So(result, ShouldNotBeEmpty)
		})
	})
}

func Test_ReplaceLikeWildcards(t *testing.T) {
	Convey("Test ReplaceLikeWildcards", t, func() {
		Convey("成功 - 空字符串", func() {
			result := ReplaceLikeWildcards("")
			So(result, ShouldEqual, "")
		})

		Convey("成功 - 普通字符串", func() {
			result := ReplaceLikeWildcards("test")
			So(result, ShouldEqual, "test")
		})

		Convey("成功 - %通配符", func() {
			result := ReplaceLikeWildcards("test%")
			So(result, ShouldEqual, "test.*")
		})

		Convey("成功 - _通配符", func() {
			result := ReplaceLikeWildcards("test_")
			So(result, ShouldEqual, "test.")
		})

		Convey("成功 - 多个%通配符", func() {
			result := ReplaceLikeWildcards("%test%")
			So(result, ShouldEqual, ".*test.*")
		})

		Convey("成功 - 多个_通配符", func() {
			result := ReplaceLikeWildcards("test__")
			So(result, ShouldEqual, "test..")
		})

		Convey("成功 - 转义%", func() {
			result := ReplaceLikeWildcards("test\\%")
			So(result, ShouldEqual, "test%")
		})

		Convey("成功 - 转义_", func() {
			result := ReplaceLikeWildcards("test\\_")
			So(result, ShouldEqual, "test_")
		})

		Convey("成功 - 转义\\", func() {
			result := ReplaceLikeWildcards("test\\\\")
			So(result, ShouldEqual, "test\\")
		})

		Convey("成功 - 转义普通字符", func() {
			result := ReplaceLikeWildcards("test\\a")
			So(result, ShouldEqual, "test\\a")
		})

		Convey("成功 - 末尾转义符", func() {
			result := ReplaceLikeWildcards("test\\")
			So(result, ShouldEqual, "test\\")
		})

		Convey("成功 - 复杂场景", func() {
			result := ReplaceLikeWildcards("test%_\\%")
			So(result, ShouldEqual, "test.*.%")
		})

		Convey("成功 - 转义后的%和_", func() {
			result := ReplaceLikeWildcards("\\%\\_")
			So(result, ShouldEqual, "%_")
		})
	})
}
