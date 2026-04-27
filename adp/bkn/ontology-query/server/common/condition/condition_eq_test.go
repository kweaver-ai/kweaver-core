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

func Test_NewEqCond(t *testing.T) {
	Convey("Test NewEqCond", t, func() {
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

		Convey("成功 - 字符串值", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 整数值", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: 18,
				},
			}
			cond, err := NewEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 数组值", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"test1", "test2"},
				},
			}
			cond, err := NewEqCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_EqCond_Convert(t *testing.T) {
	Convey("Test EqCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 字符串值", func() {
			cond := &EqCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mFilterFieldName: "name",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"term"`)
			So(result, ShouldContainSubstring, `"name"`)
			So(result, ShouldContainSubstring, `"test"`)
		})

		Convey("成功 - 整数值", func() {
			cond := &EqCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: 123,
					},
				},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"term"`)
			So(result, ShouldContainSubstring, `"age"`)
			So(result, ShouldContainSubstring, `123`)
		})
	})
}

func Test_EqCond_Convert2SQL(t *testing.T) {
	Convey("Test EqCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 字符串值", func() {
			cond := &EqCond{
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
			So(result, ShouldContainSubstring, `'test'`)
		})

		Convey("成功 - 整数值", func() {
			cond := &EqCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: 123,
					},
				},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"age"`)
			So(result, ShouldContainSubstring, `123`)
		})
	})
}

func Test_rewriteEqCond(t *testing.T) {
	Convey("Test rewriteEqCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationEq,
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
			result, err := rewriteEqCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_name")
			So(result.Operation, ShouldEqual, OperationEq)
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteEqCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("失败 - NameField为nil", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
				NameField: nil,
			}
			result, err := rewriteEqCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}
