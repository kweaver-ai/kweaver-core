package agentinoutresp

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestNewExportResp(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()

	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Agents)
	assert.IsType(t, &ExportResp{}, resp)
}

func TestExportResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := ExportResp{
		Agents: []*ExportAgentItem{
			{
				DataAgentPo: &dapo.DataAgentPo{Key: "agent-1"},
			},
			{
				DataAgentPo: &dapo.DataAgentPo{Key: "agent-2"},
			},
		},
	}

	assert.Len(t, resp.Agents, 2)
	assert.Equal(t, "agent-1", resp.Agents[0].Key)
	assert.Equal(t, "agent-2", resp.Agents[1].Key)
}

func TestExportResp_Empty(t *testing.T) {
	t.Parallel()

	resp := ExportResp{}

	assert.Nil(t, resp.Agents)
}

func TestExportAgentItem_StructFields(t *testing.T) {
	t.Parallel()

	po := &dapo.DataAgentPo{
		Key:  "agent-key-123",
		Name: "Test Agent",
	}

	item := ExportAgentItem{
		DataAgentPo: po,
	}

	assert.Equal(t, "agent-key-123", item.Key)
	assert.Equal(t, "Test Agent", item.Name)
}

func TestExportResp_AddAgent(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()
	po := &dapo.DataAgentPo{
		Key:    "agent-123",
		Name:   "Test Agent",
		Config: `{"input":{"fields":[]},"output":{}}`, // Valid JSON config
	}

	resp.AddAgent(po)

	// Agent should be added because RemoveDataSourceFromConfig succeeded
	if len(resp.Agents) > 0 {
		assert.Equal(t, "agent-123", resp.Agents[0].Key)
		assert.Equal(t, "Test Agent", resp.Agents[0].Name)
	} else {
		t.Fatal("Agent should have been added")
	}
}

func TestExportResp_AddAgent_WithInvalidConfig(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()
	// Create an agent with invalid JSON config - use a string with unmatched quotes
	po := &dapo.DataAgentPo{
		Key:    "agent-123",
		Name:   "Test Agent",
		Config: `"invalid`, // Unmatched quote - invalid JSON
	}

	resp.AddAgent(po)

	// Agent should NOT be added because RemoveDataSourceFromConfig failed
	assert.Empty(t, resp.Agents)

	// Verify that the agent list is still empty (early return happened)
	assert.Len(t, resp.Agents, 0, "Agent should not be added when RemoveDataSourceFromConfig fails")
}

func TestExportResp_AddAgent_WithEmptyConfig(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()
	// Create an agent with empty config - this should fail RemoveDataSourceFromConfig
	po := &dapo.DataAgentPo{
		Key:    "agent-123",
		Name:   "Test Agent",
		Config: "", // Empty config - will fail RemoveDataSourceFromConfig
	}

	resp.AddAgent(po)

	// Agent should NOT be added because RemoveDataSourceFromConfig failed with empty config
	assert.Len(t, resp.Agents, 0, "Agent should not be added when config is empty")
}

func TestExportResp_AddMultipleAgents(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()

	validConfig := `{"input":{"fields":[]},"output":{}}`
	agents := []*dapo.DataAgentPo{
		{Key: "agent-1", Name: "Agent 1", Config: validConfig},
		{Key: "agent-2", Name: "Agent 2", Config: validConfig},
		{Key: "agent-3", Name: "Agent 3", Config: validConfig},
	}

	for _, agent := range agents {
		resp.AddAgent(agent)
	}

	// All agents should be added
	assert.Len(t, resp.Agents, 3)
	assert.Equal(t, "agent-1", resp.Agents[0].Key)
	assert.Equal(t, "agent-2", resp.Agents[1].Key)
	assert.Equal(t, "agent-3", resp.Agents[2].Key)
}

func TestExportResp_GetSystemAgentFailItems(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()

	yes := cenum.YesNoInt8Yes
	no := cenum.YesNoInt8No
	validConfig := `{"input":{"fields":[]},"output":{}}`

	// Add system agent
	po1 := &dapo.DataAgentPo{
		Key:           "system-agent",
		Name:          "System Agent",
		IsSystemAgent: &yes,
		Config:        validConfig,
	}
	resp.AddAgent(po1)

	// Add non-system agent
	po2 := &dapo.DataAgentPo{
		Key:           "normal-agent",
		Name:          "Normal Agent",
		IsSystemAgent: &no,
		Config:        validConfig,
	}
	resp.AddAgent(po2)

	failItems := resp.GetSystemAgentFailItems()

	assert.Len(t, failItems, 1)
	assert.Equal(t, "system-agent", failItems[0].AgentKey)
	assert.Equal(t, "System Agent", failItems[0].AgentName)
}

func TestExportResp_GetSystemAgentFailItems_NoSystemAgents(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()

	no := cenum.YesNoInt8No
	validConfig := `{"input":{"fields":[]},"output":{}}`

	po := &dapo.DataAgentPo{
		Key:           "normal-agent",
		Name:          "Normal Agent",
		IsSystemAgent: &no,
		Config:        validConfig,
	}
	resp.AddAgent(po)

	failItems := resp.GetSystemAgentFailItems()

	assert.Len(t, failItems, 0)
}

func TestExportResp_GetSystemAgentFailItems_Empty(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()
	failItems := resp.GetSystemAgentFailItems()

	assert.Len(t, failItems, 0)
}

func TestExportResp_GetSystemAgentFailItems_MultipleSystemAgents(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()

	yes := cenum.YesNoInt8Yes
	validConfig := `{"input":{"fields":[]},"output":{}}`

	systemAgents := []*dapo.DataAgentPo{
		{Key: "sys-agent-1", Name: "System Agent 1", IsSystemAgent: &yes, Config: validConfig},
		{Key: "sys-agent-2", Name: "System Agent 2", IsSystemAgent: &yes, Config: validConfig},
	}

	for _, agent := range systemAgents {
		resp.AddAgent(agent)
	}

	failItems := resp.GetSystemAgentFailItems()

	assert.Len(t, failItems, 2)
}

func TestExportResp_AddAgentWithNilIsSystemAgent(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()

	validConfig := `{"input":{"fields":[]},"output":{}}`

	po := &dapo.DataAgentPo{
		Key:           "agent-123",
		Name:          "Test Agent",
		IsSystemAgent: nil,
		Config:        validConfig,
	}
	resp.AddAgent(po)

	failItems := resp.GetSystemAgentFailItems()

	assert.Len(t, failItems, 0)
}

func TestExportAgentItem_Empty(t *testing.T) {
	t.Parallel()

	item := ExportAgentItem{}

	assert.Nil(t, item.DataAgentPo)
}

func TestExportResp_AddAgentWithChineseCharacters(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()

	validConfig := `{"input":{"fields":[]},"output":{}}`

	po := &dapo.DataAgentPo{
		Key:    "中文-agent",
		Name:   "中文代理名称",
		Config: validConfig,
	}
	resp.AddAgent(po)

	// Agent should be added
	assert.Len(t, resp.Agents, 1)
	assert.Equal(t, "中文-agent", resp.Agents[0].Key)
}

func TestExportResp_GetSystemAgentFailItems_WithMixedAgents(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()

	yes := cenum.YesNoInt8Yes
	no := cenum.YesNoInt8No
	validConfig := `{"input":{"fields":[]},"output":{}}`

	agents := []*dapo.DataAgentPo{
		{Key: "normal-1", Name: "Normal 1", IsSystemAgent: &no, Config: validConfig},
		{Key: "system-1", Name: "System 1", IsSystemAgent: &yes, Config: validConfig},
		{Key: "normal-2", Name: "Normal 2", IsSystemAgent: &no, Config: validConfig},
		{Key: "system-2", Name: "System 2", IsSystemAgent: &yes, Config: validConfig},
	}

	for _, agent := range agents {
		resp.AddAgent(agent)
	}

	failItems := resp.GetSystemAgentFailItems()

	assert.Len(t, failItems, 2)
}

func TestExportResp_NewExportRespInitialization(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()

	assert.NotNil(t, resp.Agents)
	assert.Len(t, resp.Agents, 0)
}

func TestExportResp_AddAgentPreservesDataSourceRemoval(t *testing.T) {
	t.Parallel()

	resp := NewExportResp()

	validConfig := `{"input":{"fields":[]},"output":{},"data_source":{"type":"test"}}`

	po := &dapo.DataAgentPo{
		Key:    "agent-123",
		Name:   "Test Agent",
		Config: validConfig,
	}

	resp.AddAgent(po)

	// AddAgent calls RemoveDataSourceFromConfig internally
	// The method attempts to remove data source before adding
	assert.Len(t, resp.Agents, 1)
}
