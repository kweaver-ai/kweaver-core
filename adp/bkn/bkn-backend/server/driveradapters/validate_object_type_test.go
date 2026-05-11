// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

func Test_ValidateObjectType(t *testing.T) {
	Convey("Test ValidateObjectType\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid object type\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop1",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid ID\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "_invalid_id",
					OTName: "object1",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with empty name\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ObjectType_NullParameter_Name)
		})

		Convey("Failed with empty primary keys\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{},
					DisplayKey:  "prop1",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ObjectType_NullParameter_PrimaryKeys)
		})

		Convey("Failed with invalid primary key type\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "float",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop1",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with empty display key\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ObjectType_NullParameter_DisplayKey)
		})

		Convey("Failed with invalid data source type\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop1",
					DataSource: &interfaces.ResourceInfo{
						Type: "invalid_type",
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with data source type resource\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop1",
					DataSource: &interfaces.ResourceInfo{
						Type: interfaces.DATA_SOURCE_TYPE_RESOURCE,
						ID:   "res-1",
					},
				},
			}
			err := ValidateObjectType(ctx, ot, false)
			So(err, ShouldBeNil)
		})

		Convey("Failed with primary key not in data properties\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop2"},
					DisplayKey:  "prop1",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with display key not in data properties\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop2",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with invalid display key type\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
						{
							Name:        "prop2",
							Type:        "binary",
							DisplayName: "prop2",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop2",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with invalid incremental key type\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
						{
							Name:        "prop2",
							Type:        "float",
							DisplayName: "prop2",
						},
					},
					PrimaryKeys:    []string{"prop1"},
					DisplayKey:     "prop1",
					IncrementalKey: "prop2",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with incremental key not in data properties\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys:    []string{"prop1"},
					DisplayKey:     "prop1",
					IncrementalKey: "prop2",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with valid incremental key\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
						{
							Name:        "prop2",
							Type:        "integer",
							DisplayName: "prop2",
						},
					},
					PrimaryKeys:    []string{"prop1"},
					DisplayKey:     "prop1",
					IncrementalKey: "prop2",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid logic property type\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop1",
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:        "logic1",
							Type:        "invalid_type",
							DisplayName: "logic1",
						},
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with logic property type mismatch with data source\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop1",
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:        "logic1",
							Type:        "metric",
							DisplayName: "logic1",
							DataSource: &interfaces.ResourceInfo{
								Type: "operator",
								ID:   "res1",
							},
						},
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with logic property empty parameter name\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop1",
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:        "logic1",
							Type:        "metric",
							DisplayName: "logic1",
							DataSource: &interfaces.ResourceInfo{
								Type: "metric",
								ID:   "metric-model-1",
							},
							Parameters: []interfaces.Parameter{
								{
									Name: "",
								},
							},
						},
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with empty logic property type\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop1",
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:        "logic1",
							Type:        "",
							DisplayName: "logic1",
						},
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with logic property data source missing id\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop1",
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:        "logic1",
							Type:        "metric",
							DisplayName: "logic1",
							DataSource: &interfaces.ResourceInfo{
								Type: "metric",
								ID:   "",
							},
						},
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with valid logic property metric type\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "prop1",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop1",
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:        "logic1",
							Type:        "metric",
							DisplayName: "logic1",
							DataSource: &interfaces.ResourceInfo{
								Type: "metric",
								ID:   "metric-model-1",
							},
						},
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid data property\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "",
							Type:        "string",
							DisplayName: "prop1",
						},
					},
					PrimaryKeys: []string{"prop1"},
					DisplayKey:  "prop1",
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_ValidatePropertyName(t *testing.T) {
	Convey("Test ValidatePropertyName\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid property name\n", func() {
			err := ValidatePropertyName(ctx, "validProp")
			So(err, ShouldBeNil)
		})

		Convey("Failed with empty property name\n", func() {
			err := ValidatePropertyName(ctx, "")
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, berrors.BknBackend_ObjectType_NullParameter_PropertyName)
		})

		Convey("Failed with property name starting with underscore\n", func() {
			err := ValidatePropertyName(ctx, "_invalidProp")
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_ValidateDataProperties(t *testing.T) {
	Convey("Test ValidateDataProperties\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid data properties\n", func() {
			propertyNames := []string{"prop1", "prop2"}
			dataProperties := []*interfaces.DataProperty{
				{
					Name:        "prop1",
					Type:        "string",
					DisplayName: "prop1",
				},
				{
					Name:        "prop2",
					Type:        "integer",
					DisplayName: "prop2",
				},
			}
			err := ValidateDataProperties(ctx, propertyNames, dataProperties, true)
			So(err, ShouldBeNil)
		})

		Convey("Failed with length mismatch\n", func() {
			propertyNames := []string{"prop1"}
			dataProperties := []*interfaces.DataProperty{
				{
					Name:        "prop1",
					Type:        "string",
					DisplayName: "prop1",
				},
				{
					Name:        "prop2",
					Type:        "integer",
					DisplayName: "prop2",
				},
			}
			err := ValidateDataProperties(ctx, propertyNames, dataProperties, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with property not in URL\n", func() {
			propertyNames := []string{"prop1"}
			dataProperties := []*interfaces.DataProperty{
				{
					Name:        "prop2",
					Type:        "string",
					DisplayName: "prop2",
				},
			}
			err := ValidateDataProperties(ctx, propertyNames, dataProperties, true)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_ValidateDataProperty(t *testing.T) {
	Convey("Test ValidateDataProperty\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid data property\n", func() {
			prop := &interfaces.DataProperty{
				Name:        "prop1",
				Type:        "string",
				DisplayName: "prop1",
			}
			err := ValidateDataProperty(ctx, prop, true)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid property type\n", func() {
			prop := &interfaces.DataProperty{
				Name:        "prop1",
				Type:        "invalid_type",
				DisplayName: "prop1",
			}
			err := ValidateDataProperty(ctx, prop, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with empty mapped field name\n", func() {
			prop := &interfaces.DataProperty{
				Name:        "prop1",
				Type:        "string",
				DisplayName: "prop1",
				MappedField: &interfaces.Field{
					Name: "",
				},
			}
			err := ValidateDataProperty(ctx, prop, true)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_ValidateIndexConfig(t *testing.T) {
	Convey("Test ValidateIndexConfig\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid index config\n", func() {
			config := interfaces.IndexConfig{
				KeywordConfig: interfaces.KeywordConfig{
					Enabled: false,
				},
				FulltextConfig: interfaces.FulltextConfig{
					Enabled: false,
				},
				VectorConfig: interfaces.VectorConfig{
					Enabled: false,
				},
			}
			err := ValidateIndexConfig(ctx, config, true)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid keyword config\n", func() {
			config := interfaces.IndexConfig{
				KeywordConfig: interfaces.KeywordConfig{
					Enabled:        true,
					IgnoreAboveLen: 0,
				},
			}
			err := ValidateIndexConfig(ctx, config, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with invalid fulltext config\n", func() {
			config := interfaces.IndexConfig{
				FulltextConfig: interfaces.FulltextConfig{
					Enabled:  true,
					Analyzer: "invalid_analyzer",
				},
			}
			err := ValidateIndexConfig(ctx, config, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with invalid vector config\n", func() {
			config := interfaces.IndexConfig{
				VectorConfig: interfaces.VectorConfig{
					Enabled: true,
					ModelID: "",
				},
			}
			err := ValidateIndexConfig(ctx, config, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with vector enabled and empty model ID when strictMode is false\n", func() {
			config := interfaces.IndexConfig{
				VectorConfig: interfaces.VectorConfig{
					Enabled: true,
					ModelID: "",
				},
			}
			err := ValidateIndexConfig(ctx, config, false)
			So(err, ShouldBeNil)
		})
	})
}

func Test_ValidateKeywordConfig(t *testing.T) {
	Convey("Test ValidateKeywordConfig\n", t, func() {
		ctx := context.Background()

		Convey("Success with disabled config\n", func() {
			config := interfaces.KeywordConfig{
				Enabled: false,
			}
			err := ValidateKeywordConfig(ctx, config)
			So(err, ShouldBeNil)
		})

		Convey("Success with valid enabled config\n", func() {
			config := interfaces.KeywordConfig{
				Enabled:        true,
				IgnoreAboveLen: 10,
			}
			err := ValidateKeywordConfig(ctx, config)
			So(err, ShouldBeNil)
		})

		Convey("Failed with invalid ignore above len\n", func() {
			config := interfaces.KeywordConfig{
				Enabled:        true,
				IgnoreAboveLen: 0,
			}
			err := ValidateKeywordConfig(ctx, config)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_ValidateFulltextConfig(t *testing.T) {
	Convey("Test ValidateFulltextConfig\n", t, func() {
		ctx := context.Background()

		Convey("Success with disabled config\n", func() {
			config := interfaces.FulltextConfig{
				Enabled: false,
			}
			err := ValidateFulltextConfig(ctx, config)
			So(err, ShouldBeNil)
		})

		Convey("Success with valid analyzer\n", func() {
			validAnalyzers := []string{"standard", "english", "ik_max_word", "hanlp_standard", "hanlp_index"}
			for _, analyzer := range validAnalyzers {
				config := interfaces.FulltextConfig{
					Enabled:  true,
					Analyzer: analyzer,
				}
				err := ValidateFulltextConfig(ctx, config)
				So(err, ShouldBeNil)
			}
		})

		Convey("Failed with invalid analyzer\n", func() {
			config := interfaces.FulltextConfig{
				Enabled:  true,
				Analyzer: "invalid_analyzer",
			}
			err := ValidateFulltextConfig(ctx, config)
			So(err, ShouldNotBeNil)
		})
	})
}

func Test_ValidateVectorConfig(t *testing.T) {
	Convey("Test ValidateVectorConfig\n", t, func() {
		ctx := context.Background()

		Convey("Success with disabled config\n", func() {
			config := interfaces.VectorConfig{
				Enabled: false,
			}
			err := ValidateVectorConfig(ctx, config, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with valid enabled config\n", func() {
			config := interfaces.VectorConfig{
				Enabled: true,
				ModelID: "model1",
			}
			err := ValidateVectorConfig(ctx, config, true)
			So(err, ShouldBeNil)
		})

		Convey("Failed with empty model ID when strict\n", func() {
			config := interfaces.VectorConfig{
				Enabled: true,
				ModelID: "",
			}
			err := ValidateVectorConfig(ctx, config, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with empty model ID when not strict\n", func() {
			config := interfaces.VectorConfig{
				Enabled: true,
				ModelID: "",
			}
			err := ValidateVectorConfig(ctx, config, false)
			So(err, ShouldBeNil)
		})
	})
}

func Test_ValidateRidLogicProperty(t *testing.T) {
	Convey("Test ValidateRidLogicProperty\n", t, func() {
		ctx := context.Background()

		Convey("Success with valid rid logic property\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "skill_id",
							Type:        "string",
							DisplayName: "Skill ID",
						},
					},
					PrimaryKeys: []string{"skill_id"},
					DisplayKey:  "skill_id",
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:        "skillRef",
							DisplayName: "Skill Reference",
							Type:        "rid",
							Kind:        "skill",
							Field:       "skill_id",
						},
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with all valid rid kinds\n", func() {
			validKinds := []string{"skill", "tool", "operator", "agent"}
			for _, kind := range validKinds {
				ot := &interfaces.ObjectType{
					ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
						OTID:   "ot1",
						OTName: "object1",
						DataProperties: []*interfaces.DataProperty{
							{
								Name:        "res_id",
								Type:        "string",
								DisplayName: "Resource ID",
							},
						},
						PrimaryKeys: []string{"res_id"},
						DisplayKey:  "res_id",
						LogicProperties: []*interfaces.LogicProperty{
							{
								Name:        "resRef",
								DisplayName: kind + " Reference",
								Type:        "rid",
								Kind:        kind,
								Field:       "res_id",
							},
						},
					},
				}
				err := ValidateObjectType(ctx, ot, true)
				So(err, ShouldBeNil)
			}
		})

		Convey("Failed with empty kind\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "skill_id",
							Type:        "string",
							DisplayName: "Skill ID",
						},
					},
					PrimaryKeys: []string{"skill_id"},
					DisplayKey:  "skill_id",
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:        "skillRef",
							DisplayName: "Skill Reference",
							Type:        "rid",
							Kind:        "",
							Field:       "skill_id",
						},
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with invalid kind\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "skill_id",
							Type:        "string",
							DisplayName: "Skill ID",
						},
					},
					PrimaryKeys: []string{"skill_id"},
					DisplayKey:  "skill_id",
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:        "skillRef",
							DisplayName: "Skill Reference",
							Type:        "rid",
							Kind:        "invalid_kind",
							Field:       "skill_id",
						},
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with empty field\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "skill_id",
							Type:        "string",
							DisplayName: "Skill ID",
						},
					},
					PrimaryKeys: []string{"skill_id"},
					DisplayKey:  "skill_id",
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:        "skillRef",
							DisplayName: "Skill Reference",
							Type:        "rid",
							Kind:        "skill",
							Field:       "",
						},
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with invalid rid type\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "skill_id",
							Type:        "string",
							DisplayName: "Skill ID",
						},
					},
					PrimaryKeys: []string{"skill_id"},
					DisplayKey:  "skill_id",
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:        "skillRef",
							DisplayName: "Skill Reference",
							Type:        "invalid_type",
						},
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with rid and metric mixed logic properties\n", func() {
			ot := &interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID:   "ot1",
					OTName: "object1",
					DataProperties: []*interfaces.DataProperty{
						{
							Name:        "skill_id",
							Type:        "string",
							DisplayName: "Skill ID",
						},
					},
					PrimaryKeys: []string{"skill_id"},
					DisplayKey:  "skill_id",
					LogicProperties: []*interfaces.LogicProperty{
						{
							Name:        "skillRef",
							DisplayName: "Skill Reference",
							Type:        "rid",
							Kind:        "skill",
							Field:       "skill_id",
						},
						{
							Name:        "metric1",
							DisplayName: "Metric 1",
							Type:        "metric",
							DataSource: &interfaces.ResourceInfo{
								Type: "metric",
								ID:   "metric_ds_1",
							},
							Parameters: []interfaces.Parameter{
								{Name: "param1", Type: "string"},
							},
						},
					},
				},
			}
			err := ValidateObjectType(ctx, ot, true)
			So(err, ShouldBeNil)
		})
	})
}
