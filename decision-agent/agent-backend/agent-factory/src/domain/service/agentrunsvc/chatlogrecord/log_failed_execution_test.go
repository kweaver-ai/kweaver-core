package chatlogrecord

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/conversationmsgvo"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	agentresp "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/resp"
	"github.com/stretchr/testify/assert"
)

func TestLogFailedExecution(t *testing.T) {
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

	// Test with nil response - should not panic
	LogFailedExecution(ctx, req, errors.New("test error"), nil)

	// Test with empty response
	LogFailedExecution(ctx, req, errors.New("test error"), &agentresp.ChatResp{})

	// Test with response containing data
	resp := &agentresp.ChatResp{
		Message: conversationmsgvo.Message{
			Content: map[string]interface{}{
				"middle_answer": map[string]interface{}{
					"progress": []interface{}{
						map[string]interface{}{
							"stage":  "skill",
							"status": "success",
						},
						map[string]interface{}{
							"stage":  "skill",
							"status": "failed",
						},
					},
				},
			},
			Ext: &conversationmsgvo.MessageExt{
				TotalTime:   1.5,
				TotalTokens: 1000,
			},
		},
	}

	LogFailedExecution(ctx, req, errors.New("test error"), resp)
}

func TestLogFailedExecutionWithEmptySessionID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	visitorInfo := &agentreq.InternalParam{
		UserID:       "user-456",
		ReqStartTime: 1234567890,
	}

	req := &agentreq.ChatReq{
		InternalParam:  *visitorInfo,
		AgentID:        "agent-123",
		ConversationID: "conv-789",
		// ConversationSessionID is empty - should be generated
		Query: "test query",
	}

	LogFailedExecution(ctx, req, errors.New("test error"), nil)

	// After the call, ConversationSessionID should be set
	assert.NotEmpty(t, req.ConversationSessionID)
	assert.Contains(t, req.ConversationSessionID, "conv-789")
}

func TestLogFailedExecution_NilProgress(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	req := &agentreq.ChatReq{
		InternalParam: agentreq.InternalParam{
			UserID:       "user-456",
			ReqStartTime: 1234567890,
		},
		AgentID:               "agent-123",
		ConversationID:        "conv-789",
		ConversationSessionID: "session-abc",
		Query:                 "test query",
	}

	// progress 值为 nil — 这是导致 panic 的场景
	respWithNilProgress := &agentresp.ChatResp{
		Message: conversationmsgvo.Message{
			Content: map[string]interface{}{
				"middle_answer": map[string]interface{}{
					"progress": nil,
				},
			},
		},
	}

	assert.NotPanics(t, func() {
		LogFailedExecution(ctx, req, errors.New("test error"), respWithNilProgress)
	})

	// progress 值为非预期类型 string
	respWithBadProgress := &agentresp.ChatResp{
		Message: conversationmsgvo.Message{
			Content: map[string]interface{}{
				"middle_answer": map[string]interface{}{
					"progress": "not-an-array",
				},
			},
		},
	}

	assert.NotPanics(t, func() {
		LogFailedExecution(ctx, req, errors.New("test error"), respWithBadProgress)
	})
}
