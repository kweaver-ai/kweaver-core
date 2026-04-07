package dapo

import (
	"database/sql"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
)

func TestListForBenchmarkPo(t *testing.T) {
	t.Parallel()

	t.Run("create ListForBenchmarkPo", func(t *testing.T) {
		t.Parallel()

		po := &ListForBenchmarkPo{
			ID:        "agent-123",
			Key:       "test-key",
			Name:      "Test Agent",
			Version:   "1.0.0",
			Status:    cdaenum.StatusPublished,
			UpdatedAt: 1234567890,
		}

		if po.ID != "agent-123" {
			t.Errorf("Expected ID to be 'agent-123', got '%s'", po.ID)
		}

		if po.Key != "test-key" {
			t.Errorf("Expected Key to be 'test-key', got '%s'", po.Key)
		}

		if po.Name != "Test Agent" {
			t.Errorf("Expected Name to be 'Test Agent', got '%s'", po.Name)
		}

		if po.Version != "1.0.0" {
			t.Errorf("Expected Version to be '1.0.0', got '%s'", po.Version)
		}

		if po.Status != cdaenum.StatusPublished {
			t.Errorf("Expected Status to be Published, got %v", po.Status)
		}

		if po.UpdatedAt != 1234567890 {
			t.Errorf("Expected UpdatedAt to be 1234567890, got %d", po.UpdatedAt)
		}
	})
}

func TestListForBenchmarkAgentPo(t *testing.T) {
	t.Parallel()

	t.Run("create ListForBenchmarkAgentPo", func(t *testing.T) {
		t.Parallel()

		po := &ListForBenchmarkAgentPo{
			ID:        "agent-123",
			Key:       "test-key",
			Name:      "Test Agent",
			Version:   "1.0.0",
			Status:    cdaenum.StatusPublished,
			UpdatedAt: 1234567890,
		}

		if po.ID != "agent-123" {
			t.Errorf("Expected ID to be 'agent-123', got '%s'", po.ID)
		}

		if po.Key != "test-key" {
			t.Errorf("Expected Key to be 'test-key', got '%s'", po.Key)
		}
	})
}

func TestListForBenchmarkReleasePo(t *testing.T) {
	t.Parallel()

	t.Run("create ListForBenchmarkReleasePo", func(t *testing.T) {
		t.Parallel()

		po := &ListForBenchmarkReleasePo{
			ID:        sql.NullString{String: "agent-123", Valid: true},
			Key:       sql.NullString{String: "test-key", Valid: true},
			Name:      sql.NullString{String: "Test Agent", Valid: true},
			UpdatedAt: sql.NullInt64{Int64: 1234567890, Valid: true},
			Version:   sql.NullString{String: "1.0.0", Valid: true},
			Status:    sql.NullString{String: "published", Valid: true},
		}

		if !po.ID.Valid {
			t.Error("Expected ID to be valid")
		}

		if po.ID.String != "agent-123" {
			t.Errorf("Expected ID to be 'agent-123', got '%s'", po.ID.String)
		}
	})
}

func TestListForBenchmarkReleasePo_ToListForBenchmarkPo(t *testing.T) {
	t.Parallel()

	t.Run("convert to ListForBenchmarkPo", func(t *testing.T) {
		t.Parallel()

		rp := &ListForBenchmarkReleasePo{
			ID:        sql.NullString{String: "agent-123", Valid: true},
			Key:       sql.NullString{String: "test-key", Valid: true},
			Name:      sql.NullString{String: "Test Agent", Valid: true},
			UpdatedAt: sql.NullInt64{Int64: 1234567890, Valid: true},
			Version:   sql.NullString{String: "1.0.0", Valid: true},
			Status:    sql.NullString{String: "published", Valid: true},
		}

		po := rp.ToListForBenchmarkPo()

		if po == nil {
			t.Fatal("Expected po to be non-nil")
		}

		if po.ID != "agent-123" {
			t.Errorf("Expected ID to be 'agent-123', got '%s'", po.ID)
		}

		if po.Key != "test-key" {
			t.Errorf("Expected Key to be 'test-key', got '%s'", po.Key)
		}

		if po.Name != "Test Agent" {
			t.Errorf("Expected Name to be 'Test Agent', got '%s'", po.Name)
		}

		if po.Version != "1.0.0" {
			t.Errorf("Expected Version to be '1.0.0', got '%s'", po.Version)
		}

		if po.UpdatedAt != 1234567890 {
			t.Errorf("Expected UpdatedAt to be 1234567890, got %d", po.UpdatedAt)
		}
	})

	t.Run("convert with null values", func(t *testing.T) {
		t.Parallel()

		rp := &ListForBenchmarkReleasePo{
			ID:        sql.NullString{String: "", Valid: false},
			Key:       sql.NullString{String: "", Valid: false},
			Name:      sql.NullString{String: "", Valid: false},
			UpdatedAt: sql.NullInt64{Int64: 0, Valid: false},
			Version:   sql.NullString{String: "", Valid: false},
			Status:    sql.NullString{String: "", Valid: false},
		}

		po := rp.ToListForBenchmarkPo()

		if po == nil {
			t.Fatal("Expected po to be non-nil")
		}

		// Null values should be converted to zero values
		if po.ID != "" {
			t.Errorf("Expected ID to be empty, got '%s'", po.ID)
		}

		if po.Key != "" {
			t.Errorf("Expected Key to be empty, got '%s'", po.Key)
		}

		if po.UpdatedAt != 0 {
			t.Errorf("Expected UpdatedAt to be 0, got %d", po.UpdatedAt)
		}
	})
}

func TestListForBenchmarkMerge(t *testing.T) {
	t.Parallel()

	t.Run("create ListForBenchmarkMerge", func(t *testing.T) {
		t.Parallel()

		merge := &ListForBenchmarkMerge{
			ListForBenchmarkAgentPo: ListForBenchmarkAgentPo{
				ID:      "agent-123",
				Key:     "test-key",
				Name:    "Test Agent",
				Version: "1.0.0",
				Status:  cdaenum.StatusPublished,
			},
			ListForBenchmarkReleasePo: ListForBenchmarkReleasePo{
				ID: sql.NullString{String: "release-123", Valid: true},
			},
		}

		// Access fields through the embedded struct
		if merge.ListForBenchmarkAgentPo.ID != "agent-123" {
			t.Errorf("Expected ID to be 'agent-123', got '%s'", merge.ListForBenchmarkAgentPo.ID)
		}

		if merge.ListForBenchmarkAgentPo.Key != "test-key" {
			t.Errorf("Expected Key to be 'test-key', got '%s'", merge.ListForBenchmarkAgentPo.Key)
		}
	})
}
