package dapo

import (
	"testing"
)

func TestVisitHistoryPO_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &VisitHistoryPO{}
		tableName := po.TableName()

		expected := "t_data_agent_visit_history"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestVisitHistoryPO(t *testing.T) {
	t.Parallel()

	t.Run("create visit history PO", func(t *testing.T) {
		t.Parallel()

		po := &VisitHistoryPO{
			ID:            "visit-123",
			AgentID:       "agent-123",
			AgentVersion:  "v1.0",
			VisitCount:    5,
			CustomSpaceID: "space-123",
			CreateTime:    1234567890,
			UpdateTime:    1234567890,
			CreateBy:      "user-1",
			UpdateBy:      "user-1",
		}

		if po.ID != "visit-123" {
			t.Errorf("Expected ID to be 'visit-123', got '%s'", po.ID)
		}

		if po.AgentID != "agent-123" {
			t.Errorf("Expected AgentID to be 'agent-123', got '%s'", po.AgentID)
		}

		if po.AgentVersion != "v1.0" {
			t.Errorf("Expected AgentVersion to be 'v1.0', got '%s'", po.AgentVersion)
		}

		if po.VisitCount != 5 {
			t.Errorf("Expected VisitCount to be 5, got %d", po.VisitCount)
		}

		if po.CustomSpaceID != "space-123" {
			t.Errorf("Expected CustomSpaceID to be 'space-123', got '%s'", po.CustomSpaceID)
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

	t.Run("zero value visit history", func(t *testing.T) {
		t.Parallel()

		var po VisitHistoryPO

		if po.ID != "" {
			t.Errorf("Expected ID to be empty, got '%s'", po.ID)
		}

		if po.VisitCount != 0 {
			t.Errorf("Expected VisitCount to be 0, got %d", po.VisitCount)
		}
	})
}
