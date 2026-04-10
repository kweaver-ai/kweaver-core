package dapo

import (
	"database/sql"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/daenum"
)

func TestPublishedToBeStruct_SelectFieldsZero(t *testing.T) {
	t.Parallel()

	t.Run("select fields zero", func(t *testing.T) {
		t.Parallel()

		structVal := &PublishedToBeStruct{}
		result := structVal.SelectFieldsZero()

		expected := "0 as f_is_api_agent, 0 as f_is_web_sdk_agent, 0 as f_is_skill_agent, 0 as f_is_data_flow_agent"
		if result != expected {
			t.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}

func TestPublishedToBeStruct_LoadFromReleasePo(t *testing.T) {
	t.Parallel()

	t.Run("nil po", func(t *testing.T) {
		t.Parallel()

		structVal := &PublishedToBeStruct{}
		structVal.LoadFromReleasePo(nil)

		// Should not panic with nil po
		if structVal.IsAPIAgent != 0 {
			t.Errorf("Expected IsAPIAgent to remain 0, got %d", structVal.IsAPIAgent)
		}
	})

	t.Run("with all fields set", func(t *testing.T) {
		t.Parallel()

		isAPI := 1
		isWebSDK := 0
		isSkill := 1
		isDataFlow := 0

		po := &ReleasePO{
			IsAPIAgent:      &isAPI,
			IsWebSDKAgent:   &isWebSDK,
			IsSkillAgent:    &isSkill,
			IsDataFlowAgent: &isDataFlow,
		}

		structVal := &PublishedToBeStruct{}
		structVal.LoadFromReleasePo(po)

		if structVal.IsAPIAgent != 1 {
			t.Errorf("Expected IsAPIAgent to be 1, got %d", structVal.IsAPIAgent)
		}

		if structVal.IsWebSDKAgent != 0 {
			t.Errorf("Expected IsWebSDKAgent to be 0, got %d", structVal.IsWebSDKAgent)
		}

		if structVal.IsSkillAgent != 1 {
			t.Errorf("Expected IsSkillAgent to be 1, got %d", structVal.IsSkillAgent)
		}

		if structVal.IsDataFlowAgent != 0 {
			t.Errorf("Expected IsDataFlowAgent to be 0, got %d", structVal.IsDataFlowAgent)
		}
	})
}

func TestReleasePO_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePO{}
		tableName := po.TableName()

		expected := "t_data_agent_release"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}

func TestReleasePO_IsAPIAgentBool(t *testing.T) {
	t.Parallel()

	t.Run("nil field", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePO{IsAPIAgent: nil}
		result := po.IsAPIAgentBool()

		if result {
			t.Error("Expected IsAPIAgentBool to return false for nil field")
		}
	})

	t.Run("true value", func(t *testing.T) {
		t.Parallel()

		val := 1
		po := &ReleasePO{IsAPIAgent: &val}
		result := po.IsAPIAgentBool()

		if !result {
			t.Error("Expected IsAPIAgentBool to return true")
		}
	})

	t.Run("false value", func(t *testing.T) {
		t.Parallel()

		val := 0
		po := &ReleasePO{IsAPIAgent: &val}
		result := po.IsAPIAgentBool()

		if result {
			t.Error("Expected IsAPIAgentBool to return false")
		}
	})
}

func TestReleasePO_IsWebSDKAgentBool(t *testing.T) {
	t.Parallel()

	t.Run("nil field", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePO{IsWebSDKAgent: nil}
		result := po.IsWebSDKAgentBool()

		if result {
			t.Error("Expected IsWebSDKAgentBool to return false for nil field")
		}
	})

	t.Run("true value", func(t *testing.T) {
		t.Parallel()

		val := 1
		po := &ReleasePO{IsWebSDKAgent: &val}
		result := po.IsWebSDKAgentBool()

		if !result {
			t.Error("Expected IsWebSDKAgentBool to return true")
		}
	})
}

func TestReleasePO_IsSkillAgentBool(t *testing.T) {
	t.Parallel()

	t.Run("nil field", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePO{IsSkillAgent: nil}
		result := po.IsSkillAgentBool()

		if result {
			t.Error("Expected IsSkillAgentBool to return false for nil field")
		}
	})

	t.Run("true value", func(t *testing.T) {
		t.Parallel()

		val := 1
		po := &ReleasePO{IsSkillAgent: &val}
		result := po.IsSkillAgentBool()

		if !result {
			t.Error("Expected IsSkillAgentBool to return true")
		}
	})
}

func TestReleasePO_IsDataFlowAgentBool(t *testing.T) {
	t.Parallel()

	t.Run("nil field", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePO{IsDataFlowAgent: nil}
		result := po.IsDataFlowAgentBool()

		if result {
			t.Error("Expected IsDataFlowAgentBool to return false for nil field")
		}
	})

	t.Run("true value", func(t *testing.T) {
		t.Parallel()

		val := 1
		po := &ReleasePO{IsDataFlowAgent: &val}
		result := po.IsDataFlowAgentBool()

		if !result {
			t.Error("Expected IsDataFlowAgentBool to return true")
		}
	})
}

func TestReleasePO_ResetPublishToBes(t *testing.T) {
	t.Parallel()

	t.Run("reset all to zero", func(t *testing.T) {
		t.Parallel()

		val := 1
		po := &ReleasePO{
			IsAPIAgent:      &val,
			IsWebSDKAgent:   &val,
			IsSkillAgent:    &val,
			IsDataFlowAgent: &val,
		}

		po.ResetPublishToBes()

		if po.IsAPIAgent == nil || *po.IsAPIAgent != 0 {
			t.Error("Expected IsAPIAgent to be set to 0")
		}

		if po.IsWebSDKAgent == nil || *po.IsWebSDKAgent != 0 {
			t.Error("Expected IsWebSDKAgent to be set to 0")
		}

		if po.IsSkillAgent == nil || *po.IsSkillAgent != 0 {
			t.Error("Expected IsSkillAgent to be set to 0")
		}

		if po.IsDataFlowAgent == nil || *po.IsDataFlowAgent != 0 {
			t.Error("Expected IsDataFlowAgent to be set to 0")
		}
	})
}

func TestReleasePO_SetPublishToBes(t *testing.T) {
	t.Parallel()

	t.Run("set all types", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePO{}
		toBes := []cdaenum.PublishToBe{
			cdaenum.PublishToBeAPIAgent,
			cdaenum.PublishToBeWebSDKAgent,
			cdaenum.PublishToBeSkillAgent,
			cdaenum.PublishToBeDataFlowAgent,
		}

		po.SetPublishToBes(toBes)

		if po.IsAPIAgent == nil || *po.IsAPIAgent != 1 {
			t.Error("Expected IsAPIAgent to be set to 1")
		}

		if po.IsWebSDKAgent == nil || *po.IsWebSDKAgent != 1 {
			t.Error("Expected IsWebSDKAgent to be set to 1")
		}

		if po.IsSkillAgent == nil || *po.IsSkillAgent != 1 {
			t.Error("Expected IsSkillAgent to be set to 1")
		}

		if po.IsDataFlowAgent == nil || *po.IsDataFlowAgent != 1 {
			t.Error("Expected IsDataFlowAgent to be set to 1")
		}
	})
}

func TestReleasePO_IsToCustomSpaceBool(t *testing.T) {
	t.Parallel()

	t.Run("nil field", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePO{IsToCustomSpace: nil}
		result := po.IsToCustomSpaceBool()

		if result {
			t.Error("Expected IsToCustomSpaceBool to return false for nil field")
		}
	})

	t.Run("true value", func(t *testing.T) {
		t.Parallel()

		val := 1
		po := &ReleasePO{IsToCustomSpace: &val}
		result := po.IsToCustomSpaceBool()

		if !result {
			t.Error("Expected IsToCustomSpaceBool to return true")
		}
	})
}

func TestReleasePO_IsToSquareBool(t *testing.T) {
	t.Parallel()

	t.Run("nil field", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePO{IsToSquare: nil}
		result := po.IsToSquareBool()

		if result {
			t.Error("Expected IsToSquareBool to return false for nil field")
		}
	})

	t.Run("true value", func(t *testing.T) {
		t.Parallel()

		val := 1
		po := &ReleasePO{IsToSquare: &val}
		result := po.IsToSquareBool()

		if !result {
			t.Error("Expected IsToSquareBool to return true")
		}
	})
}

func TestReleasePO_ResetPublishToWhere(t *testing.T) {
	t.Parallel()

	t.Run("reset all to zero", func(t *testing.T) {
		t.Parallel()

		val := 1
		po := &ReleasePO{
			IsToCustomSpace: &val,
			IsToSquare:      &val,
		}

		po.ResetPublishToWhere()

		if po.IsToCustomSpace == nil || *po.IsToCustomSpace != 0 {
			t.Error("Expected IsToCustomSpace to be set to 0")
		}

		if po.IsToSquare == nil || *po.IsToSquare != 0 {
			t.Error("Expected IsToSquare to be set to 0")
		}
	})
}

func TestReleasePO_SetPublishToWhere(t *testing.T) {
	t.Parallel()

	t.Run("set both types", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePO{}
		tos := []daenum.PublishToWhere{
			daenum.PublishToWhereCustomSpace,
			daenum.PublishToWhereSquare,
		}

		po.SetPublishToWhere(tos)

		if po.IsToCustomSpace == nil || *po.IsToCustomSpace != 1 {
			t.Error("Expected IsToCustomSpace to be set to 1")
		}

		if po.IsToSquare == nil || *po.IsToSquare != 1 {
			t.Error("Expected IsToSquare to be set to 1")
		}
	})
}

func TestReleasePO_IsPmsCtrlBool(t *testing.T) {
	t.Parallel()

	t.Run("nil field", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePO{IsPmsCtrl: nil}
		result := po.IsPmsCtrlBool()

		if result {
			t.Error("Expected IsPmsCtrlBool to return false for nil field")
		}
	})

	t.Run("true value", func(t *testing.T) {
		t.Parallel()

		val := 1
		po := &ReleasePO{IsPmsCtrl: &val}
		result := po.IsPmsCtrlBool()

		if !result {
			t.Error("Expected IsPmsCtrlBool to return true")
		}
	})

	t.Run("false value", func(t *testing.T) {
		t.Parallel()

		val := 0
		po := &ReleasePO{IsPmsCtrl: &val}
		result := po.IsPmsCtrlBool()

		if result {
			t.Error("Expected IsPmsCtrlBool to return false")
		}
	})
}

func TestReleasePO_ResetIsPmsCtrl(t *testing.T) {
	t.Parallel()

	t.Run("reset to zero", func(t *testing.T) {
		t.Parallel()

		val := 1
		po := &ReleasePO{IsPmsCtrl: &val}

		po.ResetIsPmsCtrl()

		if po.IsPmsCtrl == nil || *po.IsPmsCtrl != 0 {
			t.Error("Expected IsPmsCtrl to be set to 0")
		}
	})
}

func TestReleasePO_SetIsPmsCtrl(t *testing.T) {
	t.Parallel()

	t.Run("set to true", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePO{}
		po.SetIsPmsCtrl(true)

		if po.IsPmsCtrl == nil || *po.IsPmsCtrl != 1 {
			t.Error("Expected IsPmsCtrl to be set to 1")
		}
	})

	t.Run("set to false", func(t *testing.T) {
		t.Parallel()

		val := 1
		po := &ReleasePO{IsPmsCtrl: &val}
		po.SetIsPmsCtrl(false)

		if po.IsPmsCtrl == nil || *po.IsPmsCtrl != 0 {
			t.Error("Expected IsPmsCtrl to be set to 0")
		}
	})
}

func TestReleaseAgentPO(t *testing.T) {
	t.Parallel()

	t.Run("create ReleaseAgentPO", func(t *testing.T) {
		t.Parallel()

		config := "test config"
		desc := "test description"
		version := "v1.0"
		po := &ReleaseAgentPO{
			DataAgentPo: DataAgentPo{
				ID: "agent-123",
			},
			AgentConfig:  sql.NullString{String: config, Valid: true},
			AgentDesc:    sql.NullString{String: desc, Valid: true},
			AgentVersion: sql.NullString{String: version, Valid: true},
		}

		if po.ID != "agent-123" {
			t.Errorf("Expected ID to be 'agent-123', got '%s'", po.ID)
		}

		if po.AgentConfig.String != config {
			t.Errorf("Expected AgentConfig to be '%s', got '%s'", config, po.AgentConfig.String)
		}
	})
}

func TestRecentVisitAgentPO(t *testing.T) {
	t.Parallel()

	t.Run("create RecentVisitAgentPO", func(t *testing.T) {
		t.Parallel()

		po := &RecentVisitAgentPO{
			ReleaseAgentPO: ReleaseAgentPO{
				DataAgentPo: DataAgentPo{
					ID: "agent-123",
				},
			},
			PublishedToBeStruct: PublishedToBeStruct{
				IsAPIAgent: 1,
			},
		}

		if po.ID != "agent-123" {
			t.Errorf("Expected ID to be 'agent-123', got '%s'", po.ID)
		}

		if po.IsAPIAgent != 1 {
			t.Errorf("Expected IsAPIAgent to be 1, got %d", po.IsAPIAgent)
		}
	})
}

func TestReleasePartPo_IsPmsCtrlBool(t *testing.T) {
	t.Parallel()

	t.Run("is pms ctrl is 1", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePartPo{IsPmsCtrl: 1}
		result := po.IsPmsCtrlBool()

		if !result {
			t.Error("Expected IsPmsCtrlBool to return true when IsPmsCtrl is 1")
		}
	})

	t.Run("is pms ctrl is 0", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePartPo{IsPmsCtrl: 0}
		result := po.IsPmsCtrlBool()

		if result {
			t.Error("Expected IsPmsCtrlBool to return false when IsPmsCtrl is 0")
		}
	})
}

func TestPublishedJoinPo_LoadFromReleasePartPo(t *testing.T) {
	t.Parallel()

	t.Run("valid po with empty agent config", func(t *testing.T) {
		t.Parallel()

		partPo := &ReleasePartPo{
			ReleaseID:   "release-123",
			PublishDesc: "test description",
			Version:     "v1.0.0",
			PublishedAt: 1234567890,
			PublishedBy: "user-1",
			IsPmsCtrl:   1,
			AgentConfig: `{}`, // Empty but valid JSON
		}

		joinPo := &PublishedJoinPo{}

		err := joinPo.LoadFromReleasePartPo(partPo)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("po with agent config", func(t *testing.T) {
		t.Parallel()

		partPo := &ReleasePartPo{
			ReleaseID:   "release-456",
			AgentConfig: `{"id": "agent-123", "name": "Test Agent"}`,
		}

		joinPo := &PublishedJoinPo{}

		err := joinPo.LoadFromReleasePartPo(partPo)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if joinPo.ID != "agent-123" {
			t.Errorf("Expected ID to be 'agent-123', got '%s'", joinPo.ID)
		}

		if joinPo.Name != "Test Agent" {
			t.Errorf("Expected Name to be 'Test Agent', got '%s'", joinPo.Name)
		}
	})

	t.Run("po with invalid agent config", func(t *testing.T) {
		t.Parallel()

		partPo := &ReleasePartPo{
			ReleaseID:   "release-789",
			AgentConfig: `{invalid json`, // Invalid JSON
		}

		joinPo := &PublishedJoinPo{}
		err := joinPo.LoadFromReleasePartPo(partPo)

		if err == nil {
			t.Error("Expected error when loading from invalid agent config")
		}
	})
}

func TestReleasePermissionPO_TableName(t *testing.T) {
	t.Parallel()

	t.Run("table name", func(t *testing.T) {
		t.Parallel()

		po := &ReleasePermissionPO{}
		tableName := po.TableName()

		expected := "t_data_agent_release_permission"
		if tableName != expected {
			t.Errorf("Expected table name to be '%s', got '%s'", expected, tableName)
		}
	})
}
