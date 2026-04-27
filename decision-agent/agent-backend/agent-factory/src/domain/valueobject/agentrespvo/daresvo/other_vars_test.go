package daresvo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentconfigvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentRes_GetOtherVarsMap(t *testing.T) {
	t.Parallel()

	res := &DataAgentRes{
		Answer: agentrespvo.NewAnswerS(),
	}
	res.finalAnswerVarHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}, VarFieldTypeFinalAnswer)
	res.otherVarsHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
		OtherVars: []string{"var1", "var2"},
	}, VarFieldTypeOther)

	res.Answer.SetField("var1", "value1")
	res.Answer.SetField("var2", "value2")

	m, err := res.GetOtherVarsMap()
	assert.NoError(t, err)
	assert.Equal(t, "value1", m["var1"])
	assert.Equal(t, "value2", m["var2"])
}

func TestDataAgentRes_GetOtherVarsMap_Empty(t *testing.T) {
	t.Parallel()

	res := &DataAgentRes{
		Answer: agentrespvo.NewAnswerS(),
	}
	res.finalAnswerVarHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}, VarFieldTypeFinalAnswer)
	res.otherVarsHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
		OtherVars: []string{},
	}, VarFieldTypeOther)

	m, err := res.GetOtherVarsMap()
	assert.NoError(t, err)
	// GetFields might return nil for empty slice
	if m != nil {
		assert.Empty(t, m)
	}
}
