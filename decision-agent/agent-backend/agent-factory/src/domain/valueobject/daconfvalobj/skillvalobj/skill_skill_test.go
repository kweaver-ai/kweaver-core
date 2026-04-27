package skillvalobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkillSkill_ValObjCheck_Valid(t *testing.T) {
	t.Parallel()

	skill := &SkillSkill{
		SkillID: "skill-123",
	}

	err := skill.ValObjCheck()

	assert.NoError(t, err)
}

func TestSkillSkill_ValObjCheck_EmptySkillID(t *testing.T) {
	t.Parallel()

	skill := &SkillSkill{}

	err := skill.ValObjCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "skill_id is required")
}

func TestSkillSkill_NewSkillSkill(t *testing.T) {
	t.Parallel()

	skill := &SkillSkill{
		SkillID: "skill-demo-001",
	}

	assert.Equal(t, "skill-demo-001", skill.SkillID)
}

func TestSkillSkill_Empty(t *testing.T) {
	t.Parallel()

	skill := &SkillSkill{}

	assert.Empty(t, skill.SkillID)
}
