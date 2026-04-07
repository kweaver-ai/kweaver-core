package daconfvalobj

import "github.com/pkg/errors"

// OpeningRemarkConfig 表示开场白配置
type OpeningRemarkConfig struct {
	Type                       string `json:"type"`                          // 开场白类型（固定/动态）
	FixedOpeningRemark         string `json:"fixed_opening_remark"`          // 固定开场白
	DynamicOpeningRemarkPrompt string `json:"dynamic_opening_remark_prompt"` // 动态开场白提示语
}

func (p *OpeningRemarkConfig) ValObjCheck() (err error) {
	// 检查Type是否为空
	if p.Type == "" {
		err = errors.New("[OpeningRemarkConfig]: type is required")
		return
	}

	// 检查Type的值是否有效（必须是"fixed"或"dynamic"）
	if p.Type != "fixed" && p.Type != "dynamic" {
		err = errors.New("[OpeningRemarkConfig]: type must be fixed or dynamic")
		return
	}

	// 如果Type为"fixed"，检查FixedOpeningRemark是否为空
	if p.Type == "fixed" && p.FixedOpeningRemark == "" {
		err = errors.New("[OpeningRemarkConfig]: fixed_opening_remark is required when type is fixed")
		return
	}

	// 如果Type为"dynamic"，检查DynamicOpeningRemarkPrompt是否为空
	if p.Type == "dynamic" && p.DynamicOpeningRemarkPrompt == "" {
		err = errors.New("[OpeningRemarkConfig]: dynamic_opening_remark_prompt is required when type is dynamic")
		return
	}

	return
}
