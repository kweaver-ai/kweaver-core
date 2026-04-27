package skillvalobj

import "github.com/pkg/errors"

// SkillSkill 表示 skill 配置
type SkillSkill struct {
	SkillID string `json:"skill_id" binding:"required"` // skill ID
}

// ValObjCheck 验证 skill 配置
func (p *SkillSkill) ValObjCheck() (err error) {
	if p.SkillID == "" {
		err = errors.New("[Tool]: skill_id is required")
		return
	}

	return
}
