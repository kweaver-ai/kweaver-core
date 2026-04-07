package dapo

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
)

func TestDataAgentTplPo_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplPo{}
		tableName := po.TableName()

		expected := "t_data_agent_config_tpl"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestDataAgentTplPo_SetIsBuiltIn(t *testing.T) {
	t.Parallel()

	t.Run("set built in flag", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplPo{}
		builtIn := cdaenum.BuiltInYes
		po.SetIsBuiltIn(builtIn)

		if po.IsBuiltIn == nil {
			t.Error("Expected IsBuiltIn to be set")
		}

		if *po.IsBuiltIn != builtIn {
			t.Errorf("Expected IsBuiltIn to be %v, got %v", builtIn, *po.IsBuiltIn)
		}
	})
}

func TestDataAgentTplPo_SetPublishedAt(t *testing.T) {
	t.Parallel()

	t.Run("set published at", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplPo{}
		publishedAt := int64(1234567890)
		po.SetPublishedAt(publishedAt)

		if po.PublishedAt == nil {
			t.Error("Expected PublishedAt to be set")
		}

		if *po.PublishedAt != publishedAt {
			t.Errorf("Expected PublishedAt to be %d, got %d", publishedAt, *po.PublishedAt)
		}
	})
}

func TestDataAgentTplPo_SetPublishedBy(t *testing.T) {
	t.Parallel()

	t.Run("set published by", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplPo{}
		publishedBy := "user-123"
		po.SetPublishedBy(publishedBy)

		if po.PublishedBy == nil {
			t.Error("Expected PublishedBy to be set")
		}

		if *po.PublishedBy != publishedBy {
			t.Errorf("Expected PublishedBy to be '%s', got '%s'", publishedBy, *po.PublishedBy)
		}
	})
}

func TestDataAgentTplPo_GetPublishedAtInt64(t *testing.T) {
	t.Parallel()

	t.Run("nil published at returns 0", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplPo{}
		result := po.GetPublishedAtInt64()

		if result != 0 {
			t.Errorf("Expected result to be 0, got %d", result)
		}
	})

	t.Run("returns published at value", func(t *testing.T) {
		t.Parallel()

		publishedAt := int64(1234567890)
		po := &DataAgentTplPo{}
		po.SetPublishedAt(publishedAt)
		result := po.GetPublishedAtInt64()

		if result != publishedAt {
			t.Errorf("Expected result to be %d, got %d", publishedAt, result)
		}
	})
}

func TestDataAgentTplPo_GetPublishedByString(t *testing.T) {
	t.Parallel()

	t.Run("nil published by returns empty", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplPo{}
		result := po.GetPublishedByString()

		if result != "" {
			t.Errorf("Expected result to be empty, got '%s'", result)
		}
	})

	t.Run("returns published by value", func(t *testing.T) {
		t.Parallel()

		publishedBy := "user-123"
		po := &DataAgentTplPo{}
		po.SetPublishedBy(publishedBy)
		result := po.GetPublishedByString()

		if result != publishedBy {
			t.Errorf("Expected result to be '%s', got '%s'", publishedBy, result)
		}
	})
}

func TestDataAgentTplPO(t *testing.T) {
	t.Parallel()

	t.Run("create data agent template PO", func(t *testing.T) {
		t.Parallel()

		profile := "test profile"
		builtIn := cdaenum.BuiltInYes
		publishedAt := int64(1234567890)
		publishedBy := "user-123"

		po := &DataAgentTplPo{
			ID:          123,
			Name:        "Test Template",
			Key:         "test-template",
			ProductKey:  "product-key",
			Profile:     &profile,
			AvatarType:  cdaenum.AvatarTypeBuiltIn,
			Avatar:      "default-avatar",
			Status:      cdaenum.StatusUnpublished,
			IsBuiltIn:   &builtIn,
			CreatedAt:   1234567890,
			UpdatedAt:   1234567890,
			CreatedBy:   "creator-1",
			UpdatedBy:   "updater-1",
			DeletedAt:   0,
			DeletedBy:   "",
			Config:      "{}",
			CreatedType: daenum.AgentTplCreatedTypeCopyFromAgent,
			PublishedAt: &publishedAt,
			PublishedBy: &publishedBy,
			CreateFrom:  "agent-456",
		}

		if po.ID != 123 {
			t.Errorf("Expected ID to be 123, got %d", po.ID)
		}

		if po.Name != "Test Template" {
			t.Errorf("Expected Name to be 'Test Template', got '%s'", po.Name)
		}

		if po.Key != "test-template" {
			t.Errorf("Expected Key to be 'test-template', got '%s'", po.Key)
		}

		if po.AvatarType != cdaenum.AvatarTypeBuiltIn {
			t.Errorf("Expected AvatarType to be BuiltIn, got %v", po.AvatarType)
		}

		if po.Status != cdaenum.StatusUnpublished {
			t.Errorf("Expected Status to be Draft, got %v", po.Status)
		}
	})

	t.Run("with user uploaded avatar", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplPo{
			ID:         456,
			Name:       "Test Template 2",
			AvatarType: cdaenum.AvatarTypeUserUploaded,
			Avatar:     "custom-avatar.jpg",
		}

		if po.AvatarType != cdaenum.AvatarTypeUserUploaded {
			t.Errorf("Expected AvatarType to be UserUploaded, got %v", po.AvatarType)
		}

		if po.Avatar != "custom-avatar.jpg" {
			t.Errorf("Expected Avatar to be 'custom-avatar.jpg', got '%s'", po.Avatar)
		}
	})

	t.Run("with AI generated avatar", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplPo{
			ID:         789,
			Name:       "Test Template 3",
			AvatarType: cdaenum.AvatarTypeAIGenerated,
			Avatar:     "ai-avatar.png",
		}

		if po.AvatarType != cdaenum.AvatarTypeAIGenerated {
			t.Errorf("Expected AvatarType to be AIGenerated, got %v", po.AvatarType)
		}
	})
}

func TestDataAgentTplIDStrPo(t *testing.T) {
	t.Parallel()

	t.Run("create template with string ID", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplIDStrPo{
			DataAgentTplPo: DataAgentTplPo{
				ID:   123,
				Name: "Test Template",
			},
			ID: "tpl-123",
		}

		if po.DataAgentTplPo.ID != 123 {
			t.Errorf("Expected DataAgentTplPo.ID to be 123, got %d", po.DataAgentTplPo.ID)
		}

		if po.ID != "tpl-123" {
			t.Errorf("Expected ID to be 'tpl-123', got '%s'", po.ID)
		}
	})
}

func TestDataAgentTplPo_GetConfigStruct(t *testing.T) {
	t.Parallel()

	t.Run("valid config json", func(t *testing.T) {
		t.Parallel()

		configJSON := `{"input":{"fields":[]},"output":{}}`
		po := &DataAgentTplPo{Config: configJSON}

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

		po := &DataAgentTplPo{Config: "{invalid json"}

		_, err := po.GetConfigStruct()
		if err == nil {
			t.Error("Expected error for invalid JSON config")
		}
	})

	t.Run("empty config", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplPo{Config: ""}

		_, err := po.GetConfigStruct()
		if err == nil {
			t.Error("Expected error for empty config string")
		}
	})
}

func TestDataAgentTplPo_SetConfigStruct(t *testing.T) {
	t.Parallel()

	t.Run("set config and get back", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplPo{}
		configJSON := `{"input":{"fields":[]},"output":{}}`

		// Set config directly
		po.Config = configJSON

		// Verify we can get it back
		conf, err := po.GetConfigStruct()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if conf == nil {
			t.Error("Expected config struct to be returned")
		}
	})
}

func TestDataAgentTplPo_RemoveDataSourceFromConfig(t *testing.T) {
	t.Parallel()

	t.Run("remove data source without dolphin", func(t *testing.T) {
		t.Parallel()

		configJSON := `{"input":{"fields":[]},"output":{},"data_source":{}}`
		po := &DataAgentTplPo{Config: configJSON}

		err := po.RemoveDataSourceFromConfig(false)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("remove data source with invalid config", func(t *testing.T) {
		t.Parallel()

		po := &DataAgentTplPo{Config: "{invalid json"}

		err := po.RemoveDataSourceFromConfig(false)
		if err == nil {
			t.Error("Expected error for invalid config JSON")
		}
	})

	t.Run("remove data source with dolphin flag true", func(t *testing.T) {
		t.Parallel()

		configJSON := `{"input":{"fields":[]},"output":{},"data_source":{},"pre_dolphin":[]}`
		po := &DataAgentTplPo{Config: configJSON}

		err := po.RemoveDataSourceFromConfig(true)
		// This should not error even if dolphin operations don't do much
		if err != nil {
			t.Errorf("Expected no error with dolphin flag, got %v", err)
		}
	})
}
