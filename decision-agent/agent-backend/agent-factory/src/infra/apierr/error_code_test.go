package apierr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentFactoryErrorCodes_Common(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "AgentFactory.InvalidParameter.RequestBody", AgentFactory_InvalidParameter_RequestBody)
	assert.Equal(t, "AgentFactory.InvalidRequestHeader.Authorization", AgentFactory_InvalidRequestHeader_Authorization)
	assert.Equal(t, "AgentFactory.Forbidden.FilterField", AgentFactory_Forbidden_FilterField)
	assert.Equal(t, "AgentFactory.InvalidRequestHeader.ContentType", AgentFactory_InvalidRequestHeader_ContentType)
	assert.Equal(t, "AgentFactory.InternalError", AgentFactory_InternalError)
}

func TestAgentFactoryErrorCodes_Release(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "AgentFactory.Release.InternalError.PublishFailed", AgentFactory_Release_InternalError_PublishFailed)
}

func TestAgentFactoryErrorCodes_NotEmpty(t *testing.T) {
	t.Parallel()

	assert.NotEmpty(t, AgentFactory_InvalidParameter_RequestBody)
	assert.NotEmpty(t, AgentFactory_InternalError)
}
