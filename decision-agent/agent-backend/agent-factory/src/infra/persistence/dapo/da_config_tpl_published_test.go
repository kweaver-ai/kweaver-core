package dapo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
)

func TestPublishedTplPo_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &PublishedTplPo{}
		tableName := po.TableName()

		expected := "t_data_agent_config_tpl_published"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestPublishedTplPo_SetIsBuiltIn(t *testing.T) {
	t.Parallel()

	t.Run("set built in flag", func(t *testing.T) {
		t.Parallel()

		po := &PublishedTplPo{}
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

func TestPublishedTplPo(t *testing.T) {
	t.Parallel()

	t.Run("create published template PO", func(t *testing.T) {
		t.Parallel()

		profile := "test profile"
		builtIn := cdaenum.BuiltInYes

		po := &PublishedTplPo{
			ID:          123,
			Name:        "Published Template",
			Key:         "published-template",
			ProductKey:  "product-key",
			Profile:     &profile,
			AvatarType:  cdaenum.AvatarTypeBuiltIn,
			Avatar:      "default-avatar",
			IsBuiltIn:   &builtIn,
			Config:      `{"key": "value"}`,
			PublishedAt: 1234567890,
			PublishedBy: "user-123",
			TplID:       456,
		}

		if po.ID != 123 {
			t.Errorf("Expected ID to be 123, got %d", po.ID)
		}

		if po.Name != "Published Template" {
			t.Errorf("Expected Name to be 'Published Template', got '%s'", po.Name)
		}

		if po.Key != "published-template" {
			t.Errorf("Expected Key to be 'published-template', got '%s'", po.Key)
		}

		if po.AvatarType != cdaenum.AvatarTypeBuiltIn {
			t.Errorf("Expected AvatarType to be BuiltIn, got %v", po.AvatarType)
		}

		if po.Avatar != "default-avatar" {
			t.Errorf("Expected Avatar to be 'default-avatar', got '%s'", po.Avatar)
		}

		if po.Config != `{"key": "value"}` {
			t.Errorf("Expected Config to be '{\"key\": \"value\"}', got '%s'", po.Config)
		}

		if po.PublishedAt != 1234567890 {
			t.Errorf("Expected PublishedAt to be 1234567890, got %d", po.PublishedAt)
		}

		if po.PublishedBy != "user-123" {
			t.Errorf("Expected PublishedBy to be 'user-123', got '%s'", po.PublishedBy)
		}

		if po.TplID != 456 {
			t.Errorf("Expected TplID to be 456, got %d", po.TplID)
		}
	})

	t.Run("with nil profile", func(t *testing.T) {
		t.Parallel()

		po := &PublishedTplPo{
			ID:         789,
			Name:       "Template 2",
			Key:        "template-2",
			Profile:    nil,
			AvatarType: cdaenum.AvatarTypeUserUploaded,
		}

		if po.Profile != nil {
			t.Error("Expected Profile to be nil")
		}
	})

	t.Run("with nil IsBuiltIn", func(t *testing.T) {
		t.Parallel()

		po := &PublishedTplPo{
			ID:         999,
			Name:       "Template 3",
			IsBuiltIn:  nil,
			AvatarType: cdaenum.AvatarTypeAIGenerated,
		}

		if po.IsBuiltIn != nil {
			t.Error("Expected IsBuiltIn to be nil")
		}
	})
}
