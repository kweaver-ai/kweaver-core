// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"vega-backend/interfaces"
)

var testSortTypes = map[string]string{
	"name":        "f_name",
	"create_time": "f_create_time",
}

func Test_Validate_Name(t *testing.T) {
	Convey("Test validateName\n", t, func() {
		Convey("Valid name\n", func() {
			err := validateName(context.Background(), "test-catalog")
			So(err, ShouldBeNil)
		})

		Convey("Empty name\n", func() {
			err := validateName(context.Background(), "")
			So(err, ShouldNotBeNil)
		})

		Convey("Max length name\n", func() {
			name := strings.Repeat("a", interfaces.NAME_MAX_LENGTH)
			err := validateName(context.Background(), name)
			So(err, ShouldBeNil)
		})

		Convey("Exceeds max length\n", func() {
			name := strings.Repeat("a", interfaces.NAME_MAX_LENGTH+1)
			err := validateName(context.Background(), name)
			So(err, ShouldNotBeNil)
		})

		Convey("UTF-8 name length\n", func() {
			name := strings.Repeat("中", interfaces.NAME_MAX_LENGTH)
			err := validateName(context.Background(), name)
			So(err, ShouldBeNil)

			name = strings.Repeat("中", interfaces.NAME_MAX_LENGTH+1)
			err = validateName(context.Background(), name)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_Validate_Tags(t *testing.T) {
	Convey("Test ValidateTags\n", t, func() {
		Convey("Valid tags\n", func() {
			err := ValidateTags(context.Background(), []string{"tag1", "tag2"})
			So(err, ShouldBeNil)
		})

		Convey("Empty tags\n", func() {
			err := ValidateTags(context.Background(), []string{})
			So(err, ShouldBeNil)
		})

		Convey("Exceeds max number\n", func() {
			tags := make([]string, interfaces.TAGS_MAX_NUMBER+1)
			for i := range tags {
				tags[i] = "tag"
			}
			err := ValidateTags(context.Background(), tags)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid tag in list\n", func() {
			err := ValidateTags(context.Background(), []string{"good-tag", "bad/tag"})
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_Validate_Tag(t *testing.T) {
	Convey("Test validateTag\n", t, func() {
		Convey("Valid tag\n", func() {
			err := validateTag(context.Background(), "my-tag")
			So(err, ShouldBeNil)
		})

		Convey("Empty tag\n", func() {
			err := validateTag(context.Background(), "")
			So(err, ShouldNotBeNil)
		})

		Convey("Only spaces\n", func() {
			err := validateTag(context.Background(), "   ")
			So(err, ShouldNotBeNil)
		})

		Convey("Exceeds max length\n", func() {
			tag := strings.Repeat("a", interfaces.TAG_MAX_LENGTH+1)
			err := validateTag(context.Background(), tag)
			So(err, ShouldNotBeNil)
		})

		Convey("Special chars\n", func() {
			invalidChars := []string{"/", ":", "?", "\\", "\"", "<", ">", "|", "#", "%", "&", "*", "$", "^", "!", "=", "."}
			for _, ch := range invalidChars {
				err := validateTag(context.Background(), "tag"+ch+"name")
				So(err, ShouldNotBeNil)
			}
		})

		Convey("Trim spaces\n", func() {
			err := validateTag(context.Background(), "  valid-tag  ")
			So(err, ShouldBeNil)
		})
	})
}

func Test_Validate_Description(t *testing.T) {
	Convey("Test validateDescription\n", t, func() {
		Convey("Valid description\n", func() {
			err := validateDescription(context.Background(), "A valid description")
			So(err, ShouldBeNil)
		})

		Convey("Empty description\n", func() {
			err := validateDescription(context.Background(), "")
			So(err, ShouldBeNil)
		})

		Convey("Exceeds max length\n", func() {
			desc := strings.Repeat("a", interfaces.DESCRIPTION_MAX_LENGTH+1)
			err := validateDescription(context.Background(), desc)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_Validate_PaginationQueryParams(t *testing.T) {
	Convey("Test validatePaginationQueryParams\n", t, func() {
		Convey("Valid pagination\n", func() {
			params, err := validatePaginationQueryParams(context.Background(),
				"0", "10", "name", "asc", testSortTypes)
			So(err, ShouldBeNil)
			So(params.Offset, ShouldEqual, 0)
			So(params.Limit, ShouldEqual, 10)
			So(params.Sort, ShouldEqual, "f_name")
			So(params.Direction, ShouldEqual, "asc")
		})

		Convey("No limit\n", func() {
			params, err := validatePaginationQueryParams(context.Background(),
				"0", "-1", "name", "desc", testSortTypes)
			So(err, ShouldBeNil)
			So(params.Limit, ShouldEqual, -1)
		})

		Convey("Invalid offset\n", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"abc", "10", "name", "asc", testSortTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Negative offset\n", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"-1", "10", "name", "asc", testSortTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid limit\n", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"0", "abc", "name", "asc", testSortTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Limit too small\n", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"0", "0", "name", "asc", testSortTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Limit too large\n", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"0", "1001", "name", "asc", testSortTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid sort\n", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"0", "10", "unknown_sort", "asc", testSortTypes)
			So(err, ShouldNotBeNil)
		})

		Convey("Invalid direction\n", func() {
			_, err := validatePaginationQueryParams(context.Background(),
				"0", "10", "name", "up", testSortTypes)
			So(err, ShouldNotBeNil)
		})
	})
}
