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

func Test_newAndCond(t *testing.T) {
	Convey("Test newAndCond", t, func() {
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

		Convey("成功 - 两个子条件", func() {
			cfg := &CondCfg{
				Operation: OperationAnd,
				SubConds: []*CondCfg{
					{
						Name:      "name",
						Operation: OperationEq,
						ValueOptCfg: ValueOptCfg{
							Value: "test",
						},
					},
					{
						Name:      "age",
						Operation: OperationGt,
						ValueOptCfg: ValueOptCfg{
							Value: 18,
						},
					},
				},
			}
			cond, err := newAndCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 子条件为空", func() {
			cfg := &CondCfg{
				Operation: OperationAnd,
				SubConds:  []*CondCfg{},
			}
			cond, err := newAndCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 子条件超过限制", func() {
			subConds := make([]*CondCfg, MaxSubCondition+1)
			for i := 0; i <= MaxSubCondition; i++ {
				subConds[i] = &CondCfg{
					Name:      "name",
					Operation: OperationEq,
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				}
			}
			cfg := &CondCfg{
				Operation: OperationAnd,
				SubConds:  subConds,
			}
			cond, err := newAndCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 子条件错误", func() {
			cfg := &CondCfg{
				Operation: OperationAnd,
				SubConds: []*CondCfg{
					{
						Name:      "nonexistent",
						Operation: OperationEq,
						ValueOptCfg: ValueOptCfg{
							Value: "test",
						},
					},
				},
			}
			cond, err := newAndCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_AndCond_Convert(t *testing.T) {
	Convey("Test AndCond Convert", t, func() {
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

		Convey("成功 - 转换DSL", func() {
			cfg := &CondCfg{
				Operation: OperationAnd,
				SubConds: []*CondCfg{
					{
						Name:      "name",
						Operation: OperationEq,
						ValueOptCfg: ValueOptCfg{
							Value: "test",
						},
					},
					{
						Name:      "age",
						Operation: OperationGt,
						ValueOptCfg: ValueOptCfg{
							Value: 18,
						},
					},
				},
			}
			cond, err := newAndCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"bool"`)
			So(result, ShouldContainSubstring, `"must"`)
		})
	})
}

func Test_AndCond_Convert2SQL(t *testing.T) {
	Convey("Test AndCond Convert2SQL", t, func() {
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

		Convey("成功 - 转换SQL", func() {
			cfg := &CondCfg{
				Operation: OperationAnd,
				SubConds: []*CondCfg{
					{
						Name:      "name",
						Operation: OperationEq,
						ValueOptCfg: ValueOptCfg{
							Value: "test",
						},
					},
					{
						Name:      "age",
						Operation: OperationGt,
						ValueOptCfg: ValueOptCfg{
							Value: 18,
						},
					},
				},
			}
			cond, err := newAndCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `AND`)
		})
	})
}

func Test_rewriteAndCondition(t *testing.T) {
	Convey("Test rewriteAndCondition", t, func() {
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
		vectorizer := func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error) {
			return []VectorResp{}, nil
		}

		Convey("成功 - 重写AND条件", func() {
			cfg := &CondCfg{
				Operation: OperationAnd,
				SubConds: []*CondCfg{
					{
						Name:      "name",
						Operation: OperationEq,
						ValueOptCfg: ValueOptCfg{
							Value: "test",
						},
					},
					{
						Name:      "age",
						Operation: OperationGt,
						ValueOptCfg: ValueOptCfg{
							Value: 18,
						},
					},
				},
			}
			// 设置 NameField
			for _, subCond := range cfg.SubConds {
				if field, ok := fieldsMap[subCond.Name]; ok {
					subCond.NameField = field
				}
			}
			result, err := rewriteAndCondition(ctx, cfg, fieldsMap, vectorizer)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.SubConds), ShouldEqual, 2)
		})

		Convey("失败 - 子条件为空", func() {
			cfg := &CondCfg{
				Operation: OperationAnd,
				SubConds:  []*CondCfg{},
			}
			result, err := rewriteAndCondition(ctx, cfg, fieldsMap, vectorizer)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("失败 - 子条件超过限制", func() {
			subConds := make([]*CondCfg, MaxSubCondition+1)
			for i := 0; i <= MaxSubCondition; i++ {
				subConds[i] = &CondCfg{
					Name:      "name",
					Operation: OperationEq,
					ValueOptCfg: ValueOptCfg{
						Value: "test",
					},
				}
				subConds[i].NameField = fieldsMap["name"]
			}
			cfg := &CondCfg{
				Operation: OperationAnd,
				SubConds:  subConds,
			}
			result, err := rewriteAndCondition(ctx, cfg, fieldsMap, vectorizer)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func Test_newOrCond(t *testing.T) {
	Convey("Test newOrCond", t, func() {
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

		Convey("成功 - 两个子条件", func() {
			cfg := &CondCfg{
				Operation: OperationOr,
				SubConds: []*CondCfg{
					{
						Name:      "name",
						Operation: OperationEq,
						ValueOptCfg: ValueOptCfg{
							Value: "test1",
						},
					},
					{
						Name:      "name",
						Operation: OperationEq,
						ValueOptCfg: ValueOptCfg{
							Value: "test2",
						},
					},
				},
			}
			cond, err := newOrCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 子条件为空", func() {
			cfg := &CondCfg{
				Operation: OperationOr,
				SubConds:  []*CondCfg{},
			}
			cond, err := newOrCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_OrCond_Convert(t *testing.T) {
	Convey("Test OrCond Convert", t, func() {
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

		Convey("成功 - 转换DSL", func() {
			cfg := &CondCfg{
				Operation: OperationOr,
				SubConds: []*CondCfg{
					{
						Name:      "name",
						Operation: OperationEq,
						ValueOptCfg: ValueOptCfg{
							Value: "test1",
						},
					},
					{
						Name:      "name",
						Operation: OperationEq,
						ValueOptCfg: ValueOptCfg{
							Value: "test2",
						},
					},
				},
			}
			cond, err := newOrCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			result, err := cond.Convert(ctx, nil)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `"bool"`)
			So(result, ShouldContainSubstring, `"should"`)
		})
	})
}

func Test_OrCond_Convert2SQL(t *testing.T) {
	Convey("Test OrCond Convert2SQL", t, func() {
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

		Convey("成功 - 转换SQL", func() {
			cfg := &CondCfg{
				Operation: OperationOr,
				SubConds: []*CondCfg{
					{
						Name:      "name",
						Operation: OperationEq,
						ValueOptCfg: ValueOptCfg{
							Value: "test1",
						},
					},
					{
						Name:      "name",
						Operation: OperationEq,
						ValueOptCfg: ValueOptCfg{
							Value: "test2",
						},
					},
				},
			}
			cond, err := newOrCond(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			result, err := cond.Convert2SQL(ctx)
			So(err, ShouldBeNil)
			So(result, ShouldContainSubstring, `OR`)
		})
	})
}
