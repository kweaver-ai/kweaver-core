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

func Test_NewInCond(t *testing.T) {
	Convey("Test NewInCond", t, func() {
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

		Convey("成功 - 字符串数组", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationIn,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"test1", "test2", "test3"},
				},
			}
			cond, err := NewInCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 整数数组", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationIn,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, 19, 20},
				},
			}
			cond, err := NewInCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 非数组值", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationIn,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewInCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 不同类型元素", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationIn,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"test1", 123, "test3"},
				},
			}
			cond, err := NewInCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_InCond_Convert(t *testing.T) {
	Convey("Test InCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 字符串数组", func() {
			cond := &InCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: []any{"test1", "test2"},
					},
				},
				mValue:           []any{"test1", "test2"},
				mFilterFieldName: "name",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"terms"`)
			So(result, ShouldContainSubstring, `"name"`)
		})

		Convey("成功 - 整数数组", func() {
			cond := &InCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: []any{1, 2, 3},
					},
				},
				mValue:           []any{1, 2, 3},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"terms"`)
			So(result, ShouldContainSubstring, `"age"`)
		})
	})
}

func Test_InCond_Convert2SQL(t *testing.T) {
	Convey("Test InCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 字符串数组", func() {
			cond := &InCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: []any{"test1", "test2"},
					},
				},
				mValue:           []any{"test1", "test2"},
				mFilterFieldName: "name",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"name"`)
			So(result, ShouldContainSubstring, `IN`)
			So(result, ShouldContainSubstring, `'test1'`)
			So(result, ShouldContainSubstring, `'test2'`)
		})

		Convey("成功 - 整数数组", func() {
			cond := &InCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: []any{1, 2, 3},
					},
				},
				mValue:           []any{1, 2, 3},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"age"`)
			So(result, ShouldContainSubstring, `IN`)
			So(result, ShouldContainSubstring, `1`)
			So(result, ShouldContainSubstring, `2`)
			So(result, ShouldContainSubstring, `3`)
		})
	})
}

func Test_rewriteInCond(t *testing.T) {
	Convey("Test rewriteInCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationIn,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"test1", "test2"},
				},
				NameField: &DataProperty{
					Name: "name",
					MappedField: Field{
						Name: "mapped_name",
					},
				},
			}
			result, err := rewriteInCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_name")
			So(result.Operation, ShouldEqual, OperationIn)
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationIn,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"test1", "test2"},
				},
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteInCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}
