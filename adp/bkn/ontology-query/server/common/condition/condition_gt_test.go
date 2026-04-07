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

func Test_NewGtCond(t *testing.T) {
	Convey("Test NewGtCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"age": {
				Name: "age",
				Type: dtype.DATATYPE_INTEGER,
				MappedField: Field{
					Name: "mapped_age",
				},
			},
			"name": {
				Name: "name",
				Type: dtype.DATATYPE_STRING,
				MappedField: Field{
					Name: "mapped_name",
				},
			},
		}

		Convey("成功 - 创建大于条件（整数）", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationGt,
				ValueOptCfg: ValueOptCfg{
					Value: 18,
				},
			}
			cond, err := NewGtCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 创建大于条件（字符串）", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationGt,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewGtCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 数组值", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationGt,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, 19},
				},
			}
			cond, err := NewGtCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_GtCond_Convert(t *testing.T) {
	Convey("Test GtCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 整数转换DSL", func() {
			cond := &GtCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: 18,
					},
				},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"range"`)
			So(result, ShouldContainSubstring, `"age"`)
			So(result, ShouldContainSubstring, `"gt"`)
			So(result, ShouldContainSubstring, `18`)
		})

		Convey("成功 - 字符串转换DSL", func() {
			cond := &GtCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mFilterFieldName: "name",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"range"`)
			So(result, ShouldContainSubstring, `"name"`)
			So(result, ShouldContainSubstring, `"gt"`)
			So(result, ShouldContainSubstring, `"test"`)
		})
	})
}

func Test_GtCond_Convert2SQL(t *testing.T) {
	Convey("Test GtCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 整数转换SQL", func() {
			cond := &GtCond{
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
			So(result, ShouldContainSubstring, `>`)
			So(result, ShouldContainSubstring, `18`)
		})

		Convey("成功 - 字符串转换SQL", func() {
			cond := &GtCond{
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
			So(result, ShouldContainSubstring, `>`)
			So(result, ShouldContainSubstring, `'test'`)
		})
	})
}

func Test_rewriteGtCond(t *testing.T) {
	Convey("Test rewriteGtCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationGt,
				ValueOptCfg: ValueOptCfg{
					Value: 18,
				},
				NameField: &DataProperty{
					Name: "age",
					MappedField: Field{
						Name: "mapped_age",
					},
				},
			}
			result, err := rewriteGtCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_age")
			So(result.Operation, ShouldEqual, OperationGt)
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationGt,
				ValueOptCfg: ValueOptCfg{
					Value: 18,
				},
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteGtCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

	})
}
