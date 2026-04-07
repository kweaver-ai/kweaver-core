package dapo

import (
	"testing"
)

func TestPubTplCatAssocPo_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &PubTplCatAssocPo{}
		tableName := po.TableName()

		expected := "t_data_agent_tpl_category_rel"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestPubTplCatAssocPo(t *testing.T) {
	t.Parallel()

	t.Run("create published template category association PO", func(t *testing.T) {
		t.Parallel()

		po := &PubTplCatAssocPo{
			ID:             123,
			PublishedTplID: 456,
			CategoryID:     "cat-789",
		}

		if po.ID != 123 {
			t.Errorf("Expected ID to be 123, got %d", po.ID)
		}

		if po.PublishedTplID != 456 {
			t.Errorf("Expected PublishedTplID to be 456, got %d", po.PublishedTplID)
		}

		if po.CategoryID != "cat-789" {
			t.Errorf("Expected CategoryID to be 'cat-789', got '%s'", po.CategoryID)
		}
	})

	t.Run("zero value association", func(t *testing.T) {
		t.Parallel()

		var po PubTplCatAssocPo

		if po.ID != 0 {
			t.Errorf("Expected ID to be 0, got %d", po.ID)
		}

		if po.PublishedTplID != 0 {
			t.Errorf("Expected PublishedTplID to be 0, got %d", po.PublishedTplID)
		}

		if po.CategoryID != "" {
			t.Errorf("Expected CategoryID to be empty, got '%s'", po.CategoryID)
		}
	})
}

func TestDataAgentTplCategoryJoinPo(t *testing.T) {
	t.Parallel()

	t.Run("create template category join PO", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplCategoryJoinPo{
			ID:             123,
			PublishedTplID: 456,
			CategoryID:     "cat-789",
			CategoryName:   "Test Category",
		}

		if po.ID != 123 {
			t.Errorf("Expected ID to be 123, got %d", po.ID)
		}

		if po.PublishedTplID != 456 {
			t.Errorf("Expected PublishedTplID to be 456, got %d", po.PublishedTplID)
		}

		if po.CategoryID != "cat-789" {
			t.Errorf("Expected CategoryID to be 'cat-789', got '%s'", po.CategoryID)
		}

		if po.CategoryName != "Test Category" {
			t.Errorf("Expected CategoryName to be 'Test Category', got '%s'", po.CategoryName)
		}
	})

	t.Run("with empty category name", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplCategoryJoinPo{
			ID:             456,
			PublishedTplID: 789,
			CategoryID:     "cat-123",
			CategoryName:   "",
		}

		if po.CategoryName != "" {
			t.Errorf("Expected CategoryName to be empty, got '%s'", po.CategoryName)
		}
	})
}
