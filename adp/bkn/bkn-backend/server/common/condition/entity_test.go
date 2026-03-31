// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestViewField_InitFieldPath(t *testing.T) {
	Convey("Test ViewField.InitFieldPath", t, func() {
		Convey("empty path should be initialized", func() {
			field := &ViewField{
				Name: "field1.field2",
				Path: []string{},
			}
			field.InitFieldPath()
			So(len(field.Path), ShouldEqual, 2)
			So(field.Path[0], ShouldEqual, "field1")
			So(field.Path[1], ShouldEqual, "field2")
		})

		Convey("existing path should not be changed", func() {
			field := &ViewField{
				Name: "field1.field2",
				Path: []string{"existing", "path"},
			}
			field.InitFieldPath()
			So(len(field.Path), ShouldEqual, 2)
			So(field.Path[0], ShouldEqual, "existing")
			So(field.Path[1], ShouldEqual, "path")
		})

		Convey("simple field name should create single element path", func() {
			field := &ViewField{
				Name: "field1",
				Path: []string{},
			}
			field.InitFieldPath()
			So(len(field.Path), ShouldEqual, 1)
			So(field.Path[0], ShouldEqual, "field1")
		})
	})
}

func TestIsSlice(t *testing.T) {
	Convey("Test IsSlice", t, func() {
		Convey("slice should return true", func() {
			So(IsSlice([]string{"a", "b"}), ShouldBeTrue)
			So(IsSlice([]int{1, 2}), ShouldBeTrue)
			So(IsSlice([]any{"a", 1}), ShouldBeTrue)
		})

		Convey("array should return true", func() {
			So(IsSlice([2]string{"a", "b"}), ShouldBeTrue)
			So(IsSlice([3]int{1, 2, 3}), ShouldBeTrue)
		})

		Convey("non-slice should return false", func() {
			So(IsSlice("string"), ShouldBeFalse)
			So(IsSlice(123), ShouldBeFalse)
			So(IsSlice(true), ShouldBeFalse)
			So(IsSlice(map[string]string{}), ShouldBeFalse)
		})
	})
}

func TestIsSameType(t *testing.T) {
	Convey("Test IsSameType", t, func() {
		Convey("empty array should return true", func() {
			So(IsSameType([]any{}), ShouldBeTrue)
		})

		Convey("same type elements should return true", func() {
			So(IsSameType([]any{"a", "b", "c"}), ShouldBeTrue)
			So(IsSameType([]any{1, 2, 3}), ShouldBeTrue)
			So(IsSameType([]any{1.1, 2.2, 3.3}), ShouldBeTrue)
		})

		Convey("different type elements should return false", func() {
			So(IsSameType([]any{"a", 1}), ShouldBeFalse)
			So(IsSameType([]any{1, "b"}), ShouldBeFalse)
			So(IsSameType([]any{1, 2.2}), ShouldBeFalse)
			So(IsSameType([]any{"a", true}), ShouldBeFalse)
		})

		Convey("single element should return true", func() {
			So(IsSameType([]any{"a"}), ShouldBeTrue)
			So(IsSameType([]any{1}), ShouldBeTrue)
		})
	})
}
