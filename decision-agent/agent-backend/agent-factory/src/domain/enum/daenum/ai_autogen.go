package daenum

import "github.com/pkg/errors"

type AiAutogenFrom string

// EnumCheck 检查枚举值是否有效
func (e AiAutogenFrom) EnumCheck() error {
	switch e {
	case AiAutogenFromSystemPrompt, AiAutogenFromOpeningRemarks, AiAutogenFromPreSetQuestion:
		return nil
	default:
		return errors.New("invalid enum value")
	}
}

const (
	AiAutogenFromSystemPrompt   AiAutogenFrom = "system_prompt"   // 人设及回复逻辑
	AiAutogenFromOpeningRemarks AiAutogenFrom = "opening_remarks" // 开场白
	AiAutogenFromPreSetQuestion AiAutogenFrom = "preset_question" // 预设问题
)
