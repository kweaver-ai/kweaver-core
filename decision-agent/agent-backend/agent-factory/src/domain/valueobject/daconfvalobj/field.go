package daconfvalobj

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/pkg/errors"
)

// Field 表示agent参数（字段）
type Field struct {
	Name string                 `json:"name" binding:"required"` // 参数（字段）名
	Type cdaenum.InputFieldType `json:"type"`                    // 类型：string-字符串, file-文件, object-json对象
	Desc string                 `json:"desc"`                    // 参数描述
}

func (p *Field) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{
		"Name.required": `"name"不能为空`,
	}
}

func (p *Field) ValObjCheck() (err error) {
	// 检查Name是否为空
	if p.Name == "" {
		err = errors.New("[Field]: name is required")
		return
	}

	// 验证Type枚举值的有效性
	if err = p.Type.EnumCheck(); err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[Field]: type is invalid")
		return
	}

	return
}
