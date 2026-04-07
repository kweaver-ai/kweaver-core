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

func Test_NewOutRangeCond(t *testing.T) {
	Convey("Test NewOutRangeCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"age": {
				Name: "age",
				Type: dtype.DATATYPE_INTEGER,
				MappedField: Field{
					Name: "mapped_age",
				},
			},
		}

		Convey("成功 - 整数范围外", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationOutRange,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, 65},
				},
			}
			cond, err := NewOutRangeCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 非数组值", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationOutRange,
				ValueOptCfg: ValueOptCfg{
					Value: 18,
				},
			}
			cond, err := NewOutRangeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 数组长度不等于2", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationOutRange,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18},
				},
			}
			cond, err := NewOutRangeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 不同类型元素", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationOutRange,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, "65"},
				},
			}
			cond, err := NewOutRangeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_OutRangeCond_Convert(t *testing.T) {
	Convey("Test OutRangeCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 整数范围外", func() {
			cond := &OutRangeCond{
				mCfg: &CondCfg{
					NameField: &DataProperty{
						Type: dtype.DATATYPE_INTEGER,
					},
					ValueOptCfg: ValueOptCfg{
						Value: []any{18, 65},
					},
				},
				mValue:           []any{18, 65},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"bool"`)
			So(result, ShouldContainSubstring, `"should"`)
			So(result, ShouldContainSubstring, `"range"`)
			So(result, ShouldContainSubstring, `"lt"`)
			So(result, ShouldContainSubstring, `"gte"`)
		})

		Convey("成功 - 日期时间范围外", func() {
			cond := &OutRangeCond{
				mCfg: &CondCfg{
					NameField: &DataProperty{
						Type: dtype.DATATYPE_DATETIME,
					},
					ValueOptCfg: ValueOptCfg{
						Value: []any{"2023-01-01", "2023-12-31"},
					},
				},
				mValue:           []any{"2023-01-01", "2023-12-31"},
				mFilterFieldName: "date_field",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"format"`)
		})
	})
}

func Test_OutRangeCond_Convert2SQL(t *testing.T) {
	Convey("Test OutRangeCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换SQL", func() {
			cond := &OutRangeCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: []any{18, 65},
					},
				},
				mValue:           []any{18, 65},
				mFilterFieldName: "age",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"age"`)
			So(result, ShouldContainSubstring, `<`)
			So(result, ShouldContainSubstring, `>=`)
			So(result, ShouldContainSubstring, `OR`)
		})
	})
}

func Test_rewriteOutRangeCond(t *testing.T) {
	Convey("Test rewriteOutRangeCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationOutRange,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, 65},
				},
				NameField: &DataProperty{
					Name: "age",
					MappedField: Field{
						Name: "mapped_age",
					},
				},
			}
			result, err := rewriteOutRangeCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_age")
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationOutRange,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, 65},
				},
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteOutRangeCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}
