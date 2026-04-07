package dapo

import (
	"testing"
)

func TestReleaseCategoryRelPO_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &ReleaseCategoryRelPO{}
		tableName := po.TableName()

		expected := "t_data_agent_release_category_rel"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestReleaseCategoryRelPO(t *testing.T) {
	t.Parallel()

	t.Run("create release category relation PO", func(t *testing.T) {
		t.Parallel()

		po := &ReleaseCategoryRelPO{
			ID:         "rel-123",
			ReleaseID:  "release-123",
			CategoryID: "category-123",
		}

		if po.ID != "rel-123" {
			t.Errorf("Expected ID to be 'rel-123', got '%s'", po.ID)
		}

		if po.ReleaseID != "release-123" {
			t.Errorf("Expected ReleaseID to be 'release-123', got '%s'", po.ReleaseID)
		}

		if po.CategoryID != "category-123" {
			t.Errorf("Expected CategoryID to be 'category-123', got '%s'", po.CategoryID)
		}
	})

	t.Run("zero value relation", func(t *testing.T) {
		t.Parallel()

		var po ReleaseCategoryRelPO

		if po.ID != "" {
			t.Errorf("Expected ID to be empty, got '%s'", po.ID)
		}

		if po.ReleaseID != "" {
			t.Errorf("Expected ReleaseID to be empty, got '%s'", po.ReleaseID)
		}

		if po.CategoryID != "" {
			t.Errorf("Expected CategoryID to be empty, got '%s'", po.CategoryID)
		}
	})
}
