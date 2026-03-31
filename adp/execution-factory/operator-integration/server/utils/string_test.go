//go:build !skiptest
// +build !skiptest

package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUniqueStrings(t *testing.T) {
	Convey("TestUniqueStrings", t, func() {
		str := []string{"a", "a", "b", "c", "d", "d"}
		uniqueStr := UniqueStrings(str)
		So(len(uniqueStr), ShouldEqual, 4)
		So(uniqueStr[0], ShouldEqual, "a")
		So(uniqueStr[1], ShouldEqual, "b")
		So(uniqueStr[2], ShouldEqual, "c")
		So(uniqueStr[3], ShouldEqual, "d")
	})
}

func TestObjectToByte(t *testing.T) {
	Convey("TestObjectToByte", t, func() {
		byt := ObjectToByte(nil)
		So(byt, ShouldNotEqual, nil)
	})
}
func TestObjectToJSON(t *testing.T) {
	Convey("ObjectToJSON", t, func() {
		byt := ObjectToJSON(nil)
		So(byt, ShouldEqual, "null")
	})
}
