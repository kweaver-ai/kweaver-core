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

func TestNewNotLikeCond(t *testing.T) {
	Convey("Test NewNotLikeCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*FieldCfg{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
			"field2": {
				Name: "field2",
				Type: dtype.DATATYPE_TEXT,
			},
			"field3": {
				Name: "field3",
				Type: dtype.DATATYPE_INTEGER,
			},
		}

		Convey("non-string field should return error", func() {
			cfg := &CondCfg{
				Operation: OperationNotLike,
				Field:     "field3",
				FieldCfg:  fieldsMap["field3"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewNotLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "not a string field")
		})

		Convey("invalid value_from should return error", func() {
			cfg := &CondCfg{
				Operation: OperationNotLike,
				Field:     "field1",
				FieldCfg:  fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Field,
					Value:     "value1",
				},
			}
			cond, err := NewNotLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "does not support value_from type")
		})

		Convey("non-string value should return error", func() {
			cfg := &CondCfg{
				Operation: OperationNotLike,
				Field:     "field1",
				FieldCfg:  fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     123,
				},
			}
			cond, err := NewNotLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "not a string value")
		})

		Convey("valid string field and value should create NotLikeCond", func() {
			cfg := &CondCfg{
				Operation: OperationNotLike,
				Field:     "field1",
				FieldCfg:  fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewNotLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			notLikeCond, ok := cond.(*NotLikeCond)
			So(ok, ShouldBeTrue)
			So(notLikeCond.mValue, ShouldEqual, "value1")
		})
	})
}

func TestNotLikeCond_Convert(t *testing.T) {
	Convey("Test NotLikeCond.Convert", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*FieldCfg{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("should create must_not regexp query", func() {
			cfg := &CondCfg{
				Operation: OperationNotLike,
				Field:     "field1",
				FieldCfg:  fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, _ := NewNotLikeCond(ctx, cfg, fieldsMap)
			notLikeCond := cond.(*NotLikeCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := notLikeCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "bool")
			So(dsl, ShouldContainSubstring, "must_not")
			So(dsl, ShouldContainSubstring, "regexp")
			So(dsl, ShouldContainSubstring, "field1")
			So(dsl, ShouldContainSubstring, ".*test.*")
		})
	})
}

func TestNotLikeCond_Convert2SQL(t *testing.T) {
	Convey("Test NotLikeCond.Convert2SQL", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*FieldCfg{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("should create NOT LIKE query", func() {
			cfg := &CondCfg{
				Operation: OperationNotLike,
				Field:     "field1",
				FieldCfg:  fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, _ := NewNotLikeCond(ctx, cfg, fieldsMap)
			notLikeCond := cond.(*NotLikeCond)

			sql, err := notLikeCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldNotBeEmpty)
			So(sql, ShouldContainSubstring, "field1")
			So(sql, ShouldContainSubstring, "NOT LIKE")
			So(sql, ShouldContainSubstring, "%test%")
		})
	})
}
