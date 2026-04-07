// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package utils

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestUniqueStrings(t *testing.T) {
	convey.Convey("TestUniqueStrings", t, func() {
		str := []string{"a", "a", "b", "c", "d", "d"}
		uniqueStr := UniqueStrings(str)
		convey.So(len(uniqueStr), convey.ShouldEqual, 4)
		convey.So(uniqueStr[0], convey.ShouldEqual, "a")
		convey.So(uniqueStr[1], convey.ShouldEqual, "b")
		convey.So(uniqueStr[2], convey.ShouldEqual, "c")
		convey.So(uniqueStr[3], convey.ShouldEqual, "d")
	})
}

func TestObjectToByte(t *testing.T) {
	convey.Convey("TestObjectToByte", t, func() {
		byt := ObjectToByte(nil)
		convey.So(byt, convey.ShouldNotEqual, nil)
	})
}
func TestObjectToJSON(t *testing.T) {
	convey.Convey("ObjectToJSON", t, func() {
		byt := ObjectToJSON(nil)
		convey.So(byt, convey.ShouldEqual, "null")
	})
}
