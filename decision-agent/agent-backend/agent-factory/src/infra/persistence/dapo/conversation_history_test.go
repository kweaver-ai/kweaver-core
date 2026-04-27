package dapo

import (
	"testing"
)

func TestConversationHistoryLatestVisitAgentPO_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &ConversationHistoryLatestVisitAgentPO{}
		tableName := po.TableName()

		expected := "tb_conversation_history_v2"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestConversationHistoryLatestVisitAgentPO(t *testing.T) {
	t.Parallel()

	t.Run("create conversation history PO", func(t *testing.T) {
		t.Parallel()

		po := &ConversationHistoryLatestVisitAgentPO{
			AgentId:      "agent-123",
			LastModifyAt: 1234567890,
		}

		if po.AgentId != "agent-123" {
			t.Errorf("Expected AgentId to be 'agent-123', got '%s'", po.AgentId)
		}

		if po.LastModifyAt != 1234567890 {
			t.Errorf("Expected LastModifyAt to be 1234567890, got %d", po.LastModifyAt)
		}
	})
}
