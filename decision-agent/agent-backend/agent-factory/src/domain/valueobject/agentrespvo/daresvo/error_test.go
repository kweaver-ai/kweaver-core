package daresvo

import (
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentresperr"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentRes_GetExecutorError_NoError(t *testing.T) {
	t.Parallel()

	res := &DataAgentRes{
		Error: nil,
	}

	respErr := res.GetExecutorError()
	assert.Nil(t, respErr)
}

func TestDataAgentRes_GetExecutorError_WithError(t *testing.T) {
	t.Parallel()

	testError := errors.New("test error")
	res := &DataAgentRes{
		Error: testError,
	}

	respErr := res.GetExecutorError()
	assert.NotNil(t, respErr)
	assert.Equal(t, agentresperr.RespErrorTypeAgentExecutor, respErr.Type)
	assert.Equal(t, testError, respErr.Error)
}
