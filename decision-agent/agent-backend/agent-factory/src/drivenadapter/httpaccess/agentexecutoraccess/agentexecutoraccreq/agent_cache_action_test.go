package agentexecutoraccreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAgentCacheActionType_ToString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "upsert", AgentCacheActionUpsert.ToString())
	assert.Equal(t, "get_info", AgentCacheActionGetInfo.ToString())
}

func TestAgentCacheActionType_EnumCheck_Valid(t *testing.T) {
	t.Parallel()

	assert.NoError(t, AgentCacheActionUpsert.EnumCheck())
	assert.NoError(t, AgentCacheActionGetInfo.EnumCheck())
}

func TestAgentCacheActionType_EnumCheck_Invalid(t *testing.T) {
	t.Parallel()

	invalid := AgentCacheActionType("invalid")
	err := invalid.EnumCheck()
	assert.Error(t, err)
}
