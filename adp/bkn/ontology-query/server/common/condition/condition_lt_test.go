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

func Test_NewLtCond(t *testing.T) {
	Convey("Test NewLtCond", t, func() {
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

		Convey("成功 - 创建小于条件（整数）", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLt,
				ValueOptCfg: ValueOptCfg{
					Value: 65,
				},
			}
			cond, err := NewLtCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 创建小于条件（字符串）", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationLt,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewLtCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 数组值", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLt,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, 19},
				},
			}
			cond, err := NewLtCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_LtCond_Convert(t *testing.T) {
	Convey("Test LtCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 整数转换DSL", func() {
			cond := &LtCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: 65,
					},
				},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"range"`)
			So(result, ShouldContainSubstring, `"age"`)
			So(result, ShouldContainSubstring, `"lt"`)
			So(result, ShouldContainSubstring, `65`)
		})

		Convey("成功 - 字符串转换DSL", func() {
			cond := &LtCond{
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
			So(result, ShouldContainSubstring, `"lt"`)
			So(result, ShouldContainSubstring, `"test"`)
		})
	})
}

func Test_LtCond_Convert2SQL(t *testing.T) {
	Convey("Test LtCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 整数转换SQL", func() {
			cond := &LtCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: 65,
					},
				},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"age"`)
			So(result, ShouldContainSubstring, `<`)
			So(result, ShouldContainSubstring, `65`)
		})

		Convey("成功 - 字符串转换SQL", func() {
			cond := &LtCond{
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
			So(result, ShouldContainSubstring, `<`)
			So(result, ShouldContainSubstring, `'test'`)
		})
	})
}

func Test_rewriteLtCond(t *testing.T) {
	Convey("Test rewriteLtCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLt,
				ValueOptCfg: ValueOptCfg{
					Value: 65,
				},
				NameField: &DataProperty{
					Name: "age",
					MappedField: Field{
						Name: "mapped_age",
					},
				},
			}
			result, err := rewriteLtCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_age")
			So(result.Operation, ShouldEqual, OperationLt)
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLt,
				ValueOptCfg: ValueOptCfg{
					Value: 65,
				},
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteLtCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

	})
}
