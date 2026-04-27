package dapo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
)

func TestConversationPO_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &ConversationPO{}
		tableName := po.TableName()

		expected := "t_data_agent_conversation"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestConversationPO(t *testing.T) {
	t.Parallel()

	t.Run("create conversation PO", func(t *testing.T) {
		t.Parallel()

		ext := "test extension"
		po := &ConversationPO{
			ID:               "conv-123",
			AgentAPPKey:      "agent-app-key",
			Title:            "Test Conversation",
			Origin:           cdaenum.ConversationWebChat,
			MessageIndex:     10,
			ReadMessageIndex: 5,
			Ext:              &ext,
			CreateTime:       1234567890,
			UpdateTime:       1234567890,
			CreateBy:         "user-1",
			UpdateBy:         "user-1",
			IsDeleted:        0,
		}

		if po.ID != "conv-123" {
			t.Errorf("Expected ID to be 'conv-123', got '%s'", po.ID)
		}

		if po.AgentAPPKey != "agent-app-key" {
			t.Errorf("Expected AgentAPPKey to be 'agent-app-key', got '%s'", po.AgentAPPKey)
		}

		if po.Title != "Test Conversation" {
			t.Errorf("Expected Title to be 'Test Conversation', got '%s'", po.Title)
		}

		if po.Origin != cdaenum.ConversationWebChat {
			t.Errorf("Expected Origin to be WebChat, got %v", po.Origin)
		}

		if po.MessageIndex != 10 {
			t.Errorf("Expected MessageIndex to be 10, got %d", po.MessageIndex)
		}

		if po.ReadMessageIndex != 5 {
			t.Errorf("Expected ReadMessageIndex to be 5, got %d", po.ReadMessageIndex)
		}

		if po.Ext == nil || *po.Ext != "test extension" {
			t.Error("Expected Ext to point to 'test extension'")
		}

		if po.CreateTime != 1234567890 {
			t.Errorf("Expected CreateTime to be 1234567890, got %d", po.CreateTime)
		}

		if po.UpdateTime != 1234567890 {
			t.Errorf("Expected UpdateTime to be 1234567890, got %d", po.UpdateTime)
		}

		if po.CreateBy != "user-1" {
			t.Errorf("Expected CreateBy to be 'user-1', got '%s'", po.CreateBy)
		}

		if po.UpdateBy != "user-1" {
			t.Errorf("Expected UpdateBy to be 'user-1', got '%s'", po.UpdateBy)
		}

		if po.IsDeleted != 0 {
			t.Errorf("Expected IsDeleted to be 0, got %d", po.IsDeleted)
		}
	})

	t.Run("conversation with nil extension", func(t *testing.T) {
		t.Parallel()

		po := &ConversationPO{
			ID:        "conv-456",
			Ext:       nil,
			IsDeleted: 1,
		}

		if po.ID != "conv-456" {
			t.Errorf("Expected ID to be 'conv-456', got '%s'", po.ID)
		}

		if po.Ext != nil {
			t.Error("Expected Ext to be nil")
		}

		if po.IsDeleted != 1 {
			t.Errorf("Expected IsDeleted to be 1, got %d", po.IsDeleted)
		}
	})
}
