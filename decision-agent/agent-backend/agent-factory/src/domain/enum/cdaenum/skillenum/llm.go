package skillenum

import (
	"github.com/pkg/errors"
)

type LLM string

const (
	LLMInheritMain  LLM = "inherit_main"    // 继承主 Agent 大模型
	LLMSelfConfiged LLM = "self_configured" // 使用自身配置（默认逻辑）
)

func (b LLM) EnumCheck() (err error) {
	if b != LLMInheritMain && b != LLMSelfConfiged {
		err = errors.New("[LLM]: invalid skill agent llm")
		return
	}

	return
}
