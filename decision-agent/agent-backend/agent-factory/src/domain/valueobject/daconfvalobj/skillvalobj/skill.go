package skillvalobj

import "github.com/pkg/errors"

// 技能配置
type Skill struct {
	Tools  []*SkillTool  `json:"tools"`
	Agents []*SkillAgent `json:"agents"`
	MCPs   []*SkillMCP   `json:"mcps"`
	Skills []*SkillSkill `json:"skills"`
}

// ValObjCheck 验证工具配置
func (p *Skill) ValObjCheck() (err error) {
	for _, tool := range p.Tools {
		err = tool.ValObjCheck()
		if err != nil {
			return errors.Wrapf(err, "[Skill]: tools is invalid")
		}
	}

	for _, agent := range p.Agents {
		err = agent.ValObjCheck()
		if err != nil {
			return errors.Wrapf(err, "[Skill]: agents is invalid")
		}
	}

	for _, mcp := range p.MCPs {
		err = mcp.ValObjCheck()
		if err != nil {
			return errors.Wrapf(err, "[Skill]: mcp is invalid")
		}
	}

	for _, skill := range p.Skills {
		err = skill.ValObjCheck()
		if err != nil {
			return errors.Wrapf(err, "[Skill]: skills is invalid")
		}
	}

	return
}
