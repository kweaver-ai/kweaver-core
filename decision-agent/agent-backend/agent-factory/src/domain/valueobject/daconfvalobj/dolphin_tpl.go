package daconfvalobj

import (
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/pkg/errors"
)

// dolphin_tpl_req:
//   type: object
//   description: dolphin块的配置
//   properties:
//     key:
//       type: string
//       description: dolphin块的key
//     name:
//       type: string
//       description: dolphin块的名称
//     value:
//       type: string
//       description: dolphin块的dolphin语句
//     enabled:
//       type: boolean
//       description: 是否启用
//       default: true
//     edited:
//       type: boolean
//       description: 是否编辑过
//       default: false
//   required:
//     - key
//     - value

type DolphinTpl struct {
	Key     cdaenum.DolphinTplKey `json:"key" binding:"required"`   // dolphin块的key
	Name    string                `json:"name"`                     // dolphin块的名称
	Value   string                `json:"value" binding:"required"` // dolphin块的dolphin语句
	Enabled bool                  `json:"enabled"`                  // 是否启用
	Edited  bool                  `json:"edited"`                   // 是否编辑过
}

func (p *DolphinTpl) GetErrMsgMap() map[string]string {
	// 返回错误信息映射，用于将验证错误转换为用户友好的错误消息
	return map[string]string{
		"Key.required":   `"key"不能为空`,
		"Value.required": `"value"不能为空`,
	}
}

func (p *DolphinTpl) ValObjCheck() (err error) {
	// 检查Key是否为空
	if p.Key == "" {
		err = errors.New("[DolphinTpl]: key is required")
		return
	}

	// 检查Value是否为空
	if p.Value == "" {
		err = errors.New("[DolphinTpl]: value is required")
		return
	}

	if err = p.Key.EnumCheck(); err != nil {
		err = errors.Wrap(err, "[DolphinTpl]: key is invalid")
		return
	}

	return
}
