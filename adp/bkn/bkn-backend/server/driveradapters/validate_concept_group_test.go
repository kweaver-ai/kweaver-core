// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"bkn-backend/interfaces"
)

func Test_ValidateConceptGroup(t *testing.T) {
	Convey("Test ValidateConceptGroup\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid concept group\n", func() {
			cg := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "concept_group1",
			}
			err := ValidateConceptGroup(ctx, cg)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid ID\n", func() {
			cg := &interfaces.ConceptGroup{
				CGID:   "_invalid_id",
				CGName: "concept_group1",
			}
			err := ValidateConceptGroup(ctx, cg)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with empty name\n", func() {
			cg := &interfaces.ConceptGroup{
				CGID:   "cg1",
				CGName: "",
			}
			err := ValidateConceptGroup(ctx, cg)
			So(err, ShouldNotBeNil)
		})
	})
}
