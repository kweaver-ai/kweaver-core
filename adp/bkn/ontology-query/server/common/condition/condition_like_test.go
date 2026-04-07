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

func Test_NewLikeCond(t *testing.T) {
	Convey("Test NewLikeCond", t, func() {
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

		Convey("成功 - 字符串字段", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationLike,
				ValueOptCfg: ValueOptCfg{
					Value: "test%",
				},
			}
			cfg.NameField = fieldsMap["name"]
			cond, err := NewLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 字段不存在", func() {
			cfg := &CondCfg{
				Name:      "nonexistent",
				Operation: OperationLike,
				ValueOptCfg: ValueOptCfg{
					Value: "test%",
				},
			}
			cond, err := NewLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 非字符串字段", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLike,
				ValueOptCfg: ValueOptCfg{
					Value: "test%",
				},
			}
			cfg.NameField = fieldsMap["age"]
			cond, err := NewLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 非字符串值", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationLike,
				ValueOptCfg: ValueOptCfg{
					Value: 123,
				},
			}
			cfg.NameField = fieldsMap["name"]
			cond, err := NewLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_LikeCond_Convert(t *testing.T) {
	Convey("Test LikeCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &LikeCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mValue:           "test",
				mFilterFieldName: "name",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"regexp"`)
			So(result, ShouldContainSubstring, `"name"`)
		})
	})
}

func Test_LikeCond_Convert2SQL(t *testing.T) {
	Convey("Test LikeCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换SQL", func() {
			cond := &LikeCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mValue:           "test",
				mFilterFieldName: "name",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"name"`)
			So(result, ShouldContainSubstring, `LIKE`)
		})
	})
}

func Test_rewriteLikeCond(t *testing.T) {
	Convey("Test rewriteLikeCond", t, func() {
		Convey("成功 - 重写条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationLike,
				ValueOptCfg: ValueOptCfg{
					Value: "test%",
				},
				NameField: &DataProperty{
					Name: "name",
					MappedField: Field{
						Name: "mapped_name",
					},
				},
			}
			result, err := rewriteLikeCond(cfg)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_name")
		})

		Convey("失败 - NameField为空", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationLike,
				ValueOptCfg: ValueOptCfg{
					Value: "test%",
				},
				NameField: &DataProperty{
					Name: "",
				},
			}
			result, err := rewriteLikeCond(cfg)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func Test_NewNotLikeCond(t *testing.T) {
	Convey("Test NewNotLikeCond", t, func() {
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

		Convey("成功 - 创建NOT LIKE条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotLike,
				ValueOptCfg: ValueOptCfg{
					Value: "test%",
				},
			}
			cfg.NameField = fieldsMap["name"]
			cond, err := NewNotLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 非字符串值", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotLike,
				ValueOptCfg: ValueOptCfg{
					Value: 123,
				},
			}
			cfg.NameField = fieldsMap["name"]
			cond, err := NewNotLikeCond(ctx, cfg, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_NotLikeCond_Convert(t *testing.T) {
	Convey("Test NotLikeCond Convert", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换DSL", func() {
			cond := &NotLikeCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mValue:           "test",
				mFilterFieldName: "name",
			}
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"bool"`)
			So(result, ShouldContainSubstring, `"must_not"`)
			So(result, ShouldContainSubstring, `"regexp"`)
		})
	})
}

func Test_NotLikeCond_Convert2SQL(t *testing.T) {
	Convey("Test NotLikeCond Convert2SQL", t, func() {
		ctx := context.Background()

		Convey("成功 - 转换SQL", func() {
			cond := &NotLikeCond{
				mCfg: &CondCfg{
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				},
				mValue:           "test",
				mFilterFieldName: "name",
			}
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"name"`)
			So(result, ShouldContainSubstring, `NOT LIKE`)
		})
	})
}
