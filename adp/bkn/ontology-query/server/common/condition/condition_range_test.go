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

func Test_NewRangeCond(t *testing.T) {
	Convey("Test NewRangeCond", t, func() {
		ctx := context.Background()
		fieldsMap := map[string]*DataProperty{
			"age": {
				Name: "age",
				Type: dtype.DATATYPE_INTEGER,
				MappedField: Field{
					Name: "mapped_age",
				},
			},
			"date_field": {
				Name: "date_field",
				Type: dtype.DATATYPE_DATETIME,
				MappedField: Field{
					Name: "mapped_date",
				},
			},
		}

		Convey("成功 - 整数范围", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationRange,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, 65},
				},
			}
			cond, err := NewRangeCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 字符串范围", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationRange,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"a", "z"},
				},
			}
			cond, err := NewRangeCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 非数组值", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationRange,
				ValueOptCfg: ValueOptCfg{
					Value: 18,
				},
			}
			cond, err := NewRangeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 数组长度不等于2", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationRange,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18},
				},
			}
			cond, err := NewRangeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 不同类型元素", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationRange,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, "65"},
				},
			}
			cond, err := NewRangeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_RangeCond_Convert(t *testing.T) {
	Convey("Test RangeCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 整数范围", func() {
			cond := &RangeCond{
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
			So(result, ShouldContainSubstring, `"range"`)
			So(result, ShouldContainSubstring, `"age"`)
			So(result, ShouldContainSubstring, `"gte"`)
			So(result, ShouldContainSubstring, `"lt"`)
		})

		Convey("成功 - 字符串范围", func() {
			cond := &RangeCond{
				mCfg: &CondCfg{
					NameField: &DataProperty{
						Type: dtype.DATATYPE_STRING,
					},
					ValueOptCfg: ValueOptCfg{
						Value: []any{"a", "z"},
					},
				},
				mValue:           []any{"a", "z"},
				mFilterFieldName: "name",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"range"`)
			So(result, ShouldContainSubstring, `"name"`)
		})

		Convey("成功 - 日期时间范围", func() {
			cond := &RangeCond{
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
			So(result, ShouldContainSubstring, `"range"`)
			So(result, ShouldContainSubstring, `"date_field"`)
			So(result, ShouldContainSubstring, `"format"`)
		})

		Convey("成功 - 日期时间范围（时间戳）", func() {
			cond := &RangeCond{
				mCfg: &CondCfg{
					NameField: &DataProperty{
						Type: dtype.DATATYPE_DATETIME,
					},
					ValueOptCfg: ValueOptCfg{
						Value: []any{1672531200000.0, 1704067199000.0},
					},
				},
				mValue:           []any{1672531200000.0, 1704067199000.0},
				mFilterFieldName: "date_field",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"range"`)
			So(result, ShouldContainSubstring, `"format"`)
			So(result, ShouldContainSubstring, `epoch_millis`)
		})
	})
}

func Test_RangeCond_Convert2SQL(t *testing.T) {
	Convey("Test RangeCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 整数范围", func() {
			cond := &RangeCond{
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
			So(result, ShouldContainSubstring, `>=`)
			So(result, ShouldContainSubstring, `<`)
			So(result, ShouldContainSubstring, `AND`)
		})

		Convey("成功 - 字符串范围", func() {
			cond := &RangeCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: []any{"a", "z"},
					},
				},
				mValue:           []any{"a", "z"},
				mFilterFieldName: "name",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"name"`)
			So(result, ShouldContainSubstring, `>=`)
			So(result, ShouldContainSubstring, `<`)
		})
	})
}

func Test_rewriteRangeCond(t *testing.T) {
	Convey("Test rewriteRangeCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationRange,
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
			result, err := rewriteRangeCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_age")
			So(result.Operation, ShouldEqual, OperationRange)
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationRange,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, 65},
				},
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteRangeCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}
