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

func Test_NewNotInCond(t *testing.T) {
	Convey("Test NewNotInCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"name": {
				Name: "name",
				Type: dtype.DATATYPE_STRING,
				MappedField: Field{
					Name: "mapped_name",
				},
			},
		}

		Convey("成功 - 字符串数组", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotIn,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"test1", "test2"},
				},
			}
			cond, err := NewNotInCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 非数组值", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotIn,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewNotInCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 不同类型元素", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotIn,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"test1", 123},
				},
			}
			cond, err := NewNotInCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_NotInCond_Convert(t *testing.T) {
	Convey("Test NotInCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &NotInCond{
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
			So(result, ShouldContainSubstring, `"bool"`)
			So(result, ShouldContainSubstring, `"must_not"`)
			So(result, ShouldContainSubstring, `"terms"`)
		})
	})
}

func Test_NotInCond_Convert2SQL(t *testing.T) {
	Convey("Test NotInCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换SQL", func() {
			cond := &NotInCond{
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
			So(result, ShouldContainSubstring, `NOT IN`)
		})
	})
}

func Test_rewriteNotInCond(t *testing.T) {
	Convey("Test rewriteNotInCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotIn,
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
			result, err := rewriteNotInCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_name")
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotIn,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"test1", "test2"},
				},
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteNotInCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}
