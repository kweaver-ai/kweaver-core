// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package condition

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_NewExistCond(t *testing.T) {
	Convey("Test NewExistCond", t, func() {
		Convey("成功 - 创建EXIST条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationExist,
			}
			cond, err := NewExistCond(cfg)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})
	})
}

func Test_ExistCond_Convert(t *testing.T) {
	Convey("Test ExistCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &ExistCond{
				mCfg: &CondCfg{
					Name: "name",
				},
				mfieldName: "name",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"exists"`)
			So(result, ShouldContainSubstring, `"field"`)
			So(result, ShouldContainSubstring, `"name"`)
		})
	})
}

func Test_ExistCond_Convert2SQL(t *testing.T) {
	Convey("Test ExistCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换SQL", func() {
			cond := &ExistCond{
				mCfg: &CondCfg{
					Name: "name",
				},
				mfieldName: "name",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"name"`)
			So(result, ShouldContainSubstring, `IS NOT NULL`)
		})
	})
}

func Test_rewriteExistCond(t *testing.T) {
	Convey("Test rewriteExistCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationExist,
				NameField: &DataProperty{
					Name: "name",
					MappedField: Field{
						Name: "mapped_name",
					},
				},
			}
			result, err := rewriteExistCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_name")
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationExist,
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteExistCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func Test_NewNotExistCond(t *testing.T) {
	Convey("Test NewNotExistCond", t, func() {
		Convey("成功 - 创建NOT EXIST条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotExist,
			}
			cond, err := NewNotExistCond(cfg)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})
	})
}

func Test_NotExistCond_Convert(t *testing.T) {
	Convey("Test NotExistCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &NotExistCond{
				mCfg: &CondCfg{
					Name: "name",
				},
				mfieldName: "name",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"bool"`)
			So(result, ShouldContainSubstring, `"must_not"`)
			So(result, ShouldContainSubstring, `"exists"`)
		})
	})
}

func Test_NotExistCond_Convert2SQL(t *testing.T) {
	Convey("Test NotExistCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换SQL", func() {
			cond := &NotExistCond{
				mCfg: &CondCfg{
					Name: "name",
				},
				mfieldName: "name",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"name"`)
			So(result, ShouldContainSubstring, `IS NULL`)
		})
	})
}
