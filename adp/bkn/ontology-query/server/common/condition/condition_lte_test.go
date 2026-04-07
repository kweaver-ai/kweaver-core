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

func Test_NewLteCond(t *testing.T) {
	Convey("Test NewLteCond", t, func() {
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

		Convey("成功 - 创建小于等于条件（整数）", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLte,
				ValueOptCfg: ValueOptCfg{
					Value: 65,
				},
			}
			cond, err := NewLteCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 创建小于等于条件（字符串）", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationLte,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewLteCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 数组值", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLte,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, 19},
				},
			}
			cond, err := NewLteCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_LteCond_Convert(t *testing.T) {
	Convey("Test LteCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 整数转换DSL", func() {
			cond := &LteCond{
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
			So(result, ShouldContainSubstring, `"lte"`)
			So(result, ShouldContainSubstring, `65`)
		})

		Convey("成功 - 字符串转换DSL", func() {
			cond := &LteCond{
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
			So(result, ShouldContainSubstring, `"lte"`)
			So(result, ShouldContainSubstring, `"test"`)
		})
	})
}

func Test_LteCond_Convert2SQL(t *testing.T) {
	Convey("Test LteCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 整数转换SQL", func() {
			cond := &LteCond{
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
			So(result, ShouldContainSubstring, `<=`)
			So(result, ShouldContainSubstring, `65`)
		})

		Convey("成功 - 字符串转换SQL", func() {
			cond := &LteCond{
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
			So(result, ShouldContainSubstring, `<=`)
			So(result, ShouldContainSubstring, `'test'`)
		})
	})
}

func Test_rewriteLteCond(t *testing.T) {
	Convey("Test rewriteLteCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLte,
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
			result, err := rewriteLteCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_age")
			So(result.Operation, ShouldEqual, OperationLte)
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLte,
				ValueOptCfg: ValueOptCfg{
					Value: 65,
				},
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteLteCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

	})
}
