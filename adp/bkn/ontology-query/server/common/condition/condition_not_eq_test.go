// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"testing"

	dtype "ontology-query/interfaces/data_type"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_NewNotEqCond(t *testing.T) {
	Convey("Test NewNotEqCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"name": {
				Name: "name",
				Type: dtype.DATATYPE_STRING,
				MappedField: Field{
					Name: "mapped_name",
				},
			},
			"age": {
				Name: "age",
				Type: dtype.DATATYPE_INTEGER,
				MappedField: Field{
					Name: "mapped_age",
				},
			},
		}

		Convey("成功 - 创建不等于条件（字符串）", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewNotEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 创建不等于条件（整数）", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationNotEq,
				ValueOptCfg: ValueOptCfg{
					Value: 18,
				},
			}
			cond, err := NewNotEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 数组值", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotEq,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"test1", "test2"},
				},
			}
			cond, err := NewNotEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_NotEqCond_Convert(t *testing.T) {
	Convey("Test NotEqCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 字符串转换DSL", func() {
			cond := &NotEqCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mFilterFieldName: "name",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"bool"`)
			So(result, ShouldContainSubstring, `"must_not"`)
			So(result, ShouldContainSubstring, `"term"`)
			So(result, ShouldContainSubstring, `"name"`)
			So(result, ShouldContainSubstring, `"test"`)
		})

		Convey("成功 - 整数转换DSL", func() {
			cond := &NotEqCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: 18,
					},
				},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"bool"`)
			So(result, ShouldContainSubstring, `"must_not"`)
			So(result, ShouldContainSubstring, `"term"`)
			So(result, ShouldContainSubstring, `"age"`)
			So(result, ShouldContainSubstring, `18`)
		})
	})
}

func Test_NotEqCond_Convert2SQL(t *testing.T) {
	Convey("Test NotEqCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 字符串转换SQL", func() {
			cond := &NotEqCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mFilterFieldName: "name",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"name"`)
			So(result, ShouldContainSubstring, `<>`)
			So(result, ShouldContainSubstring, `'test'`)
		})

		Convey("成功 - 整数转换SQL", func() {
			cond := &NotEqCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: 18,
					},
				},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"age"`)
			So(result, ShouldContainSubstring, `<>`)
			So(result, ShouldContainSubstring, `18`)
		})
	})
}

func Test_rewriteNotEqCond(t *testing.T) {
	Convey("Test rewriteNotEqCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				NameField: &DataProperty{
					Name: "name",
					MappedField: Field{
						Name: "mapped_name",
					},
				},
			}
			result, err := rewriteNotEqCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_name")
			So(result.Operation, ShouldEqual, OperationNotEq)
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteNotEqCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

	})
}
