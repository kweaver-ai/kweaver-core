package chatlogrecord

import (
	"context"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
)

func TestLogSuccessExecution(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	visitorInfo := &agentreq.InternalParam{
		UserID:       "user-456",
		CallType:     "chat",
		ReqStartTime: 1234567890,
		TTFT:         100,
	}

	req := &agentreq.ChatReq{
		InternalParam:         *visitorInfo,
		AgentID:               "agent-123",
		AgentVersion:          "1.0",
		ConversationID:        "conv-789",
		ConversationSessionID: "session-abc",
		AgentRunID:            "run-xyz",
		Query:                 "test query",
	}

	progresses := []*agentrespvo.Progress{
		{
			Stage:  "llm",
			Status: "success",
		},
		{
			Stage:  "skill",
			Status: "success",
		},
		{
			Stage:  "skill",
			Status: "failed",
		},
	}

	// Test logging successful execution - should not panic
	LogSuccessExecution(ctx, req, progresses, 1.5, 1000)

	// Test with empty progress
	LogSuccessExecution(ctx, req, []*agentrespvo.Progress{}, 0, 0)

	// Test with nil progress
	LogSuccessExecution(ctx, req, nil, 0.5, 500)
}

func TestLogSuccessExecutionWithVariousProgressTypes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	visitorInfo := &agentreq.InternalParam{
		UserID:       "user-456",
		CallType:     "chat",
		ReqStartTime: 1234567890,
	}

	req := &agentreq.ChatReq{
		InternalParam:         *visitorInfo,
		AgentID:               "agent-123",
		ConversationID:        "conv-789",
		ConversationSessionID: "session-abc",
		AgentRunID:            "run-xyz",
		Query:                 "test query",
	}

	// Test with only LLM stages
	llmProgresses := []*agentrespvo.Progress{
		{Stage: "llm", Status: "success"},
		{Stage: "llm", Status: "success"},
	}
	LogSuccessExecution(ctx, req, llmProgresses, 2.0, 1500)

	// Test with only skill stages
	skillProgresses := []*agentrespvo.Progress{
		{Stage: "skill", Status: "success"},
		{Stage: "skill", Status: "failed"},
		{Stage: "skill", Status: "success"},
	}
	LogSuccessExecution(ctx, req, skillProgresses, 3.0, 2000)

	// Test with mixed stages
	mixedProgresses := []*agentrespvo.Progress{
		{Stage: "llm", Status: "success"},
		{Stage: "skill", Status: "success"},
		{Stage: "llm", Status: "success"},
		{Stage: "skill", Status: "failed"},
	}
	LogSuccessExecution(ctx, req, mixedProgresses, 4.0, 2500)
}
