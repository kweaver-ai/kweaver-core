package daresvo

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentconfigvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefOutputConf(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, DefOutputConf)
	assert.Equal(t, "answer", DefOutputConf.AnswerVar)
	assert.Equal(t, "doc_retrieval_res", DefOutputConf.DocRetrievalVar)
	assert.Equal(t, "graph_retrieval_res", DefOutputConf.GraphRetrievalVar)
	assert.Equal(t, "related_questions", DefOutputConf.RelatedQuestionsVar)
	assert.Empty(t, DefOutputConf.OtherVars)
}

func TestDataAgentRes_NewDataAgentRes_ValidJSON(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := []byte(`{
		"answer": {
			"final_answer": "Test answer"
		},
		"status": "completed"
	}`)

	outputVars := &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}

	res, err := NewDataAgentRes(ctx, data, outputVars)
	require.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Answer)
	assert.Equal(t, "completed", res.Status)
	assert.NotNil(t, res.finalAnswerVarHelper)
	assert.NotNil(t, res.docRetrievalVarHelper)
	assert.NotNil(t, res.graphRetrievalVarHelper)
	assert.NotNil(t, res.relatedQuestionsVarHelper)
	assert.NotNil(t, res.otherVarsHelper)
	assert.NotNil(t, res.middleOutputVarsHelper)
}

func TestDataAgentRes_NewDataAgentRes_InvalidJSON(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := []byte(`{invalid json`)

	outputVars := &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}

	res, err := NewDataAgentRes(ctx, data, outputVars)
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "loadFromMessage error")
}

func TestDataAgentRes_GetAnswerHelper(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	data := []byte(`{
		"answer": {
			"final_answer": "Test answer"
		}
	}`)

	outputVars := &agentconfigvo.OutputVariablesS{
		AnswerVar: "answer",
	}

	res, err := NewDataAgentRes(ctx, data, outputVars)
	require.NoError(t, err)

	helper := res.GetAnswerHelper()
	assert.NotNil(t, helper)
	assert.Same(t, res.finalAnswerVarHelper, helper)
}

func TestDataAgentRes_Fields(t *testing.T) {
	t.Parallel()

	res := &DataAgentRes{
		Answer:     agentrespvo.NewAnswerS(),
		Status:     "running",
		Error:      "test error",
		AgentRunID: "run-123",
	}

	assert.NotNil(t, res.Answer)
	assert.Equal(t, "running", res.Status)
	assert.Equal(t, "test error", res.Error)
	assert.Equal(t, "run-123", res.AgentRunID)
}
