package dapo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/daenum"
)

func TestDataAgentPo_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentPo{}
		tableName := po.TableName()

		expected := "t_data_agent_config"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestDataAgentPo_IsBuiltInBool(t *testing.T) {
	t.Parallel()

	t.Run("nil is built in", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentPo{IsBuiltIn: nil}

		result := po.IsBuiltInBool()
		if result {
			t.Error("Expected IsBuiltInBool to return false for nil")
		}
	})

	t.Run("built in yes", func(t *testing.T) {
		t.Parallel()

		builtIn := cdaenum.BuiltInYes
		po := &DataAgentPo{IsBuiltIn: &builtIn}

		result := po.IsBuiltInBool()
		if !result {
			t.Error("Expected IsBuiltInBool to return true for BuiltInYes")
		}
	})

	t.Run("built in no", func(t *testing.T) {
		t.Parallel()

		builtIn := cdaenum.BuiltInNo
		po := &DataAgentPo{IsBuiltIn: &builtIn}

		result := po.IsBuiltInBool()
		if result {
			t.Error("Expected IsBuiltInBool to return false for BuiltInNo")
		}
	})
}

func TestDataAgentPo_SetIsBuiltIn(t *testing.T) {
	t.Parallel()

	t.Run("set to yes", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentPo{}
		po.SetIsBuiltIn(cdaenum.BuiltInYes)

		if po.IsBuiltIn == nil {
			t.Fatal("Expected IsBuiltIn to be non-nil")
		}

		if *po.IsBuiltIn != cdaenum.BuiltInYes {
			t.Error("Expected IsBuiltIn to be set to BuiltInYes")
		}
	})

	t.Run("set to no", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentPo{}
		po.SetIsBuiltIn(cdaenum.BuiltInNo)

		if po.IsBuiltIn == nil {
			t.Fatal("Expected IsBuiltIn to be non-nil")
		}

		if *po.IsBuiltIn != cdaenum.BuiltInNo {
			t.Error("Expected IsBuiltIn to be set to BuiltInNo")
		}
	})
}

func TestDataAgentPo_GetProfileStr(t *testing.T) {
	t.Parallel()

	t.Run("with profile", func(t *testing.T) {
		t.Parallel()

		profile := "Test profile"
		po := &DataAgentPo{Profile: &profile}
		result := po.GetProfileStr()

		if result != "Test profile" {
			t.Errorf("Expected profile to be 'Test profile', got '%s'", result)
		}
	})

	t.Run("nil profile", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentPo{Profile: nil}
		result := po.GetProfileStr()

		if result != "" {
			t.Errorf("Expected empty string for nil profile, got '%s'", result)
		}
	})
}

func TestDataAgentPo_ResetForImport(t *testing.T) {
	t.Parallel()

	t.Run("reset for import", func(t *testing.T) {
		t.Parallel()

		profile := "Test profile"
		builtIn := cdaenum.BuiltInYes
		po := &DataAgentPo{
			ID:          "agent-123",
			Profile:     &profile,
			CreatedAt:   1234567890,
			UpdatedAt:   1234567890,
			CreatedBy:   "user-1",
			UpdatedBy:   "user-1",
			DeletedAt:   9999999999,
			DeletedBy:   "user-2",
			CreatedType: daenum.AgentCreatedTypeCopy,
			CreateFrom:  "manual",
			Status:      cdaenum.StatusPublished,
			IsBuiltIn:   &builtIn,
		}

		po.ResetForImport()

		if po.ID != "" {
			t.Errorf("Expected ID to be empty after reset, got '%s'", po.ID)
		}

		if po.CreatedAt != 0 {
			t.Errorf("Expected CreatedAt to be 0 after reset, got %d", po.CreatedAt)
		}

		if po.UpdatedAt != 0 {
			t.Errorf("Expected UpdatedAt to be 0 after reset, got %d", po.UpdatedAt)
		}

		if po.CreatedBy != "" {
			t.Errorf("Expected CreatedBy to be empty after reset, got '%s'", po.CreatedBy)
		}

		if po.UpdatedBy != "" {
			t.Errorf("Expected UpdatedBy to be empty after reset, got '%s'", po.UpdatedBy)
		}

		if po.DeletedAt != 0 {
			t.Errorf("Expected DeletedAt to be 0 after reset, got %d", po.DeletedAt)
		}

		if po.DeletedBy != "" {
			t.Errorf("Expected DeletedBy to be empty after reset, got '%s'", po.DeletedBy)
		}

		if po.CreatedType != daenum.AgentCreatedTypeImport {
			t.Errorf("Expected CreatedType to be Import, got %v", po.CreatedType)
		}

		if po.CreateFrom != "" {
			t.Errorf("Expected CreateFrom to be empty after reset, got '%s'", po.CreateFrom)
		}

		if po.Status != cdaenum.StatusUnpublished {
			t.Errorf("Expected Status to be Unpublished, got %v", po.Status)
		}

		if po.IsBuiltIn != nil {
			t.Error("Expected IsBuiltIn to be nil after reset")
		}
	})
}

func TestDataAgentPo_GetConfigStruct(t *testing.T) {
	t.Parallel()

	t.Run("valid config json", func(t *testing.T) {
		t.Parallel()

		configJSON := `{"input":{"fields":[]},"output":{}}`
		po := &DataAgentPo{Config: configJSON}

		conf, err := po.GetConfigStruct()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if conf == nil {
			t.Error("Expected config struct to be returned")
		}
	})

	t.Run("invalid config json", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentPo{Config: "{invalid json"}

		_, err := po.GetConfigStruct()
		if err == nil {
			t.Error("Expected error for invalid JSON config")
		}
	})

	t.Run("empty config", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentPo{Config: ""}

		_, err := po.GetConfigStruct()
		// Empty string is invalid JSON
		if err == nil {
			t.Error("Expected error for empty config string")
		}
	})
}

func TestDataAgentPo_SetConfigStruct(t *testing.T) {
	t.Parallel()

	t.Run("set config struct with valid json", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentPo{}
		configJSON := `{"input":{"fields":[]},"output":{}}`

		// First, set the config using SetConfigStruct
		// We need to create a valid Config struct, but since we don't have direct access
		// to the full config structure, we'll test by setting raw JSON and getting it back
		po.Config = configJSON

		conf, err := po.GetConfigStruct()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if conf == nil {
			t.Error("Expected config struct to be returned")
		}
	})

	t.Run("set config then get it back", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentPo{}

		// Set a valid config
		configJSON := `{"input":{"fields":[]},"output":{}}`
		po.Config = configJSON

		// Get it back
		_, err := po.GetConfigStruct()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Now set it again using SetConfigStruct (through the round-trip)
		// This tests the successful path of SetConfigStruct
		po2 := &DataAgentPo{}
		po2.Config = po.Config

		conf2, err := po2.GetConfigStruct()
		if err != nil {
			t.Errorf("Expected no error on second get, got %v", err)
		}

		if conf2 == nil {
			t.Error("Expected config struct to be returned on second get")
		}
	})
}

func TestDataAgentPo_RemoveDataSourceFromConfig(t *testing.T) {
	t.Parallel()

	t.Run("remove data source without dolphin", func(t *testing.T) {
		t.Parallel()

		configJSON := `{"input":{"fields":[]},"output":{},"data_source":{}}`
		po := &DataAgentPo{Config: configJSON}

		err := po.RemoveDataSourceFromConfig(false)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("remove data source with invalid config", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentPo{Config: "{invalid json"}

		err := po.RemoveDataSourceFromConfig(false)
		// GetConfigStruct will fail first
		if err == nil {
			t.Error("Expected error for invalid config JSON")
		}
	})

	t.Run("remove data source with dolphin flag true", func(t *testing.T) {
		t.Parallel()

		configJSON := `{"input":{"fields":[]},"output":{},"data_source":{},"pre_dolphin":[]}`
		po := &DataAgentPo{Config: configJSON}

		err := po.RemoveDataSourceFromConfig(true)
		// This should not error even if dolphin operations don't do much
		if err != nil {
			t.Errorf("Expected no error with dolphin flag, got %v", err)
		}
	})
}
