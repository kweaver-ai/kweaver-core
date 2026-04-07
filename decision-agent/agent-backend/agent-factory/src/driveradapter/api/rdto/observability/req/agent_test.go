package observabilityreq

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestAgentReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &AgentReq{
		AgentID:       "agent-123",
		AgentVersion:  "v1.0.0",
		IncludeConfig: true,
		StartTime:     1234567890,
		EndTime:       1234567999,
		XAccountID:    "account-456",
		XAccountType:  cenum.AccountTypeUser,
	}

	assert.Equal(t, "agent-123", req.AgentID)
	assert.Equal(t, "v1.0.0", req.AgentVersion)
	assert.True(t, req.IncludeConfig)
	assert.Equal(t, int64(1234567890), req.StartTime)
	assert.Equal(t, int64(1234567999), req.EndTime)
	assert.Equal(t, "account-456", req.XAccountID)
	assert.Equal(t, cenum.AccountTypeUser, req.XAccountType)
}

func TestAgentReq_Empty(t *testing.T) {
	t.Parallel()

	req := &AgentReq{}

	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentVersion)
	assert.False(t, req.IncludeConfig)
	assert.Equal(t, int64(0), req.StartTime)
	assert.Equal(t, int64(0), req.EndTime)
	assert.Empty(t, req.XAccountID)
}

func TestAgentReq_WithIncludeConfig(t *testing.T) {
	t.Parallel()

	req := &AgentReq{
		IncludeConfig: true,
	}

	assert.True(t, req.IncludeConfig)
}

func TestAgentReq_WithTimeRange(t *testing.T) {
	t.Parallel()

	req := &AgentReq{
		StartTime: 1000000000,
		EndTime:   2000000000,
	}

	assert.Equal(t, int64(1000000000), req.StartTime)
	assert.Equal(t, int64(2000000000), req.EndTime)
}

func TestAgentReq_WithAccountInfo(t *testing.T) {
	t.Parallel()

	req := &AgentReq{
		XAccountID:   "user-123",
		XAccountType: cenum.AccountTypeUser,
	}

	assert.Equal(t, "user-123", req.XAccountID)
	assert.Equal(t, cenum.AccountTypeUser, req.XAccountType)
}

func TestAgentReq_WithoutIncludeConfig(t *testing.T) {
	t.Parallel()

	req := &AgentReq{
		IncludeConfig: false,
	}

	assert.False(t, req.IncludeConfig)
}
