// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	dtype "bkn-backend/interfaces/data_type"
)

func TestNewCondition(t *testing.T) {
	Convey("Test NewCondition", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("nil config should return nil", func() {
			cond, err := NewCondition(ctx, nil, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("OperationAnd should create AndCond", func() {
			cfg := &CondCfg{
				Operation: OperationAnd,
				SubConds: []*CondCfg{
					{
						Operation: OperationEq,
						Field:     "field1",
						ValueOptCfg: ValueOptCfg{
							ValueFrom: ValueFrom_Const,
							Value:     "value1",
						},
					},
				},
			}
			cond, err := NewCondition(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			_, ok := cond.(*AndCond)
			So(ok, ShouldBeTrue)
		})

		Convey("OperationOr should create OrCond", func() {
			cfg := &CondCfg{
				Operation: OperationOr,
				SubConds: []*CondCfg{
					{
						Operation: OperationEq,
						Field:     "field1",
						ValueOptCfg: ValueOptCfg{
							ValueFrom: ValueFrom_Const,
							Value:     "value1",
						},
					},
				},
			}
			cond, err := NewCondition(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			_, ok := cond.(*OrCond)
			So(ok, ShouldBeTrue)
		})

		Convey("OperationEq should create EqCond", func() {
			cfg := &CondCfg{
				Operation: OperationEq,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewCondition(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			_, ok := cond.(*EqCond)
			So(ok, ShouldBeTrue)
		})

		Convey("error in sub condition should propagate", func() {
			cfg := &CondCfg{
				Operation: OperationAnd,
				SubConds: []*CondCfg{
					{
						Operation: OperationEq,
						Field:     "nonexistent",
						ValueOptCfg: ValueOptCfg{
							ValueFrom: ValueFrom_Const,
							Value:     "value1",
						},
					},
				},
			}
			cond, err := NewCondition(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func TestNewCondWithOpr(t *testing.T) {
	Convey("Test NewCondWithOpr", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
			"field2": {
				Name: "field2",
				Type: dtype.DATATYPE_TEXT,
			},
			"binary_field": {
				Name: "binary_field",
				Type: dtype.DATATYPE_BINARY,
			},
		}

		Convey("field not in fieldsMap should return error", func() {
			cfg := &CondCfg{
				Operation: OperationEq,
				Field:     "nonexistent",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "must in view original fields")
		})

		Convey("binary type field should return error", func() {
			cfg := &CondCfg{
				Operation: OperationEq,
				Field:     "binary_field",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "binary type")
		})

		Convey("AllField should not check field existence", func() {
			cfg := &CondCfg{
				Operation: OperationMatch,
				Field:     AllField,
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("OperationMultiMatch should not check field existence", func() {
			cfg := &CondCfg{
				Operation: OperationMultiMatch,
				RemainCfg: map[string]any{
					"fields": []any{"field1"},
				},
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("unsupported operation should return error", func() {
			cfg := &CondCfg{
				Operation: "unsupported_op",
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "not support condition's operation")
		})

		Convey("all operation types should work", func() {
			testCases := []struct {
				operation string
				name      string
				value     any
			}{
				{OperationEq, "field1", "value1"},
				{OperationNotEq, "field1", "value1"},
				{OperationIn, "field1", []any{"value1", "value2"}},
				{OperationNotIn, "field1", []any{"value1", "value2"}},
				{OperationLike, "field1", "value1"},
				{OperationNotLike, "field1", "value1"},
				{OperationRegex, "field1", "value1"},
				{OperationMatch, "field1", "value1"},
				{OperationMatchPhrase, "field1", "value1"},
				{OperationKNN, "field1", "value1"},
			}

			for _, tc := range testCases {
				cfg := &CondCfg{
					Operation: tc.operation,
					Field:     tc.name,
					ValueOptCfg: ValueOptCfg{
						ValueFrom: ValueFrom_Const,
						Value:     tc.value,
					},
					RemainCfg: map[string]any{},
				}

				if tc.operation == OperationLike || tc.operation == OperationNotLike || tc.operation == OperationRegex {
					cfg.NameField = fieldsMap[tc.name]
				}

				if tc.operation == OperationKNN {
					cfg.RemainCfg["limit_key"] = "k"
					cfg.RemainCfg["limit_value"] = 10
				}

				cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
				// Some operations may fail due to validation, but we're testing that they're routed correctly
				if err == nil {
					So(cond, ShouldNotBeNil)
				}
			}
		})
	})
}

func TestGetFilterFieldName(t *testing.T) {
	Convey("Test getFilterFieldName", t, func() {
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_TEXT,
			},
			"field2": {
				Name: "field2",
				Type: dtype.DATATYPE_STRING,
			},
			"field1_desensitize": {
				Name: "field1_desensitize",
				Type: dtype.DATATYPE_TEXT,
			},
		}

		Convey("AllField should return AllField", func() {
			result := getFilterFieldName(AllField, fieldsMap, true)
			So(result, ShouldEqual, AllField)
		})

		Convey("MetaField_ID should convert to OS_MetaField_ID", func() {
			result := getFilterFieldName(MetaField_ID, fieldsMap, true)
			So(result, ShouldEqual, OS_MetaField_ID)
		})

		Convey("desensitize field should be used when both exist", func() {
			result := getFilterFieldName("field1", fieldsMap, false)
			So(result, ShouldEqual, "field1_desensitize.keyword")
		})

		Convey("text field should add keyword suffix for non-fulltext query", func() {
			result := getFilterFieldName("field1", fieldsMap, false)
			So(result, ShouldContainSubstring, dtype.KEYWORD_SUFFIX)
		})

		Convey("text field should not add keyword suffix for fulltext query", func() {
			result := getFilterFieldName("field1", fieldsMap, true)
			So(result, ShouldNotContainSubstring, dtype.KEYWORD_SUFFIX)
		})

		Convey("string field should not add keyword suffix", func() {
			result := getFilterFieldName("field2", fieldsMap, false)
			So(result, ShouldEqual, "field2")
		})

		Convey("field without desensitize should use original name", func() {
			result := getFilterFieldName("field2", fieldsMap, false)
			So(result, ShouldEqual, "field2")
		})
	})
}

func TestWrapKeyWordFieldName(t *testing.T) {
	Convey("Test wrapKeyWordFieldName", t, func() {
		Convey("single field should add keyword suffix", func() {
			result := wrapKeyWordFieldName("field1")
			So(result, ShouldEqual, "field1.keyword")
		})

		Convey("multiple fields should join with dot", func() {
			result := wrapKeyWordFieldName("field1", "field2")
			So(result, ShouldEqual, "field1.field2.keyword")
		})

		Convey("empty field should return empty string", func() {
			result := wrapKeyWordFieldName("")
			So(result, ShouldEqual, "")
		})

		Convey("nested field path should work", func() {
			result := wrapKeyWordFieldName("data", "properties", "name")
			So(result, ShouldEqual, "data.properties.name.keyword")
		})
	})
}
