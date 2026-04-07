package cdaenum

import "github.com/pkg/errors"

type ToolInputParamValType string

const (
	// ToolInputParamValTypeFixedValue 固定值
	ToolInputParamValTypeFixedValue ToolInputParamValType = "fixedValue"

	// ToolInputParamValTypeVar 引用变量
	ToolInputParamValTypeVar ToolInputParamValType = "var"

	// ToolInputParamValTypeModel 选择模型
	ToolInputParamValTypeModel ToolInputParamValType = "model"

	// ToolInputParamValTypeAuto 模型生成
	ToolInputParamValTypeAuto ToolInputParamValType = "auto"
)

// Check 检查输入字段类型是否合法
func (t ToolInputParamValType) EnumCheck() (err error) {
	if t != ToolInputParamValTypeFixedValue && t != ToolInputParamValTypeVar && t != ToolInputParamValTypeModel && t != ToolInputParamValTypeAuto {
		err = errors.New("[ToolInputParamValType]: invalid input field type")
		return
	}

	return
}
