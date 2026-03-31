package cdaenum

import "github.com/pkg/errors"

type InputFieldType string

// Check 检查输入字段类型是否合法
func (t InputFieldType) EnumCheck() (err error) {
	if t != InputFieldTypeString && t != InputFieldTypeFile && t != InputFieldTypeJSONObject {
		err = errors.New("[InputFieldType]: invalid input field type")
		return
	}

	return
}

const (
	// InputFieldTypeString 字符串
	InputFieldTypeString InputFieldType = "string"

	// InputFieldTypeFile 文件
	InputFieldTypeFile InputFieldType = "file"

	// InputFieldTypeJSONObject JSON对象
	InputFieldTypeJSONObject InputFieldType = "object"
)
