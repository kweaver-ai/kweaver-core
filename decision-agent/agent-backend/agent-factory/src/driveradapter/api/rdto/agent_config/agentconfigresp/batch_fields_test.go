package agentconfigresp

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBatchFieldsRespField(t *testing.T) {
	t.Parallel()

	field := NewBatchFieldsRespField()

	assert.NotNil(t, field)
	assert.IsType(t, &BatchFieldsRespField{}, field)
}

func TestBatchFieldsRespField_StructFields(t *testing.T) {
	t.Parallel()

	field := &BatchFieldsRespField{
		Name: "TestAgent",
	}

	assert.Equal(t, "TestAgent", field.Name)
}

func TestBatchFieldsRespField_Empty(t *testing.T) {
	t.Parallel()

	field := &BatchFieldsRespField{}

	assert.Empty(t, field.Name)
}

func TestAgentFieldsItem_StructFields(t *testing.T) {
	t.Parallel()

	field := &BatchFieldsRespField{
		Name: "TestAgent",
	}
	item := &AgentFieldsItem{
		AgentID: "agent-123",
		Field:   field,
	}

	assert.Equal(t, "agent-123", item.AgentID)
	assert.Equal(t, field, item.Field)
	assert.Equal(t, "TestAgent", item.Field.Name)
}

func TestAgentFieldsItem_NilField(t *testing.T) {
	t.Parallel()

	item := &AgentFieldsItem{
		AgentID: "agent-456",
		Field:   nil,
	}

	assert.Equal(t, "agent-456", item.AgentID)
	assert.Nil(t, item.Field)
}

func TestNewBatchFieldsResp(t *testing.T) {
	t.Parallel()

	resp := NewBatchFieldsResp()

	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Entries)
	assert.Empty(t, resp.Entries)
	assert.IsType(t, []*AgentFieldsItem{}, resp.Entries)
}

func TestBatchFieldsResp_StructFields(t *testing.T) {
	t.Parallel()

	items := []*AgentFieldsItem{
		{
			AgentID: "agent-1",
			Field:   &BatchFieldsRespField{Name: "Agent1"},
		},
		{
			AgentID: "agent-2",
			Field:   &BatchFieldsRespField{Name: "Agent2"},
		},
	}
	resp := &BatchFieldsResp{
		Entries: items,
	}

	assert.Len(t, resp.Entries, 2)
	assert.Equal(t, "agent-1", resp.Entries[0].AgentID)
	assert.Equal(t, "Agent1", resp.Entries[0].Field.Name)
	assert.Equal(t, "agent-2", resp.Entries[1].AgentID)
	assert.Equal(t, "Agent2", resp.Entries[1].Field.Name)
}

func TestBatchFieldsResp_Empty(t *testing.T) {
	t.Parallel()

	resp := &BatchFieldsResp{
		Entries: []*AgentFieldsItem{},
	}

	assert.Empty(t, resp.Entries)
}

func TestBatchFieldsResp_LoadFromAgentPOs(t *testing.T) {
	t.Parallel()

	pos := []*dapo.DataAgentPo{
		{
			ID:   "agent-1",
			Name: "Agent One",
		},
		{
			ID:   "agent-2",
			Name: "Agent Two",
		},
		{
			ID:   "agent-3",
			Name: "Agent Three",
		},
	}
	requestedFields := []agentconfigreq.BatchFieldsReqField{
		agentconfigreq.BatchFieldsReqFieldName,
	}

	resp := NewBatchFieldsResp()
	err := resp.LoadFromAgentPOs(pos, requestedFields)

	require.NoError(t, err)
	assert.Len(t, resp.Entries, 3)
	assert.Equal(t, "agent-1", resp.Entries[0].AgentID)
	assert.Equal(t, "Agent One", resp.Entries[0].Field.Name)
	assert.Equal(t, "agent-2", resp.Entries[1].AgentID)
	assert.Equal(t, "Agent Two", resp.Entries[1].Field.Name)
	assert.Equal(t, "agent-3", resp.Entries[2].AgentID)
	assert.Equal(t, "Agent Three", resp.Entries[2].Field.Name)
}

func TestBatchFieldsResp_LoadFromAgentPOs_Empty(t *testing.T) {
	t.Parallel()

	pos := []*dapo.DataAgentPo{}
	requestedFields := []agentconfigreq.BatchFieldsReqField{
		agentconfigreq.BatchFieldsReqFieldName,
	}

	resp := NewBatchFieldsResp()
	err := resp.LoadFromAgentPOs(pos, requestedFields)

	require.NoError(t, err)
	assert.Empty(t, resp.Entries)
}

func TestBatchFieldsResp_LoadFromAgentPOs_NoRequestedFields(t *testing.T) {
	t.Parallel()

	pos := []*dapo.DataAgentPo{
		{
			ID:   "agent-1",
			Name: "Agent One",
		},
	}
	requestedFields := []agentconfigreq.BatchFieldsReqField{}

	resp := NewBatchFieldsResp()
	err := resp.LoadFromAgentPOs(pos, requestedFields)

	require.NoError(t, err)
	assert.Len(t, resp.Entries, 1)
	assert.Equal(t, "agent-1", resp.Entries[0].AgentID)
	assert.Empty(t, resp.Entries[0].Field.Name)
}

func TestBatchFieldsResp_LoadFromAgentPOs_WithMultipleFields(t *testing.T) {
	t.Parallel()

	pos := []*dapo.DataAgentPo{
		{
			ID:   "agent-1",
			Name: "MultiFieldAgent",
		},
	}
	requestedFields := []agentconfigreq.BatchFieldsReqField{
		agentconfigreq.BatchFieldsReqFieldName,
		agentconfigreq.BatchFieldsReqFieldName,
	}

	resp := NewBatchFieldsResp()
	err := resp.LoadFromAgentPOs(pos, requestedFields)

	require.NoError(t, err)
	assert.Len(t, resp.Entries, 1)
	// Name should be set twice (last value wins)
	assert.Equal(t, "MultiFieldAgent", resp.Entries[0].Field.Name)
}

func TestBatchFieldsResp_Append(t *testing.T) {
	t.Parallel()

	resp := NewBatchFieldsResp()

	resp.Entries = append(resp.Entries, &AgentFieldsItem{
		AgentID: "agent-1",
		Field:   &BatchFieldsRespField{Name: "Agent1"},
	})

	resp.Entries = append(resp.Entries, &AgentFieldsItem{
		AgentID: "agent-2",
		Field:   &BatchFieldsRespField{Name: "Agent2"},
	})

	assert.Len(t, resp.Entries, 2)
	assert.Equal(t, "agent-1", resp.Entries[0].AgentID)
	assert.Equal(t, "agent-2", resp.Entries[1].AgentID)
}
