package daresvo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentconfigvo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentRes_GetMiddleOutputVars_EmptyConfig(t *testing.T) {
	t.Parallel()

	res := &DataAgentRes{
		Answer: agentrespvo.NewAnswerS(),
	}
	res.finalAnswerVarHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}, VarFieldTypeFinalAnswer)
	res.middleOutputVarsHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
		MiddleOutputVars: []string{},
	}, VarFieldTypeMiddleOutputVars)

	middleVars, err := res.GetMiddleOutputVars()
	assert.NoError(t, err)
	assert.NotNil(t, middleVars)
}

func TestDataAgentRes_GetMiddleOutputVars_WithVars(t *testing.T) {
	t.Parallel()

	res := &DataAgentRes{
		Answer: agentrespvo.NewAnswerS(),
	}
	res.finalAnswerVarHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}, VarFieldTypeFinalAnswer)
	res.middleOutputVarsHelper = NewResHelper(res.Answer, &agentconfigvo.OutputVariablesS{
		MiddleOutputVars: []string{"mid_var1", "mid_var2"},
	}, VarFieldTypeMiddleOutputVars)

	// Set middle output vars
	res.Answer.SetField("mid_var1", "value1")
	res.Answer.SetField("mid_var2", "value2")

	middleVars, err := res.GetMiddleOutputVars()
	assert.NoError(t, err)
	assert.NotNil(t, middleVars)
}
