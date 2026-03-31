package daconfvalobj

import (
	"github.com/pkg/errors"
)

// Input 表示输入参数配置
type Input struct {
	Fields  Fields   `json:"fields" binding:"required"` // 参数列表
	Rewrite *Rewrite `json:"rewrite"`                   // query重写
	Augment *Augment `json:"augment"`                   // query增强
}

func (p *Input) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{
		"Fields.required": `"fields"不能为空`,
	}
}

func (p *Input) ValObjCheck() (err error) {
	// 1. 检查Fields是否为空
	if p.Fields == nil {
		err = errors.New("[Input]: fields is required")
		return
	}

	// 2. 验证字段的有效性
	if err = p.Fields.ValObjCheck(); err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[Input]: fields is invalid")
		return
	}

	// 3. 如果Rewrite不为空，验证其有效性
	if p.Rewrite != nil {
		if err = p.Rewrite.ValObjCheck(); err != nil {
			// 包装错误信息，提供更详细的上下文
			err = errors.Wrap(err, "[Input]: rewrite is invalid")
			return
		}
	}

	// 4. 如果Augment不为空，验证其有效性
	if p.Augment != nil {
		if err = p.Augment.ValObjCheck(); err != nil {
			// 包装错误信息，提供更详细的上下文
			err = errors.Wrap(err, "[Input]: augment is invalid")
			return
		}
	}

	return
}
