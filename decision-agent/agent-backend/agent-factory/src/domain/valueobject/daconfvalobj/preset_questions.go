package daconfvalobj

import "github.com/pkg/errors"

// PresetQuestion 表示预设问题
type PresetQuestion struct {
	Question string `json:"question"` // 问题内容
}

func (p *PresetQuestion) ValObjCheck() (err error) {
	// 检查Question是否为空
	if p.Question == "" {
		err = errors.New("[PresetQuestion]: question is required")
		return
	}

	return
}
