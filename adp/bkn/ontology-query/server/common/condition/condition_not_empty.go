package condition

import (
	"context"
	"fmt"
	dtype "ontology-query/interfaces/data_type"
)

type NotEmptyCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewNotEmptyCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	// 只允许字符串类型
	if !dtype.DataType_IsString(cfg.NameField.Type) {
		return nil, fmt.Errorf("condition [empty] left field %s is not of string type, but %s", cfg.Name, cfg.NameField.Type)
	}

	return &NotEmptyCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil

}

func (cond *NotEmptyCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	dslStr := fmt.Sprintf(`
	{
		"bool": {
			"must": {
				"exists": {
					"field": "%s"
				}
			},
			"must_not": {
				"term": {
					"%s": ""
				}
			}
		}
	}`, cond.mFilterFieldName, cond.mFilterFieldName)

	return dslStr, nil
}

// sql中没有字段存在的过滤条件,暂时用非空表达
func (cond *NotEmptyCond) Convert2SQL(ctx context.Context) (string, error) {
	sqlStr := fmt.Sprintf(`"%s" IS NOT NULL AND "%s" <> ''`, cond.mFilterFieldName, cond.mFilterFieldName)
	return sqlStr, nil
}

func rewriteNotEmptyCond(cfg *CondCfg) (*CondCfg, error) {

	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("非空值[not_empty]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}

	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}
