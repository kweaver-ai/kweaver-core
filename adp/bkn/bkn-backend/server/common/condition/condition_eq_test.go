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

func TestNewEqCond(t *testing.T) {
	Convey("Test NewEqCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*FieldCfg{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("invalid value_from should return error", func() {
			cfg := &CondCfg{
				Operation: OperationEq,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Field,
					Value:     "value1",
				},
			}
			cond, err := NewEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "does not support value_from type")
		})

		Convey("slice value should return error", func() {
			cfg := &CondCfg{
				Operation: OperationEq,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     []any{"value1", "value2"},
				},
			}
			cond, err := NewEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "only supports single value")
		})

		Convey("valid config should create EqCond", func() {
			cfg := &CondCfg{
				Operation: OperationEq,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			eqCond, ok := cond.(*EqCond)
			So(ok, ShouldBeTrue)
			So(eqCond.mFilterFieldName, ShouldEqual, "field1")
		})

		Convey("numeric value should work", func() {
			cfg := &CondCfg{
				Operation: OperationEq,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     123,
				},
			}
			cond, err := NewEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})
	})
}

func TestEqCond_Convert(t *testing.T) {
	Convey("Test EqCond.Convert", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*FieldCfg{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("string value should be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationEq,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, _ := NewEqCond(ctx, cfg, fieldsMap)
			eqCond := cond.(*EqCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := eqCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "term")
			So(dsl, ShouldContainSubstring, "field1")
			So(dsl, ShouldContainSubstring, `"value1"`)
		})

		Convey("numeric value should not be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationEq,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     123,
				},
			}
			cond, _ := NewEqCond(ctx, cfg, fieldsMap)
			eqCond := cond.(*EqCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := eqCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "term")
			So(dsl, ShouldContainSubstring, "123")
		})
	})
}

func TestEqCond_Convert2SQL(t *testing.T) {
	Convey("Test EqCond.Convert2SQL", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*FieldCfg{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("string value should be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationEq,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, _ := NewEqCond(ctx, cfg, fieldsMap)
			eqCond := cond.(*EqCond)

			sql, err := eqCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldNotBeEmpty)
			So(sql, ShouldContainSubstring, "field1")
			So(sql, ShouldContainSubstring, "=")
			So(sql, ShouldContainSubstring, "'value1'")
		})

		Convey("numeric value should not be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationEq,
				Field:     "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     123,
				},
			}
			cond, _ := NewEqCond(ctx, cfg, fieldsMap)
			eqCond := cond.(*EqCond)

			sql, err := eqCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldNotBeEmpty)
			So(sql, ShouldContainSubstring, "field1")
			So(sql, ShouldContainSubstring, "=")
			So(sql, ShouldContainSubstring, "123")
		})
	})
}
