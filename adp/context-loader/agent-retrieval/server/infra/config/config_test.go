// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package config

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestParseAuthEnabled(t *testing.T) {
	convey.Convey("parseAuthEnabled", t, func() {
		cases := []struct {
			input    string
			expected bool
			desc     string
		}{
			{"", true, "empty string defaults to enabled"},
			{"true", true, "explicit true"},
			{"TRUE", true, "uppercase TRUE"},
			{"True", true, "mixed case True"},
			{"1", true, "numeric 1"},
			{"yes", true, "unrecognized value defaults to enabled"},
			{"on", true, "unrecognized 'on' defaults to enabled"},
			{"  true  ", true, "trimmed true"},
			{"false", false, "explicit false"},
			{"FALSE", false, "uppercase FALSE"},
			{"False", false, "mixed case False"},
			{" false ", false, "trimmed false"},
			{"0", false, "numeric 0"},
			{" 0 ", false, "trimmed 0"},
		}

		for _, tc := range cases {
			convey.Convey(tc.desc, func() {
				result := parseAuthEnabled(tc.input)
				convey.So(result, convey.ShouldEqual, tc.expected)
			})
		}
	})
}
