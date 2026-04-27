package afresvo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAgentFactoryError(t *testing.T) {
	t.Parallel()

	err := NewAgentFactoryError()
	assert.NotNil(t, err)
}

func TestAgentFactoryError_Fields(t *testing.T) {
	t.Parallel()

	err := &AgentFactoryError{
		Description:  "Test error description",
		ErrorCode:    "ERR_001",
		ErrorDetails: "Detailed error info",
		Solution:     "Restart the service",
	}

	assert.Equal(t, "Test error description", err.Description)
	assert.Equal(t, "ERR_001", err.ErrorCode)
	assert.Equal(t, "Detailed error info", err.ErrorDetails)
	assert.Equal(t, "Restart the service", err.Solution)
}

func TestIsAgentFactoryError_ValidError(t *testing.T) {
	t.Parallel()

	jsonData := []byte(`{
		"Description": "Agent execution failed",
		"ErrorCode": "AGENT_ERR_001",
		"ErrorDetails": "LLM timeout",
		"Solution": "Check LLM service status"
	}`)

	afErr, isErr := IsAgentFactoryError(jsonData)

	assert.True(t, isErr)
	assert.Equal(t, "Agent execution failed", afErr.Description)
	assert.Equal(t, "AGENT_ERR_001", afErr.ErrorCode)
	assert.Equal(t, "LLM timeout", afErr.ErrorDetails)
	assert.Equal(t, "Check LLM service status", afErr.Solution)
}

func TestIsAgentFactoryError_NoErrorCode(t *testing.T) {
	t.Parallel()

	jsonData := []byte(`{
		"Description": "Normal response",
		"Data": "some data"
	}`)

	afErr, isErr := IsAgentFactoryError(jsonData)

	assert.False(t, isErr)
	assert.Empty(t, afErr.ErrorCode)
}

func TestIsAgentFactoryError_EmptyErrorCode(t *testing.T) {
	t.Parallel()

	jsonData := []byte(`{
		"ErrorCode": "",
		"Description": "Empty error code"
	}`)

	afErr, isErr := IsAgentFactoryError(jsonData)

	assert.False(t, isErr)
	assert.Empty(t, afErr.ErrorCode)
}

func TestIsAgentFactoryError_InvalidJSON(t *testing.T) {
	t.Parallel()

	jsonData := []byte(`invalid json`)

	afErr, isErr := IsAgentFactoryError(jsonData)

	assert.False(t, isErr)
	assert.Empty(t, afErr.ErrorCode)
}

func TestIsAgentFactoryError_PartialError(t *testing.T) {
	t.Parallel()

	jsonData := []byte(`{
		"ErrorCode": "ERR_001"
	}`)

	afErr, isErr := IsAgentFactoryError(jsonData)

	assert.True(t, isErr)
	assert.Equal(t, "ERR_001", afErr.ErrorCode)
}

func TestHandleAFErrorForChatProcess_NotAnError(t *testing.T) {
	t.Parallel()

	jsonData := []byte(`{"data": "normal response"}`)

	newData, isErr := HandleAFErrorForChatProcess(jsonData)

	assert.False(t, isErr)
	assert.Nil(t, newData)
}

func TestHandleAFErrorForChatProcess_IsError(t *testing.T) {
	t.Parallel()

	jsonData := []byte(`{
		"ErrorCode": "AGENT_ERR_001",
		"Description": "Test error"
	}`)

	newData, isErr := HandleAFErrorForChatProcess(jsonData)

	assert.True(t, isErr)
	assert.NotNil(t, newData)
	assert.NotEmpty(t, newData)
}
