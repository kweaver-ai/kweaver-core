package daconfvalobj

import "github.com/pkg/errors"

// Rewrite 表示query重写配置
type Rewrite struct {
	Enable    *bool      `json:"enable" binding:"required"`     // 是否启用
	LlmConfig *LlmConfig `json:"llm_config" binding:"required"` // LLM配置
}

func (r *Rewrite) ValObjCheck() (err error) {
	// 检查Enable是否为空
	if r.Enable == nil {
		err = errors.New("[Rewrite]: enable is required")
		return
	}

	// 如果未启用重写功能，直接返回
	if !*r.Enable {
		return
	}

	// 如果启用了重写功能，检查LlmConfig是否为空
	if r.LlmConfig == nil {
		err = errors.New("[Rewrite]: llm_config is required when enable is true")
		return
	}

	// 验证LlmConfig的有效性
	if err = r.LlmConfig.ValObjCheck(); err != nil {
		// 包装错误信息，提供更详细的上下文
		err = errors.Wrap(err, "[Rewrite]: llm_config is invalid")
		return
	}

	return
}
