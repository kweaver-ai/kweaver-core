package dapo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
)

func TestConversationMsgPO_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &ConversationMsgPO{}
		tableName := po.TableName()

		expected := "t_data_agent_conversation_message"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestConversationMsgPO(t *testing.T) {
	t.Parallel()

	t.Run("create conversation message PO", func(t *testing.T) {
		t.Parallel()

		content := "Test content"
		ext := "test ext"
		po := &ConversationMsgPO{
			ID:             "msg-123",
			AgentAPPKey:    "agent-app-key",
			ConversationID: "conv-123",
			AgentID:        "agent-123",
			AgentVersion:   "v1.0",
			ReplyID:        "reply-123",
			Index:          1,
			Role:           cdaenum.MsgRoleUser,
			Content:        &content,
			ContentType:    cdaenum.MsgText,
			Status:         cdaenum.MsgStatusReceived,
			Ext:            &ext,
			CreateTime:     1234567890,
			UpdateTime:     1234567890,
			CreateBy:       "user-1",
			UpdateBy:       "user-1",
			IsDeleted:      0,
		}

		if po.ID != "msg-123" {
			t.Errorf("Expected ID to be 'msg-123', got '%s'", po.ID)
		}

		if po.AgentAPPKey != "agent-app-key" {
			t.Errorf("Expected AgentAPPKey to be 'agent-app-key', got '%s'", po.AgentAPPKey)
		}

		if po.ConversationID != "conv-123" {
			t.Errorf("Expected ConversationID to be 'conv-123', got '%s'", po.ConversationID)
		}

		if po.AgentID != "agent-123" {
			t.Errorf("Expected AgentID to be 'agent-123', got '%s'", po.AgentID)
		}

		if po.Index != 1 {
			t.Errorf("Expected Index to be 1, got %d", po.Index)
		}

		if po.Role != cdaenum.MsgRoleUser {
			t.Errorf("Expected Role to be User, got %v", po.Role)
		}

		if po.Content == nil || *po.Content != "Test content" {
			t.Error("Expected Content to point to 'Test content'")
		}

		if po.Ext == nil || *po.Ext != "test ext" {
			t.Error("Expected Ext to point to 'test ext'")
		}
	})

	t.Run("with nil content", func(t *testing.T) {
		t.Parallel()

		po := &ConversationMsgPO{
			ID:      "msg-456",
			Content: nil,
			Ext:     nil,
		}

		if po.ID != "msg-456" {
			t.Errorf("Expected ID to be 'msg-456', got '%s'", po.ID)
		}

		if po.Content != nil {
			t.Error("Expected Content to be nil")
		}

		if po.Ext != nil {
			t.Error("Expected Ext to be nil")
		}
	})
}
