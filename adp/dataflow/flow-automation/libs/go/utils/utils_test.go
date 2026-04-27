package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIsIPv6(t *testing.T) {
	Convey("Is IPv6", t, func() {
		str := "127.0.0.1"
		result := IsIPv6(str)
		So(result, ShouldEqual, false)
	})
	Convey("Not IPv6", t, func() {
		str := "fc99:168::c0a8:c061"
		result := IsIPv6(str)
		So(result, ShouldEqual, true)
	})
}
