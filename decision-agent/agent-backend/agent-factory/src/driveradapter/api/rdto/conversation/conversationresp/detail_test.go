package conversationresp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity/conversationeo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestNewConversationDetail(t *testing.T) {
	t.Parallel()

	detail := NewConversationDetail()

	assert.NotNil(t, detail)
}

func TestConversationDetail_LoadFromEo(t *testing.T) {
	t.Parallel()

	t.Run("load from conversation eo", func(t *testing.T) {
		t.Parallel()

		detail := NewConversationDetail()
		eo := &conversationeo.Conversation{}

		err := detail.LoadFromEo(eo)
		assert.NoError(t, err)
		assert.NotNil(t, detail.Conversation)
	})

	t.Run("with nil eo causes panic", func(t *testing.T) {
		t.Parallel()

		detail := NewConversationDetail()

		assert.Panics(t, func() {
			_ = detail.LoadFromEo(nil)
		})
	})
}

func TestConversationDetail_StructFields(t *testing.T) {
	t.Parallel()

	detail := &ConversationDetail{
		TempareaId: "temp-123",
		Status:     cdaenum.ConvStatusCompleted,
	}

	assert.Equal(t, "temp-123", detail.TempareaId)
	assert.Equal(t, cdaenum.ConvStatusCompleted, detail.Status)
}

func TestConversationDetail_Empty(t *testing.T) {
	t.Parallel()

	detail := &ConversationDetail{}

	assert.Empty(t, detail.TempareaId)
	assert.Empty(t, string(detail.Status))
}

func TestListConversationResp_Type(t *testing.T) {
	t.Parallel()

	var resp ListConversationResp

	// Verify it's a slice type
	resp = append(resp, ConversationDetail{})
	assert.Len(t, resp, 1)
}

func TestListConversationResp_Empty(t *testing.T) {
	t.Parallel()

	var resp ListConversationResp

	assert.Empty(t, resp)
}

func TestInitConversationResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := &InitConversationResp{
		ID:               "conv-123",
		SandboxSessionID: "session-456",
		XAccountID:       "user-789",
	}

	assert.Equal(t, "conv-123", resp.ID)
	assert.Equal(t, "session-456", resp.SandboxSessionID)
	assert.Equal(t, "user-789", resp.XAccountID)
}

func TestInitConversationResp_Empty(t *testing.T) {
	t.Parallel()

	resp := &InitConversationResp{}

	assert.Empty(t, resp.ID)
	assert.Empty(t, resp.SandboxSessionID)
	assert.Empty(t, resp.XAccountID)
}

func TestInitConversationResp_WithAllFields(t *testing.T) {
	t.Parallel()

	resp := &InitConversationResp{
		ID:               "test-id",
		SandboxSessionID: "test-session",
		XAccountID:       "test-account",
	}

	assert.Equal(t, "test-id", resp.ID)
	assert.Equal(t, "test-session", resp.SandboxSessionID)
	assert.Equal(t, "test-account", resp.XAccountID)
}

func TestConversationDetail_WithValues(t *testing.T) {
	t.Parallel()

	detail := &ConversationDetail{
		TempareaId: "area-123",
		Status:     cdaenum.ConvStatusProcessing,
	}

	assert.Equal(t, "area-123", detail.TempareaId)
	assert.Equal(t, cdaenum.ConvStatusProcessing, detail.Status)
}

func TestConversationDetail_StatusValues(t *testing.T) {
	t.Parallel()

	statuses := []cdaenum.ConversationStatus{
		cdaenum.ConvStatusCompleted,
		cdaenum.ConvStatusProcessing,
		cdaenum.ConvStatusFailed,
		cdaenum.ConvStatusCancelled,
	}

	for _, status := range statuses {
		detail := &ConversationDetail{
			Status: status,
		}
		assert.Equal(t, status, detail.Status)
	}
}

func TestConversationDetail_EmbeddedConversation(t *testing.T) {
	t.Parallel()

	detail := &ConversationDetail{}

	// The embedded Conversation struct should be accessible
	assert.NotNil(t, detail.Conversation)
}

func TestListConversationResp_Append(t *testing.T) {
	t.Parallel()

	var resp ListConversationResp

	detail1 := ConversationDetail{TempareaId: "1"}
	detail2 := ConversationDetail{TempareaId: "2"}

	resp = append(resp, detail1, detail2)

	assert.Len(t, resp, 2)
	assert.Equal(t, "1", resp[0].TempareaId)
	assert.Equal(t, "2", resp[1].TempareaId)
}
