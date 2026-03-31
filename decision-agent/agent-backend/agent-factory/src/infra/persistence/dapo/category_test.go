package dapo

import (
	"testing"
)

func TestCategoryPO_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &CategoryPO{}
		tableName := po.TableName()

		expected := "t_data_agent_release_category"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestCategoryPO(t *testing.T) {
	t.Parallel()

	t.Run("create category PO", func(t *testing.T) {
		t.Parallel()

		po := &CategoryPO{
			ID:          "cat-123",
			Name:        "Test Category",
			Description: "Test Description",
			CreateTime:  1234567890,
			UpdateTime:  1234567890,
			CreateBy:    "user-1",
			UpdateBy:    "user-1",
		}

		if po.ID != "cat-123" {
			t.Errorf("Expected ID to be 'cat-123', got '%s'", po.ID)
		}

		if po.Name != "Test Category" {
			t.Errorf("Expected Name to be 'Test Category', got '%s'", po.Name)
		}

		if po.Description != "Test Description" {
			t.Errorf("Expected Description to be 'Test Description', got '%s'", po.Description)
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

	t.Run("zero value category PO", func(t *testing.T) {
		t.Parallel()

		var po CategoryPO

		if po.ID != "" {
			t.Errorf("Expected ID to be empty, got '%s'", po.ID)
		}

		if po.Name != "" {
			t.Errorf("Expected Name to be empty, got '%s'", po.Name)
		}

		if po.Description != "" {
			t.Errorf("Expected Description to be empty, got '%s'", po.Description)
		}

		if po.CreateTime != 0 {
			t.Errorf("Expected CreateTime to be 0, got %d", po.CreateTime)
		}
	})
}
