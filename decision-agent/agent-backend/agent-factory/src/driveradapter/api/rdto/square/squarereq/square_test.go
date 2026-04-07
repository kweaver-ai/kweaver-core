package squarereq

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestAgentInfoReq_StructFields(t *testing.T) {
	t.Parallel()

	req := AgentInfoReq{
		UserID:       "user-123",
		AgentID:      "agent-456",
		AgentVersion: "1.0.0",
		IsVisit:      true,
	}

	assert.Equal(t, "user-123", req.UserID)
	assert.Equal(t, "agent-456", req.AgentID)
	assert.Equal(t, "1.0.0", req.AgentVersion)
	assert.True(t, req.IsVisit)
}

func TestAgentInfoReq_EmptyValues(t *testing.T) {
	t.Parallel()

	req := AgentInfoReq{}

	assert.Empty(t, req.UserID)
	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentVersion)
	assert.False(t, req.IsVisit)
}

func TestAgentInfoReq_WithFalseIsVisit(t *testing.T) {
	t.Parallel()

	req := AgentInfoReq{
		UserID:       "user-test",
		AgentID:      "agent-test",
		AgentVersion: "2.0.0",
		IsVisit:      false,
	}

	assert.False(t, req.IsVisit)
}

func TestAgentSquareAgentReq_StructFields(t *testing.T) {
	t.Parallel()

	req := AgentSquareAgentReq{
		Name:        "TestAgent",
		CategoryID:  "cat-123",
		ReleaseIDS:  []string{"rel-1", "rel-2"},
		PublishToBe: cdaenum.PublishToBeAPIAgent,
	}

	assert.Equal(t, "TestAgent", req.Name)
	assert.Equal(t, "cat-123", req.CategoryID)
	assert.Len(t, req.ReleaseIDS, 2)
	assert.Equal(t, cdaenum.PublishToBeAPIAgent, req.PublishToBe)
}

func TestAgentSquareAgentReq_EmptyValues(t *testing.T) {
	t.Parallel()

	req := AgentSquareAgentReq{}

	assert.Empty(t, req.Name)
	assert.Empty(t, req.CategoryID)
	assert.Nil(t, req.ReleaseIDS)
	assert.Empty(t, req.PublishToBe)
}

func TestAgentSquareAgentReq_WithPageSize(t *testing.T) {
	t.Parallel()

	req := AgentSquareAgentReq{
		Name:        "PageAgent",
		CategoryID:  "cat-page",
		ReleaseIDS:  []string{"rel-1"},
		PublishToBe: cdaenum.PublishToBeWebSDKAgent,
	}
	req.Size = 25
	req.Page = 2

	assert.Equal(t, "PageAgent", req.Name)
	assert.Equal(t, 25, req.GetSize())
	assert.Equal(t, 2, req.GetPage())
	assert.Equal(t, 25, req.GetOffset()) // (2-1)*25
}

func TestAgentSquareAgentReq_EmptyReleaseIDS(t *testing.T) {
	t.Parallel()

	req := AgentSquareAgentReq{
		Name:        "EmptyReleases",
		CategoryID:  "cat-empty",
		ReleaseIDS:  []string{},
		PublishToBe: cdaenum.PublishToBeSkillAgent,
	}

	assert.Empty(t, req.ReleaseIDS)
	assert.Len(t, req.ReleaseIDS, 0)
}

func TestAgentSquareMyAgentReq_StructFields(t *testing.T) {
	t.Parallel()

	req := AgentSquareMyAgentReq{
		UserID:                    "user-my",
		Name:                      "MyAgent",
		ShouldContainBuiltInAgent: true,
	}

	assert.Equal(t, "user-my", req.UserID)
	assert.Equal(t, "MyAgent", req.Name)
	assert.True(t, req.ShouldContainBuiltInAgent)
}

func TestAgentSquareMyAgentReq_EmptyValues(t *testing.T) {
	t.Parallel()

	req := AgentSquareMyAgentReq{}

	assert.Empty(t, req.UserID)
	assert.Empty(t, req.Name)
	assert.False(t, req.ShouldContainBuiltInAgent)
}

func TestAgentSquareMyAgentReq_WithPageSize(t *testing.T) {
	t.Parallel()

	req := AgentSquareMyAgentReq{
		UserID:                    "user-page",
		Name:                      "PageAgent",
		ShouldContainBuiltInAgent: false,
	}
	req.Size = 50
	req.Page = 3

	assert.Equal(t, 50, req.GetSize())
	assert.Equal(t, 3, req.GetPage())
	assert.Equal(t, 100, req.GetOffset()) // (3-1)*50
}

func TestAgentSquareRecentAgentReq_StructFields(t *testing.T) {
	t.Parallel()

	req := AgentSquareRecentAgentReq{
		UserID:    "user-recent",
		Name:      "RecentAgent",
		StartTime: 1640995200000,
		EndTime:   1641081600000,
	}

	assert.Equal(t, "user-recent", req.UserID)
	assert.Equal(t, "RecentAgent", req.Name)
	assert.Equal(t, int64(1640995200000), req.StartTime)
	assert.Equal(t, int64(1641081600000), req.EndTime)
}

func TestAgentSquareRecentAgentReq_EmptyValues(t *testing.T) {
	t.Parallel()

	req := AgentSquareRecentAgentReq{}

	assert.Empty(t, req.UserID)
	assert.Empty(t, req.Name)
	assert.Equal(t, int64(0), req.StartTime)
	assert.Equal(t, int64(0), req.EndTime)
}

func TestAgentSquareRecentAgentReq_WithPageSize(t *testing.T) {
	t.Parallel()

	req := AgentSquareRecentAgentReq{
		UserID:    "user-time",
		Name:      "TimeAgent",
		StartTime: 1640995200000,
		EndTime:   1641081600000,
	}
	req.Size = 20
	req.Page = 1

	assert.Equal(t, 20, req.GetSize())
	assert.Equal(t, 1, req.GetPage())
	assert.Equal(t, 0, req.GetOffset()) // (1-1)*20
}

func TestAgentSquareRecentAgentReq_TimeRange(t *testing.T) {
	t.Parallel()

	req := AgentSquareRecentAgentReq{
		UserID:    "user-range",
		StartTime: 1640995200000,
		EndTime:   1672531200000, // One year later
	}

	assert.True(t, req.EndTime > req.StartTime)
}

func TestAgentSquareAgentReq_DifferentPublishToBe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		publishToBe cdaenum.PublishToBe
	}{
		{
			name:        "api agent",
			publishToBe: cdaenum.PublishToBeAPIAgent,
		},
		{
			name:        "web sdk agent",
			publishToBe: cdaenum.PublishToBeWebSDKAgent,
		},
		{
			name:        "skill agent",
			publishToBe: cdaenum.PublishToBeSkillAgent,
		},
		{
			name:        "data flow agent",
			publishToBe: cdaenum.PublishToBeDataFlowAgent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := AgentSquareAgentReq{
				Name:        tt.name,
				PublishToBe: tt.publishToBe,
			}

			assert.Equal(t, tt.publishToBe, req.PublishToBe)
		})
	}
}
