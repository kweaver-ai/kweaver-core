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

func Test_NewRegexCond(t *testing.T) {
	Convey("Test NewRegexCond", t, func() {
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

		Convey("成功 - 创建正则条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationRegex,
				ValueOptCfg: ValueOptCfg{
					Value: "test.*",
				},
			}
			cfg.NameField = fieldsMap["name"]
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 字段不存在", func() {
			cfg := &CondCfg{
				Name:      "nonexistent",
				Operation: OperationRegex,
				ValueOptCfg: ValueOptCfg{
					Value: "test.*",
				},
			}
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 非字符串字段", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationRegex,
				ValueOptCfg: ValueOptCfg{
					Value: "test.*",
				},
			}
			cfg.NameField = fieldsMap["age"]
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 非字符串值", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationRegex,
				ValueOptCfg: ValueOptCfg{
					Value: 123,
				},
			}
			cfg.NameField = fieldsMap["name"]
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 无效的正则表达式", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationRegex,
				ValueOptCfg: ValueOptCfg{
					Value: "[invalid",
				},
			}
			cfg.NameField = fieldsMap["name"]
			cond, err := NewRegexCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_RegexCond_Convert(t *testing.T) {
	Convey("Test RegexCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &RegexCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test.*",
					},
				},
				mValue:           "test.*",
				mFilterFieldName: "name",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"regexp"`)
			So(result, ShouldContainSubstring, `"name"`)
		})
	})
}

func Test_RegexCond_Convert2SQL(t *testing.T) {
	Convey("Test RegexCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换SQL（返回空）", func() {
			cond := &RegexCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test.*",
					},
				},
				mValue:           "test.*",
				mFilterFieldName: "name",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldEqual, "")
		})
	})
}

func Test_rewriteRegexCond(t *testing.T) {
	Convey("Test rewriteRegexCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationRegex,
				ValueOptCfg: ValueOptCfg{
					Value: "test.*",
				},
				NameField: &DataProperty{
					Name: "name",
					MappedField: Field{
						Name: "mapped_name",
					},
				},
			}
			result, err := rewriteRegexCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_name")
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationRegex,
				ValueOptCfg: ValueOptCfg{
					Value: "test.*",
				},
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteRegexCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}
