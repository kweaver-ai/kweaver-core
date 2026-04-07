package dapo

import (
	"testing"
)

func TestProductPo_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &ProductPo{}
		tableName := po.TableName()

		expected := "t_product"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestProductPo(t *testing.T) {
	t.Parallel()

	t.Run("create product PO", func(t *testing.T) {
		t.Parallel()

		po := &ProductPo{
			ID:        123,
			Name:      "Test Product",
			Key:       "test-product",
			Profile:   "Test profile",
			CreatedAt: 1234567890,
			UpdatedAt: 1234567890,
			CreatedBy: "user-1",
			UpdatedBy: "user-1",
		}

		if po.ID != 123 {
			t.Errorf("Expected ID to be 123, got %d", po.ID)
		}

		if po.Name != "Test Product" {
			t.Errorf("Expected Name to be 'Test Product', got '%s'", po.Name)
		}

		if po.Key != "test-product" {
			t.Errorf("Expected Key to be 'test-product', got '%s'", po.Key)
		}

		if po.Profile != "Test profile" {
			t.Errorf("Expected Profile to be 'Test profile', got '%s'", po.Profile)
		}

		if po.CreatedAt != 1234567890 {
			t.Errorf("Expected CreatedAt to be 1234567890, got %d", po.CreatedAt)
		}

		if po.UpdatedAt != 1234567890 {
			t.Errorf("Expected UpdatedAt to be 1234567890, got %d", po.UpdatedAt)
		}

		if po.CreatedBy != "user-1" {
			t.Errorf("Expected CreatedBy to be 'user-1', got '%s'", po.CreatedBy)
		}

		if po.UpdatedBy != "user-1" {
			t.Errorf("Expected UpdatedBy to be 'user-1', got '%s'", po.UpdatedBy)
		}
	})

	t.Run("product with deletion info", func(t *testing.T) {
		t.Parallel()

		po := &ProductPo{
			ID:        456,
			DeletedBy: "user-2",
			DeletedAt: 9999999999,
		}

		if po.ID != 456 {
			t.Errorf("Expected ID to be 456, got %d", po.ID)
		}

		if po.DeletedBy != "user-2" {
			t.Errorf("Expected DeletedBy to be 'user-2', got '%s'", po.DeletedBy)
		}

		if po.DeletedAt != 9999999999 {
			t.Errorf("Expected DeletedAt to be 9999999999, got %d", po.DeletedAt)
		}
	})

	t.Run("zero value product", func(t *testing.T) {
		t.Parallel()

		var po ProductPo

		if po.ID != 0 {
			t.Errorf("Expected ID to be 0, got %d", po.ID)
		}

		if po.Name != "" {
			t.Errorf("Expected Name to be empty, got '%s'", po.Name)
		}

		if po.Key != "" {
			t.Errorf("Expected Key to be empty, got '%s'", po.Key)
		}

		if po.CreatedAt != 0 {
			t.Errorf("Expected CreatedAt to be 0, got %d", po.CreatedAt)
		}
	})
}
