package dapo

import (
	"testing"
)

func TestReleaseHistoryPO_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &ReleaseHistoryPO{}
		tableName := po.TableName()

		expected := "t_data_agent_release_history"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestReleaseHistoryPO(t *testing.T) {
	t.Parallel()

	t.Run("create release history PO", func(t *testing.T) {
		t.Parallel()

		po := &ReleaseHistoryPO{
			ID:           "history-123",
			AgentID:      "agent-123",
			AgentConfig:  `{"key": "value"}`,
			AgentVersion: "v1.0",
			AgentDesc:    "Test description",
			CreateTime:   1234567890,
			UpdateTime:   1234567890,
			CreateBy:     "user-1",
			UpdateBy:     "user-1",
		}

		if po.ID != "history-123" {
			t.Errorf("Expected ID to be 'history-123', got '%s'", po.ID)
		}

		if po.AgentID != "agent-123" {
			t.Errorf("Expected AgentID to be 'agent-123', got '%s'", po.AgentID)
		}

		if po.AgentConfig != `{"key": "value"}` {
			t.Errorf("Expected AgentConfig to be '{\"key\": \"value\"}', got '%s'", po.AgentConfig)
		}

		if po.AgentVersion != "v1.0" {
			t.Errorf("Expected AgentVersion to be 'v1.0', got '%s'", po.AgentVersion)
		}

		if po.AgentDesc != "Test description" {
			t.Errorf("Expected AgentDesc to be 'Test description', got '%s'", po.AgentDesc)
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
	})

	t.Run("zero value release history", func(t *testing.T) {
		t.Parallel()

		var po ReleaseHistoryPO

		if po.ID != "" {
			t.Errorf("Expected ID to be empty, got '%s'", po.ID)
		}

		if po.AgentID != "" {
			t.Errorf("Expected AgentID to be empty, got '%s'", po.AgentID)
		}
	})
}
