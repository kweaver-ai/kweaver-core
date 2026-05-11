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

func TestNewLikeCond(t *testing.T) {
	Convey("Test NewLikeCond", t, func() {
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
			"field4": {
				Name: "field4",
				Type: dtype.CHAR,
			},
		}

		Convey("non-string field should return error", func() {
			cfg := &CondCfg{
				Operation: OperationLike,
				Field:     "field3",
				FieldCfg:  fieldsMap["field3"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "not a string field")
		})

		Convey("invalid value_from should return error", func() {
			cfg := &CondCfg{
				Operation: OperationLike,
				Field:     "field1",
				FieldCfg:  fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Field,
					Value:     "value1",
				},
			}
			cond, err := NewLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "does not support value_from type")
		})

		Convey("non-string value should return error", func() {
			cfg := &CondCfg{
				Operation: OperationLike,
				Field:     "field1",
				FieldCfg:  fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     123,
				},
			}
			cond, err := NewLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, "not a string value")
		})

		Convey("valid string field and value should create LikeCond", func() {
			cfg := &CondCfg{
				Operation: OperationLike,
				Field:     "field1",
				FieldCfg:  fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
			likeCond, ok := cond.(*LikeCond)
			So(ok, ShouldBeTrue)
			So(likeCond.mValue, ShouldEqual, "value1")
		})

		Convey("text field should work", func() {
			cfg := &CondCfg{
				Operation: OperationLike,
				Field:     "field2",
				FieldCfg:  fieldsMap["field2"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("char field should work", func() {
			cfg := &CondCfg{
				Operation: OperationLike,
				Field:     "field4",
				FieldCfg:  fieldsMap["field4"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "value1",
				},
			}
			cond, err := NewLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})
	})
}

func TestLikeCond_Convert(t *testing.T) {
	Convey("Test LikeCond.Convert", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*FieldCfg{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("should create regexp query", func() {
			cfg := &CondCfg{
				Operation: OperationLike,
				Field:     "field1",
				FieldCfg:  fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, _ := NewLikeCond(ctx, cfg, fieldsMap)
			likeCond := cond.(*LikeCond)

			vectorizer := func(ctx context.Context, words []string) ([]*VectorResp, error) {
				return nil, nil
			}

			dsl, err := likeCond.Convert(ctx, vectorizer)
			So(err, ShouldBeNil)
			So(dsl, ShouldNotBeEmpty)
			So(dsl, ShouldContainSubstring, "regexp")
			So(dsl, ShouldContainSubstring, "field1")
			So(dsl, ShouldContainSubstring, ".*test.*")
		})
	})
}

func TestLikeCond_Convert2SQL(t *testing.T) {
	Convey("Test LikeCond.Convert2SQL", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*FieldCfg{
			"field1": {
				Name: "field1",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("should create LIKE query", func() {
			cfg := &CondCfg{
				Operation: OperationLike,
				Field:     "field1",
				FieldCfg:  fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test",
				},
			}
			cond, _ := NewLikeCond(ctx, cfg, fieldsMap)
			likeCond := cond.(*LikeCond)

			sql, err := likeCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldNotBeEmpty)
			So(sql, ShouldContainSubstring, "field1")
			So(sql, ShouldContainSubstring, "LIKE")
			So(sql, ShouldContainSubstring, "test")
		})

		Convey("special characters should be escaped", func() {
			cfg := &CondCfg{
				Operation: OperationLike,
				Field:     "field1",
				FieldCfg:  fieldsMap["field1"],
				ValueOptCfg: ValueOptCfg{
					ValueFrom: ValueFrom_Const,
					Value:     "test'%_\\",
				},
			}
			cond, _ := NewLikeCond(ctx, cfg, fieldsMap)
			likeCond := cond.(*LikeCond)

			sql, err := likeCond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(sql, ShouldNotBeEmpty)
			So(sql, ShouldContainSubstring, "LIKE")
		})
	})
}
