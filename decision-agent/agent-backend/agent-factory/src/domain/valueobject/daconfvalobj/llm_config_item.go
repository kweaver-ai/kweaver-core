package daconfvalobj

import (
	"github.com/pkg/errors"
)

// LlmItem 表示LLM配置项
type LlmItem struct {
	IsDefault bool `json:"is_default"` // 是否默认（从选择的所有llm中选择一个为默认llm）
	// 其他LLM属性
	LlmConfig *LlmConfig `json:"llm_config" binding:"required"` // LLM配置
}

func (p *LlmItem) ValObjCheck() (err error) {
	// 1. 检查LlmConfig是否为空
	if p.LlmConfig == nil {
		err = errors.New("[LlmItem]: llm_config is required")
		return
	}

	// 2. 验证LlmConfig的有效性
	if err = p.LlmConfig.ValObjCheck(); err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[LlmItem]: llm_config is invalid")
		return
	}

	return
}
