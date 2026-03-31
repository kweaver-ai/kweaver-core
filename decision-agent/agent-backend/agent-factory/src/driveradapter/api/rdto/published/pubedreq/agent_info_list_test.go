package pubedreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPAInfoListReq(t *testing.T) {
	t.Parallel()

	req := NewPAInfoListReq()

	assert.NotNil(t, req)
	assert.IsType(t, &PAInfoListReq{}, req)
	assert.Empty(t, req.AgentKeys)
	assert.Empty(t, req.NeedConfigFields)
}

func TestPAInfoListReq_StructFields(t *testing.T) {
	t.Parallel()

	req := PAInfoListReq{
		AgentKeys:        []string{"agent-1", "agent-2"},
		NeedConfigFields: []string{"input"},
	}

	assert.Len(t, req.AgentKeys, 2)
	assert.Equal(t, "agent-1", req.AgentKeys[0])
	assert.Equal(t, "agent-2", req.AgentKeys[1])
	assert.Len(t, req.NeedConfigFields, 1)
	assert.Equal(t, "input", req.NeedConfigFields[0])
}

func TestPAInfoListReq_Empty(t *testing.T) {
	t.Parallel()

	req := PAInfoListReq{}

	assert.Nil(t, req.AgentKeys)
	assert.Nil(t, req.NeedConfigFields)
}

func TestPAInfoListReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := PAInfoListReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, "agent_keys is required", errMsgMap["AgentKeys.required"])
}

func TestPAInfoListReq_ReqCheck_Valid(t *testing.T) {
	t.Parallel()

	req := PAInfoListReq{
		AgentKeys:        []string{"agent-1", "agent-2"},
		NeedConfigFields: []string{"input"},
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestPAInfoListReq_ReqCheck_EmptyAgentKeys(t *testing.T) {
	t.Parallel()

	req := PAInfoListReq{
		AgentKeys: []string{},
	}

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent_keys length must be greater than 0")
}

func TestPAInfoListReq_ReqCheck_ExceedsMaxSize(t *testing.T) {
	t.Parallel()

	agentKeys := make([]string, 1001)
	for i := 0; i < 1001; i++ {
		agentKeys[i] = "agent-" + string(rune(i))
	}

	req := PAInfoListReq{
		AgentKeys: agentKeys,
	}

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent_keys length max is 1000")
}

func TestPAInfoListReq_ReqCheck_MaxSize(t *testing.T) {
	t.Parallel()

	agentKeys := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		agentKeys[i] = "agent-" + string(rune(i))
	}

	req := PAInfoListReq{
		AgentKeys: agentKeys,
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestPAInfoListReq_ReqCheck_InvalidNeedConfigFields(t *testing.T) {
	t.Parallel()

	req := PAInfoListReq{
		AgentKeys:        []string{"agent-1"},
		NeedConfigFields: []string{"invalid"},
	}

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "need_config_fields 目前只支持 [input]")
}

func TestPAInfoListReq_ReqCheck_MultipleNeedConfigFields(t *testing.T) {
	t.Parallel()

	req := PAInfoListReq{
		AgentKeys:        []string{"agent-1"},
		NeedConfigFields: []string{"input", "output"},
	}

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "need_config_fields 目前只支持 [input]")
}

func TestPAInfoListReq_ReqCheck_ValidNeedConfigFields(t *testing.T) {
	t.Parallel()

	req := PAInfoListReq{
		AgentKeys:        []string{"agent-1"},
		NeedConfigFields: []string{"input"},
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestPAInfoListReq_ReqCheck_NoNeedConfigFields(t *testing.T) {
	t.Parallel()

	req := PAInfoListReq{
		AgentKeys:        []string{"agent-1"},
		NeedConfigFields: []string{},
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestPAInfoListReq_HlDefaultVal(t *testing.T) {
	t.Parallel()

	req := PAInfoListReq{
		AgentKeys: []string{"agent-1"},
	}

	req.HlDefaultVal()

	assert.Len(t, req.NeedConfigFields, 1)
	assert.Equal(t, "input", req.NeedConfigFields[0])
}

func TestPAInfoListReq_HlDefaultVal_AlreadySet(t *testing.T) {
	t.Parallel()

	req := PAInfoListReq{
		AgentKeys:        []string{"agent-1"},
		NeedConfigFields: []string{"input"},
	}

	req.HlDefaultVal()

	assert.Len(t, req.NeedConfigFields, 1)
	assert.Equal(t, "input", req.NeedConfigFields[0])
}

func TestPAInfoListReq_WithSingleAgentKey(t *testing.T) {
	t.Parallel()

	req := PAInfoListReq{
		AgentKeys: []string{"agent-123"},
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
	assert.Len(t, req.AgentKeys, 1)
}

func TestPAInfoListReq_WithMultipleAgentKeys(t *testing.T) {
	t.Parallel()

	agentKeys := []string{
		"agent-001",
		"agent-002",
		"agent-003",
	}
	req := PAInfoListReq{
		AgentKeys: agentKeys,
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
	assert.Len(t, req.AgentKeys, 3)
}

func TestPAInfoListReq_WithSpecialAgentKeys(t *testing.T) {
	t.Parallel()

	agentKeys := []string{
		"agent-中文",
		"agent-123",
		"agent-abc-xyz",
	}
	req := PAInfoListReq{
		AgentKeys: agentKeys,
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
	assert.Len(t, req.AgentKeys, 3)
}
