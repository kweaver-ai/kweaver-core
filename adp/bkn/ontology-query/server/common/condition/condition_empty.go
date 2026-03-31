package condition

import (
	"context"
	"fmt"
	dtype "ontology-query/interfaces/data_type"
)

type EmptyCond struct {
	mCfg             *CondCfg
	mFilterFieldName string
}

func NewEmptyCond(ctx context.Context, cfg *CondCfg, fieldsMap map[string]*DataProperty) (Condition, error) {
	// 只允许字符串类型
	if !dtype.DataType_IsString(cfg.NameField.Type) {
		return nil, fmt.Errorf("condition [empty] left field %s is not of string type, but %s", cfg.Name, cfg.NameField.Type)
	}

	return &EmptyCond{
		mCfg:             cfg,
		mFilterFieldName: getFilterFieldName(cfg.Name, fieldsMap, false),
	}, nil

}

func (cond *EmptyCond) Convert(ctx context.Context, vectorizer func(ctx context.Context, property *DataProperty, word string) ([]VectorResp, error)) (string, error) {
	dslStr := `
	{
		"exists": {
			"field": "%s"
		}
	}
	`

	return fmt.Sprintf(dslStr, cond.mFilterFieldName), nil
}

// sql中没有字段存在的过滤条件,暂时用非空表达
func (cond *EmptyCond) Convert2SQL(ctx context.Context) (string, error) {
	return fmt.Sprintf(`"%s" IS NOT NULL`, cond.mFilterFieldName), nil
}

func rewriteEmptyCond(cfg *CondCfg) (*CondCfg, error) {

	// 过滤条件中的属性字段换成映射的视图字段
	if cfg.NameField.Name == "" {
		return nil, fmt.Errorf("空值[empty]操作符使用的过滤字段[%s]在对象类的属性中不存在", cfg.Name)
	}

	return &CondCfg{
		Name:        cfg.NameField.MappedField.Name,
		Operation:   cfg.Operation,
		ValueOptCfg: cfg.ValueOptCfg,
	}, nil
}
