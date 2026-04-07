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

func TestNewNotEqCond(t *testing.T) {
	Convey("Test NewNotEqCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("invalid value_from should return error", func() {
			cfg := &CondCfg{
				Operation: OperationNotEq,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Field,
					Value:     "value1",
				},
			}
			cond, err := NewNotEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "does not support value_from type")
		})

		Convey("slice value should return error", func() {
			cfg := &CondCfg{
				Operation: OperationNotEq,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     []any{"value1", "value2"},
				},
			}
			cond, err := NewNotEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "only supports single value")
		})

		Convey("valid config should create NotEqCond", func() {
			cfg := &CondCfg{
				Operation: OperationNotEq,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewNotEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			notEqCond, ok := cond.(*NotEqCond)
			So(ok, ShouldBeTrue)
			So(notEqCond.mFilterFieldName, ShouldEqual, "field1")
		})
	})
}

func TestNotEqCond_Convert(t *testing.T) {
	Convey("Test NotEqCond.Convert", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("string value should be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationNotEq,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, _ := NewNotEqCond(ctx, cfg, fieldsMap)
			notEqCond := cond.(*NotEqCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := notEqCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "bool")
			So(dsl, ShouldContainSubstring, "must_not")
			So(dsl, ShouldContainSubstring, "term")
			So(dsl, ShouldContainSubstring, "field1")
			So(dsl, ShouldContainSubstring, `"value1"`)
		})

		Convey("numeric value should not be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationNotEq,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     123,
				},
			}
			cond, _ := NewNotEqCond(ctx, cfg, fieldsMap)
			notEqCond := cond.(*NotEqCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := notEqCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "bool")
			So(dsl, ShouldContainSubstring, "must_not")
			So(dsl, ShouldContainSubstring, "123")
		})
	})
}

func TestNotEqCond_Convert2SQL(t *testing.T) {
	Convey("Test NotEqCond.Convert2SQL", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*ViewField{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("string value should be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationNotEq,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, _ := NewNotEqCond(ctx, cfg, fieldsMap)
			notEqCond := cond.(*NotEqCond)

			sql, err := notEqCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldNotBeEmpty)
			So(sql, ShouldContainSubstring, "field1")
			So(sql, ShouldContainSubstring, "<>")
			So(sql, ShouldContainSubstring, "'value1'")
		})

		Convey("numeric value should not be quoted", func() {
			cfg := &CondCfg{
				Operation: OperationNotEq,
				Name:      "field1",
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     123,
				},
			}
			cond, _ := NewNotEqCond(ctx, cfg, fieldsMap)
			notEqCond := cond.(*NotEqCond)

			sql, err := notEqCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldNotBeEmpty)
			So(sql, ShouldContainSubstring, "field1")
			So(sql, ShouldContainSubstring, "<>")
			So(sql, ShouldContainSubstring, "123")
		})
	})
}
