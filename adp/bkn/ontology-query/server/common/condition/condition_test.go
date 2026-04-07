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

func Test_NewCondition(t *testing.T) {
	Convey("Test NewCondition", t, func() {
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

		Convey("成功 - nil条件", func() {
			cond, err := NewCondition(ctx, nil, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("成功 - 等于条件", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewCondition(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - AND条件", func() {
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
			cond, err := NewCondition(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - OR条件", func() {
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
			cond, err := NewCondition(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 字段不存在", func() {
			cfg := &CondCfg{
				Name:      "nonexistent",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewCondition(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 二进制类型字段不支持过滤", func() {
			binaryFieldsMap := map[string]*DataProperty{
				"binary_field": {
					Name: "binary_field",
					Type: dtype.DATATYPE_BINARY,
				},
			}
			cfg := &CondCfg{
				Name:      "binary_field",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewCondition(ctx, cfg, CUSTOM, binaryFieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 不支持的操作", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: "unsupported_op",
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewCondition(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_NewCondWithOpr(t *testing.T) {
	Convey("Test NewCondWithOpr", t, func() {
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
			"text_field": {
				Name: "text_field",
				Type: dtype.DATATYPE_TEXT,
				MappedField: Field{
					Name: "mapped_text",
				},
			},
		}

		Convey("成功 - 等于操作", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 不等于操作", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 大于操作", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationGt,
				ValueOptCfg: ValueOptCfg{
					Value: 18,
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 大于等于操作", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationGte,
				ValueOptCfg: ValueOptCfg{
					Value: 18,
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 小于操作", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLt,
				ValueOptCfg: ValueOptCfg{
					Value: 65,
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - 小于等于操作", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationLte,
				ValueOptCfg: ValueOptCfg{
					Value: 65,
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - IN操作", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationIn,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"test1", "test2"},
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - NOT IN操作", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotIn,
				ValueOptCfg: ValueOptCfg{
					Value: []any{"test1", "test2"},
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - LIKE操作", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationLike,
				ValueOptCfg: ValueOptCfg{
					Value: "test%",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - NOT LIKE操作", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotLike,
				ValueOptCfg: ValueOptCfg{
					Value: "test%",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - RANGE操作", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationRange,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, 65},
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - OUT RANGE操作", func() {
			cfg := &CondCfg{
				Name:      "age",
				Operation: OperationOutRange,
				ValueOptCfg: ValueOptCfg{
					Value: []any{18, 65},
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - EXIST操作", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationExist,
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - NOT EXIST操作", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationNotExist,
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - REGEX操作", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationRegex,
				ValueOptCfg: ValueOptCfg{
					Value: "test.*",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - MATCH操作", func() {
			textFieldsMap := map[string]*DataProperty{
				"text_field": {
					Name: "text_field",
					Type: dtype.DATATYPE_TEXT,
					MappedField: Field{
						Name: "mapped_text",
					},
				},
			}
			cfg := &CondCfg{
				Name:      "text_field",
				Operation: OperationMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, textFieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - MATCH PHRASE操作", func() {
			textFieldsMap := map[string]*DataProperty{
				"text_field": {
					Name: "text_field",
					Type: dtype.DATATYPE_TEXT,
					MappedField: Field{
						Name: "mapped_text",
					},
					IndexConfig: &IndexConfig{
						FulltextConfig: FulltextConfig{
							Enabled: true,
						},
					},
				},
			}
			cfg := &CondCfg{
				Name:      "text_field",
				Operation: OperationMatchPhrase,
				ValueOptCfg: ValueOptCfg{
					Value: "test phrase",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, textFieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - MULTI MATCH操作", func() {
			cfg := &CondCfg{
				Name:      AllField,
				Operation: OperationMultiMatch,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("成功 - KNN操作", func() {
			vectorFieldsMap := map[string]*DataProperty{
				"vector_field": {
					Name: "vector_field",
					Type: dtype.DATATYPE_VECTOR,
					IndexConfig: &IndexConfig{
						VectorConfig: VectorConfig{
							Enabled: true,
							ModelID: "model1",
						},
					},
					MappedField: Field{
						Name: "mapped_vector",
					},
				},
			}
			cfg := &CondCfg{
				Name:      "vector_field",
				Operation: OperationKNN,
				ValueOptCfg: ValueOptCfg{
					Value: []float32{0.1, 0.2, 0.3},
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, vectorFieldsMap)
			So(err, ShouldBeNil)
			So(cond, ShouldNotBeNil)
		})

		Convey("失败 - 字段不存在", func() {
			cfg := &CondCfg{
				Name:      "nonexistent",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 二进制类型字段", func() {
			binaryFieldsMap := map[string]*DataProperty{
				"binary_field": {
					Name: "binary_field",
					Type: dtype.DATATYPE_BINARY,
				},
			}
			cfg := &CondCfg{
				Name:      "binary_field",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, binaryFieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})

		Convey("失败 - 不支持的操作", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: "unsupported",
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cond, err := NewCondWithOpr(ctx, cfg, CUSTOM, fieldsMap)
			So(err, ShouldNotBeNil)
			So(cond, ShouldBeNil)
		})
	})
}

func Test_getFilterFieldName(t *testing.T) {
	Convey("Test getFilterFieldName", t, func() {
		fieldsMap := map[string]*DataProperty{
			"name": {
				Name: "name",
				Type: dtype.DATATYPE_STRING,
				IndexConfig: &IndexConfig{
					KeywordConfig: KeywordConfig{
						Enabled: true,
					},
				},
			},
			"text_field": {
				Name: "text_field",
				Type: dtype.DATATYPE_TEXT,
			},
			"name_desensitize": {
				Name: "name_desensitize",
				Type: dtype.DATATYPE_STRING,
			},
		}

		Convey("成功 - 普通字段", func() {
			result := getFilterFieldName("name", fieldsMap, false)
			So(result, ShouldEqual, "name_desensitize")
		})

		Convey("成功 - 全文检索字段", func() {
			result := getFilterFieldName("text_field", fieldsMap, true)
			So(result, ShouldEqual, "text_field")
		})

		Convey("成功 - 全文检索时text字段不加keyword", func() {
			result := getFilterFieldName("text_field", fieldsMap, true)
			So(result, ShouldEqual, "text_field")
		})

		Convey("成功 - 精确查询时text字段加keyword", func() {
			textFieldsMap := map[string]*DataProperty{
				"text_field": {
					Name: "text_field",
					Type: dtype.DATATYPE_TEXT,
					IndexConfig: &IndexConfig{
						KeywordConfig: KeywordConfig{
							Enabled: true,
						},
					},
				},
			}
			result := getFilterFieldName("text_field", textFieldsMap, false)
			So(result, ShouldEqual, "text_field.keyword")
		})

		Convey("成功 - AllField", func() {
			result := getFilterFieldName(AllField, fieldsMap, true)
			So(result, ShouldEqual, AllField)
		})

		Convey("成功 - MetaField_ID", func() {
			result := getFilterFieldName(MetaField_ID, fieldsMap, false)
			So(result, ShouldEqual, OS_MetaField_ID)
		})

		Convey("成功 - 脱敏字段", func() {
			result := getFilterFieldName("name", fieldsMap, false)
			// 由于存在 name_desensitize，应该返回脱敏字段名
			So(result, ShouldContainSubstring, "name")
		})
	})
}

func Test_wrapKeyWordFieldName(t *testing.T) {
	Convey("Test wrapKeyWordFieldName", t, func() {
		Convey("成功 - 单个字段", func() {
			result := wrapKeyWordFieldName("name")
			So(result, ShouldEqual, "name.keyword")
		})

		Convey("成功 - 多个字段", func() {
			result := wrapKeyWordFieldName("field1", "field2")
			So(result, ShouldEqual, "field1.field2.keyword")
		})

		Convey("成功 - 空字段", func() {
			result := wrapKeyWordFieldName("")
			So(result, ShouldEqual, "")
		})
	})
}

func Test_RewriteCondition(t *testing.T) {
	Convey("Test RewriteCondition", t, func() {
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

		Convey("成功 - nil条件", func() {
			result, err := RewriteCondition(ctx, nil, fieldsMap, vectorizer)
			So(err, ShouldBeNil)
			So(result, ShouldBeNil)
		})

		Convey("成功 - AND条件重写", func() {
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
			// 需要先设置 NameField
			for _, subCond := range cfg.SubConds {
				if field, ok := fieldsMap[subCond.Name]; ok {
					subCond.NameField = field
				}
			}
			result, err := RewriteCondition(ctx, cfg, fieldsMap, vectorizer)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.SubConds), ShouldEqual, 2)
		})

		Convey("成功 - OR条件重写", func() {
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
			// 需要先设置 NameField
			for _, subCond := range cfg.SubConds {
				if field, ok := fieldsMap[subCond.Name]; ok {
					subCond.NameField = field
				}
			}
			result, err := RewriteCondition(ctx, cfg, fieldsMap, vectorizer)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(len(result.SubConds), ShouldEqual, 2)
		})

		Convey("成功 - 等于条件重写", func() {
			cfg := &CondCfg{
				Name:      "name",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			cfg.NameField = fieldsMap["name"]
			result, err := RewriteCondition(ctx, cfg, fieldsMap, vectorizer)
			So(err, ShouldBeNil)
			So(result, ShouldNotBeNil)
			So(result.Name, ShouldEqual, "mapped_name")
		})

		Convey("失败 - 字段不存在", func() {
			cfg := &CondCfg{
				Name:      "nonexistent",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			result, err := RewriteCondition(ctx, cfg, fieldsMap, vectorizer)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})

		Convey("失败 - 二进制类型字段", func() {
			binaryFieldsMap := map[string]*DataProperty{
				"binary_field": {
					Name: "binary_field",
					Type: dtype.DATATYPE_BINARY,
				},
			}
			cfg := &CondCfg{
				Name:      "binary_field",
				Operation: OperationEq,
				ValueOptCfg: ValueOptCfg{
					Value: "test",
				},
			}
			result, err := RewriteCondition(ctx, cfg, binaryFieldsMap, vectorizer)
			So(err, ShouldNotBeNil)
			So(result, ShouldBeNil)
		})
	})
}

func Test_IsSlice(t *testing.T) {
	Convey("Test IsSlice", t, func() {
		Convey("成功 - 切片", func() {
			result := IsSlice([]string{"a", "b"})
			So(result, ShouldBeTrue)
		})

		Convey("成功 - 数组", func() {
			result := IsSlice([2]int{1, 2})
			So(result, ShouldBeTrue)
		})

		Convey("失败 - 字符串", func() {
			result := IsSlice("test")
			So(result, ShouldBeFalse)
		})

		Convey("失败 - 整数", func() {
			result := IsSlice(123)
			So(result, ShouldBeFalse)
		})

		Convey("失败 - map", func() {
			result := IsSlice(map[string]int{"a": 1})
			So(result, ShouldBeFalse)
		})
	})
}

func Test_IsSameType(t *testing.T) {
	Convey("Test IsSameType", t, func() {
		Convey("成功 - 空数组", func() {
			result := IsSameType([]any{})
			So(result, ShouldBeTrue)
		})

		Convey("成功 - 相同类型", func() {
			result := IsSameType([]any{1, 2, 3})
			So(result, ShouldBeTrue)
		})

		Convey("成功 - 相同字符串类型", func() {
			result := IsSameType([]any{"a", "b", "c"})
			So(result, ShouldBeTrue)
		})

		Convey("失败 - 不同类型", func() {
			result := IsSameType([]any{1, "2", 3})
			So(result, ShouldBeFalse)
		})

		Convey("失败 - 混合类型", func() {
			result := IsSameType([]any{1, 2.0, "3"})
			So(result, ShouldBeFalse)
		})
	})
}
