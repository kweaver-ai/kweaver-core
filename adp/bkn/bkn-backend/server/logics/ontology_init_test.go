// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"bkn-backend/interfaces"
)

// ── comparePropertyFeature ────────────────────────────────────────────────────

func Test_comparePropertyFeature(t *testing.T) {
	Convey("Test comparePropertyFeature\n", t, func() {
		base := &interfaces.PropertyFeature{
			FeatureName: "keyword",
			FeatureType: "keyword",
			RefProperty: "name",
			IsDefault:   true,
			IsNative:    false,
			Config:      map[string]any{"dim": 768},
		}

		Convey("Equal features returns true\n", func() {
			other := &interfaces.PropertyFeature{
				FeatureName: "keyword",
				FeatureType: "keyword",
				RefProperty: "name",
				IsDefault:   true,
				IsNative:    false,
				Config:      map[string]any{"dim": 768},
			}
			So(comparePropertyFeature(base, other), ShouldBeTrue)
		})

		Convey("Different FeatureName returns false\n", func() {
			other := *base
			other.FeatureName = "vector"
			So(comparePropertyFeature(base, &other), ShouldBeFalse)
		})

		Convey("Different FeatureType returns false\n", func() {
			other := *base
			other.FeatureType = "fulltext"
			So(comparePropertyFeature(base, &other), ShouldBeFalse)
		})

		Convey("Different RefProperty returns false\n", func() {
			other := *base
			other.RefProperty = "title"
			So(comparePropertyFeature(base, &other), ShouldBeFalse)
		})

		Convey("Different IsDefault returns false\n", func() {
			other := *base
			other.IsDefault = false
			So(comparePropertyFeature(base, &other), ShouldBeFalse)
		})

		Convey("Different IsNative returns false\n", func() {
			other := *base
			other.IsNative = true
			So(comparePropertyFeature(base, &other), ShouldBeFalse)
		})

		Convey("Different Config length returns false\n", func() {
			other := *base
			other.Config = map[string]any{"dim": 768, "extra": "x"}
			So(comparePropertyFeature(base, &other), ShouldBeFalse)
		})

		Convey("Missing Config key returns false\n", func() {
			other := *base
			other.Config = map[string]any{"other_key": 768}
			So(comparePropertyFeature(base, &other), ShouldBeFalse)
		})

		Convey("Different Config value returns false\n", func() {
			other := *base
			other.Config = map[string]any{"dim": 512}
			So(comparePropertyFeature(base, &other), ShouldBeFalse)
		})

		Convey("Both nil Config returns true\n", func() {
			f1 := &interfaces.PropertyFeature{FeatureName: "f", Config: nil}
			f2 := &interfaces.PropertyFeature{FeatureName: "f", Config: nil}
			So(comparePropertyFeature(f1, f2), ShouldBeTrue)
		})
	})
}

// ── compareProperty ───────────────────────────────────────────────────────────

func Test_compareProperty(t *testing.T) {
	Convey("Test compareProperty\n", t, func() {
		base := &interfaces.Property{
			Name:        "content",
			Type:        "text",
			DisplayName: "Content",
			Description: "desc",
			Features: []interfaces.PropertyFeature{
				{FeatureName: "keyword", FeatureType: "keyword"},
			},
		}

		Convey("Equal properties returns true\n", func() {
			other := &interfaces.Property{
				Name:        "content",
				Type:        "text",
				DisplayName: "Content",
				Description: "desc",
				Features: []interfaces.PropertyFeature{
					{FeatureName: "keyword", FeatureType: "keyword"},
				},
			}
			So(compareProperty(base, other), ShouldBeTrue)
		})

		Convey("Different Name returns false\n", func() {
			other := *base
			other.Name = "title"
			So(compareProperty(base, &other), ShouldBeFalse)
		})

		Convey("Different Type returns false\n", func() {
			other := *base
			other.Type = "keyword"
			So(compareProperty(base, &other), ShouldBeFalse)
		})

		Convey("Different DisplayName returns false\n", func() {
			other := *base
			other.DisplayName = "Other"
			So(compareProperty(base, &other), ShouldBeFalse)
		})

		Convey("Different Description returns false\n", func() {
			other := *base
			other.Description = "other desc"
			So(compareProperty(base, &other), ShouldBeFalse)
		})

		Convey("Different Features length returns false\n", func() {
			other := *base
			other.Features = []interfaces.PropertyFeature{
				{FeatureName: "keyword"},
				{FeatureName: "vector"},
			}
			So(compareProperty(base, &other), ShouldBeFalse)
		})

		Convey("Feature not found in other returns false\n", func() {
			other := *base
			other.Features = []interfaces.PropertyFeature{
				{FeatureName: "fulltext", FeatureType: "fulltext"},
			}
			So(compareProperty(base, &other), ShouldBeFalse)
		})

		Convey("Feature differs in other returns false\n", func() {
			other := *base
			other.Features = []interfaces.PropertyFeature{
				{FeatureName: "keyword", FeatureType: "fulltext"},
			}
			So(compareProperty(base, &other), ShouldBeFalse)
		})

		Convey("No features, all fields equal returns true\n", func() {
			p1 := &interfaces.Property{Name: "id", Type: "long"}
			p2 := &interfaces.Property{Name: "id", Type: "long"}
			So(compareProperty(p1, p2), ShouldBeTrue)
		})
	})
}

// ── deepCompareSchemas ────────────────────────────────────────────────────────

func Test_deepCompareSchemas(t *testing.T) {
	Convey("Test deepCompareSchemas\n", t, func() {
		p1 := &interfaces.Property{Name: "id", Type: "long"}
		p2 := &interfaces.Property{Name: "content", Type: "text"}

		Convey("Equal schemas returns true\n", func() {
			s1 := []*interfaces.Property{p1, p2}
			s2 := []*interfaces.Property{
				{Name: "id", Type: "long"},
				{Name: "content", Type: "text"},
			}
			So(deepCompareSchemas(s1, s2), ShouldBeTrue)
		})

		Convey("Both empty returns true\n", func() {
			So(deepCompareSchemas(nil, nil), ShouldBeTrue)
		})

		Convey("Different lengths returns false\n", func() {
			So(deepCompareSchemas([]*interfaces.Property{p1}, []*interfaces.Property{p1, p2}), ShouldBeFalse)
		})

		Convey("Same length but missing property name in schema2 returns false\n", func() {
			s1 := []*interfaces.Property{{Name: "id", Type: "long"}}
			s2 := []*interfaces.Property{{Name: "other", Type: "long"}}
			So(deepCompareSchemas(s1, s2), ShouldBeFalse)
		})

		Convey("Same properties but different field value returns false\n", func() {
			s1 := []*interfaces.Property{{Name: "id", Type: "long"}}
			s2 := []*interfaces.Property{{Name: "id", Type: "keyword"}}
			So(deepCompareSchemas(s1, s2), ShouldBeFalse)
		})
	})
}
