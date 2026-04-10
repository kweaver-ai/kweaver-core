package observabilityresp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
)

func TestPublishInfo_StructFields(t *testing.T) {
	t.Parallel()

	info := PublishInfo{
		IsAPIAgent:      1,
		IsSDKAgent:      0,
		IsSkillAgent:    1,
		IsDataFlowAgent: 0,
	}

	assert.Equal(t, 1, info.IsAPIAgent)
	assert.Equal(t, 0, info.IsSDKAgent)
	assert.Equal(t, 1, info.IsSkillAgent)
	assert.Equal(t, 0, info.IsDataFlowAgent)
}

func TestPublishInfo_Empty(t *testing.T) {
	t.Parallel()

	info := PublishInfo{}

	assert.Zero(t, info.IsAPIAgent)
	assert.Zero(t, info.IsSDKAgent)
	assert.Zero(t, info.IsSkillAgent)
	assert.Zero(t, info.IsDataFlowAgent)
}

func TestAgent_StructFields(t *testing.T) {
	t.Parallel()

	config := daconfvalobj.Config{}
	agent := Agent{
		ID:           "agent-123",
		Version:      "v1.0.0",
		Key:          "test-key",
		IsBuiltIn:    1,
		Name:         "Test Agent",
		CategoryID:   "cat-456",
		CategoryName: "Test Category",
		Profile:      "Test profile",
		Config:       config,
		AvatarType:   1,
		Avatar:       "🤖",
		ProductID:    100,
		ProductName:  "Test Product",
	}

	assert.Equal(t, "agent-123", agent.ID)
	assert.Equal(t, "v1.0.0", agent.Version)
	assert.Equal(t, "test-key", agent.Key)
	assert.Equal(t, 1, agent.IsBuiltIn)
	assert.Equal(t, "Test Agent", agent.Name)
	assert.Equal(t, "cat-456", agent.CategoryID)
	assert.Equal(t, "Test Category", agent.CategoryName)
	assert.Equal(t, "Test profile", agent.Profile)
	assert.Equal(t, 1, agent.AvatarType)
	assert.Equal(t, "🤖", agent.Avatar)
	assert.Equal(t, 100, agent.ProductID)
	assert.Equal(t, "Test Product", agent.ProductName)
}

func TestAgent_Empty(t *testing.T) {
	t.Parallel()

	agent := Agent{}

	assert.Empty(t, agent.ID)
	assert.Empty(t, agent.Version)
	assert.Empty(t, agent.Key)
	assert.Zero(t, agent.IsBuiltIn)
	assert.Empty(t, agent.Name)
	assert.Empty(t, agent.CategoryID)
	assert.Empty(t, agent.CategoryName)
	assert.Empty(t, agent.Profile)
	assert.Zero(t, agent.AvatarType)
	assert.Empty(t, agent.Avatar)
	assert.Zero(t, agent.ProductID)
	assert.Empty(t, agent.ProductName)
}

func TestAgentResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := AgentResp{
		TotalRequests:      1000,
		TotalSessions:      500,
		AvgSessionRounds:   3,
		RunSuccessRate:     0.95,
		AvgExecuteDuration: 2000,
		AvgTTFTDuration:    150,
		ToolSuccessRate:    0.88,
	}

	assert.Equal(t, 1000, resp.TotalRequests)
	assert.Equal(t, 500, resp.TotalSessions)
	assert.Equal(t, 3, resp.AvgSessionRounds)
	assert.Equal(t, float32(0.95), resp.RunSuccessRate)
	assert.Equal(t, 2000, resp.AvgExecuteDuration)
	assert.Equal(t, 150, resp.AvgTTFTDuration)
	assert.Equal(t, float32(0.88), resp.ToolSuccessRate)
}

func TestAgentResp_Empty(t *testing.T) {
	t.Parallel()

	resp := AgentResp{}

	assert.Zero(t, resp.TotalRequests)
	assert.Zero(t, resp.TotalSessions)
	assert.Zero(t, resp.AvgSessionRounds)
	assert.Zero(t, resp.RunSuccessRate)
	assert.Zero(t, resp.AvgExecuteDuration)
	assert.Zero(t, resp.AvgTTFTDuration)
	assert.Zero(t, resp.ToolSuccessRate)
}

func TestAgentResp_WithAgent(t *testing.T) {
	t.Parallel()

	agent := Agent{
		ID:   "test-agent",
		Name: "Test",
	}
	resp := AgentResp{
		Agent:          agent,
		TotalRequests:  100,
		RunSuccessRate: 0.9,
	}

	assert.Equal(t, "test-agent", resp.Agent.ID)
	assert.Equal(t, 100, resp.TotalRequests)
	assert.Equal(t, float32(0.9), resp.RunSuccessRate)
}

func TestPublishInfo_AllTrue(t *testing.T) {
	t.Parallel()

	info := PublishInfo{
		IsAPIAgent:      1,
		IsSDKAgent:      1,
		IsSkillAgent:    1,
		IsDataFlowAgent: 1,
	}

	assert.Equal(t, 1, info.IsAPIAgent)
	assert.Equal(t, 1, info.IsSDKAgent)
	assert.Equal(t, 1, info.IsSkillAgent)
	assert.Equal(t, 1, info.IsDataFlowAgent)
}

func TestAgent_WithPublishInfo(t *testing.T) {
	t.Parallel()

	publishInfo := PublishInfo{
		IsAPIAgent:      1,
		IsSDKAgent:      0,
		IsSkillAgent:    1,
		IsDataFlowAgent: 0,
	}
	agent := Agent{
		ID:          "agent-with-publish",
		PublishInfo: publishInfo,
	}

	assert.Equal(t, "agent-with-publish", agent.ID)
	assert.Equal(t, 1, agent.PublishInfo.IsAPIAgent)
	assert.Equal(t, 0, agent.PublishInfo.IsSDKAgent)
}

func TestAgentResp_WithChineseName(t *testing.T) {
	t.Parallel()

	agent := Agent{
		Name: "测试智能体",
	}
	resp := AgentResp{
		Agent: agent,
	}

	assert.Equal(t, "测试智能体", resp.Agent.Name)
}

func TestAgentResp_WithAllMetrics(t *testing.T) {
	t.Parallel()

	resp := AgentResp{
		TotalRequests:      10000,
		TotalSessions:      5000,
		AvgSessionRounds:   5,
		RunSuccessRate:     0.98,
		AvgExecuteDuration: 3000,
		AvgTTFTDuration:    200,
		ToolSuccessRate:    0.95,
	}

	assert.Equal(t, 10000, resp.TotalRequests)
	assert.Equal(t, 5000, resp.TotalSessions)
	assert.Equal(t, 5, resp.AvgSessionRounds)
	assert.InDelta(t, float32(0.98), resp.RunSuccessRate, 0.01)
	assert.Equal(t, 3000, resp.AvgExecuteDuration)
	assert.Equal(t, 200, resp.AvgTTFTDuration)
	assert.InDelta(t, float32(0.95), resp.ToolSuccessRate, 0.01)
}
