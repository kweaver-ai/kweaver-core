package agentresperr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRespErrorType_Constants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, RespErrorType("agent_factory"), RespErrorTypeAgentFactory)
	assert.Equal(t, RespErrorType("agent_executor"), RespErrorTypeAgentExecutor)
}

func TestNewRespError(t *testing.T) {
	t.Parallel()

	err := NewRespError(RespErrorTypeAgentFactory, "test error")
	assert.NotNil(t, err)
	assert.Equal(t, RespErrorTypeAgentFactory, err.Type)
	assert.Equal(t, "test error", err.Error)
}

func TestNewRespError_AgentExecutor(t *testing.T) {
	t.Parallel()

	err := NewRespError(RespErrorTypeAgentExecutor, map[string]string{"code": "500"})
	assert.NotNil(t, err)
	assert.Equal(t, RespErrorTypeAgentExecutor, err.Type)
	assert.NotNil(t, err.Error)
}

func TestNewRespError_NilError(t *testing.T) {
	t.Parallel()

	err := NewRespError(RespErrorTypeAgentFactory, nil)
	assert.NotNil(t, err)
	assert.Equal(t, RespErrorTypeAgentFactory, err.Type)
	assert.Nil(t, err.Error)
}
