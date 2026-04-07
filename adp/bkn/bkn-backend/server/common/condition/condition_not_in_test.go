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

func TestNewNotInCond(t *testing.T) {
	Convey("Test NewNotInCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("invalid value_from should return error", func() {
			cfg := &CondCfg{
				Operation: OperationNotIn,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Field,
					Value:     []any{"value1"},
				},
			}
			cond, err := NewNotInCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "does not support value_from type")
		})

		Convey("non-slice value should return error", func() {
			cfg := &CondCfg{
				Operation: OperationNotIn,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewNotInCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "should be an array")
		})

		Convey("mixed type array should return error", func() {
			cfg := &CondCfg{
				Operation: OperationNotIn,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     []any{"value1", 123},
				},
			}
			cond, err := NewNotInCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "same type")
		})

		Convey("valid string array should create NotInCond", func() {
			cfg := &CondCfg{
				Operation: OperationNotIn,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     []any{"value1", "value2"},
				},
			}
			cond, err := NewNotInCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			notInCond, ok := cond.(*NotInCond)
			So(ok, ShouldBeTrue)
			So(len(notInCond.mValue), ShouldEqual, 2)
		})
	})
}

func TestNotInCond_Convert(t *testing.T) {
	Convey("Test NotInCond.Convert", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("string array should work", func() {
			cfg := &CondCfg{
				Operation: OperationNotIn,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     []any{"value1", "value2"},
				},
			}
			cond, _ := NewNotInCond(ctx, cfg, fieldsMap)
			notInCond := cond.(*NotInCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := notInCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "bool")
			So(dsl, ShouldContainSubstring, "must_not")
			So(dsl, ShouldContainSubstring, "terms")
			So(dsl, ShouldContainSubstring, "field1")
		})
	})
}

func TestNotInCond_Convert2SQL(t *testing.T) {
	Convey("Test NotInCond.Convert2SQL", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("string array should work", func() {
			cfg := &CondCfg{
				Operation: OperationNotIn,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     []any{"value1", "value2"},
				},
			}
			cond, _ := NewNotInCond(ctx, cfg, fieldsMap)
			notInCond := cond.(*NotInCond)

			sql, err := notInCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldNotBeEmpty)
			So(sql, ShouldContainSubstring, "field1")
			So(sql, ShouldContainSubstring, "NOT IN")
			So(sql, ShouldContainSubstring, "'value1'")
			So(sql, ShouldContainSubstring, "'value2'")
		})

		Convey("numeric array should work", func() {
			cfg := &CondCfg{
				Operation: OperationNotIn,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     []any{1, 2, 3},
				},
			}
			cond, _ := NewNotInCond(ctx, cfg, fieldsMap)
			notInCond := cond.(*NotInCond)

			sql, err := notInCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldNotBeEmpty)
			So(sql, ShouldContainSubstring, "NOT IN")
		})
	})
}
