package agentinoutresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewImportResp(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()

	assert.NotNil(t, resp)
	assert.NotNil(t, resp.ConfigInvalid)
	assert.NotNil(t, resp.NoCreateSystemAgentPms)
	assert.NotNil(t, resp.AgentKeyConflict)
	assert.NotNil(t, resp.BizDomainConflict)
	assert.IsType(t, &ImportResp{}, resp)
}

func TestImportResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := ImportResp{
		IsSuccess: true,
		ConfigInvalid: []*ImportFailItem{
			{AgentKey: "key-1", AgentName: "name-1"},
		},
		NoCreateSystemAgentPms: []*ImportFailItem{
			{AgentKey: "key-2", AgentName: "name-2"},
		},
		AgentKeyConflict: []*ImportFailItem{
			{AgentKey: "key-3", AgentName: "name-3"},
		},
		BizDomainConflict: []*ImportFailItem{
			{AgentKey: "key-4", AgentName: "name-4"},
		},
	}

	assert.True(t, resp.IsSuccess)
	assert.Len(t, resp.ConfigInvalid, 1)
	assert.Len(t, resp.NoCreateSystemAgentPms, 1)
	assert.Len(t, resp.AgentKeyConflict, 1)
	assert.Len(t, resp.BizDomainConflict, 1)
}

func TestImportResp_Empty(t *testing.T) {
	t.Parallel()

	resp := ImportResp{}

	assert.False(t, resp.IsSuccess)
	assert.Nil(t, resp.ConfigInvalid)
	assert.Nil(t, resp.NoCreateSystemAgentPms)
	assert.Nil(t, resp.AgentKeyConflict)
	assert.Nil(t, resp.BizDomainConflict)
}

func TestImportFailItem_StructFields(t *testing.T) {
	t.Parallel()

	item := ImportFailItem{
		AgentKey:  "agent-key-123",
		AgentName: "Agent Name",
	}

	assert.Equal(t, "agent-key-123", item.AgentKey)
	assert.Equal(t, "Agent Name", item.AgentName)
}

func TestImportFailItem_Empty(t *testing.T) {
	t.Parallel()

	item := ImportFailItem{}

	assert.Empty(t, item.AgentKey)
	assert.Empty(t, item.AgentName)
}

func TestImportResp_HasFail_WithConfigInvalid(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.ConfigInvalid = append(resp.ConfigInvalid, &ImportFailItem{
		AgentKey:  "key-1",
		AgentName: "name-1",
	})

	assert.True(t, resp.HasFail())
}

func TestImportResp_HasFail_WithNoCreateSystemAgentPms(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.NoCreateSystemAgentPms = append(resp.NoCreateSystemAgentPms, &ImportFailItem{
		AgentKey:  "key-1",
		AgentName: "name-1",
	})

	assert.True(t, resp.HasFail())
}

func TestImportResp_HasFail_WithAgentKeyConflict(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AgentKeyConflict = append(resp.AgentKeyConflict, &ImportFailItem{
		AgentKey:  "key-1",
		AgentName: "name-1",
	})

	assert.True(t, resp.HasFail())
}

func TestImportResp_HasFail_WithBizDomainConflict(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.BizDomainConflict = append(resp.BizDomainConflict, &ImportFailItem{
		AgentKey:  "key-1",
		AgentName: "name-1",
	})

	assert.True(t, resp.HasFail())
}

func TestImportResp_HasFail_NoFail(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()

	assert.False(t, resp.HasFail())
}

func TestImportResp_HasFail_MultipleFailTypes(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.ConfigInvalid = append(resp.ConfigInvalid, &ImportFailItem{
		AgentKey:  "key-1",
		AgentName: "name-1",
	})
	resp.AgentKeyConflict = append(resp.AgentKeyConflict, &ImportFailItem{
		AgentKey:  "key-2",
		AgentName: "name-2",
	})

	assert.True(t, resp.HasFail())
}

func TestImportResp_AddConfigInvalid(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AddConfigInvalid("key-1", "name-1")

	assert.Len(t, resp.ConfigInvalid, 1)
	assert.Equal(t, "key-1", resp.ConfigInvalid[0].AgentKey)
	assert.Equal(t, "name-1", resp.ConfigInvalid[0].AgentName)
}

func TestImportResp_AddConfigInvalid_Multiple(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AddConfigInvalid("key-1", "name-1")
	resp.AddConfigInvalid("key-2", "name-2")
	resp.AddConfigInvalid("key-3", "name-3")

	assert.Len(t, resp.ConfigInvalid, 3)
}

func TestImportResp_AddNoCreateSystemAgentPms(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AddNoCreateSystemAgentPms("key-1", "name-1")

	assert.Len(t, resp.NoCreateSystemAgentPms, 1)
	assert.Equal(t, "key-1", resp.NoCreateSystemAgentPms[0].AgentKey)
	assert.Equal(t, "name-1", resp.NoCreateSystemAgentPms[0].AgentName)
}

func TestImportResp_AddNoCreateSystemAgentPms_Multiple(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AddNoCreateSystemAgentPms("key-1", "name-1")
	resp.AddNoCreateSystemAgentPms("key-2", "name-2")

	assert.Len(t, resp.NoCreateSystemAgentPms, 2)
}

func TestImportResp_AddAgentKeyConflict(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AddAgentKeyConflict("key-1", "name-1")

	assert.Len(t, resp.AgentKeyConflict, 1)
	assert.Equal(t, "key-1", resp.AgentKeyConflict[0].AgentKey)
	assert.Equal(t, "name-1", resp.AgentKeyConflict[0].AgentName)
}

func TestImportResp_AddAgentKeyConflict_Multiple(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AddAgentKeyConflict("key-1", "name-1")
	resp.AddAgentKeyConflict("key-2", "name-2")

	assert.Len(t, resp.AgentKeyConflict, 2)
}

func TestImportResp_AddBizDomainConflict(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AddBizDomainConflict("key-1", "name-1")

	assert.Len(t, resp.BizDomainConflict, 1)
	assert.Equal(t, "key-1", resp.BizDomainConflict[0].AgentKey)
	assert.Equal(t, "name-1", resp.BizDomainConflict[0].AgentName)
}

func TestImportResp_AddBizDomainConflict_Multiple(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AddBizDomainConflict("key-1", "name-1")
	resp.AddBizDomainConflict("key-2", "name-2")

	assert.Len(t, resp.BizDomainConflict, 2)
}

func TestImportResp_WithChineseCharacters(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AddConfigInvalid("中文-key", "中文-name")

	assert.Len(t, resp.ConfigInvalid, 1)
	assert.Equal(t, "中文-key", resp.ConfigInvalid[0].AgentKey)
	assert.Equal(t, "中文-name", resp.ConfigInvalid[0].AgentName)
}

func TestImportResp_AllFailTypes(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AddConfigInvalid("key-1", "name-1")
	resp.AddNoCreateSystemAgentPms("key-2", "name-2")
	resp.AddAgentKeyConflict("key-3", "name-3")
	resp.AddBizDomainConflict("key-4", "name-4")

	assert.Len(t, resp.ConfigInvalid, 1)
	assert.Len(t, resp.NoCreateSystemAgentPms, 1)
	assert.Len(t, resp.AgentKeyConflict, 1)
	assert.Len(t, resp.BizDomainConflict, 1)
	assert.True(t, resp.HasFail())
}

func TestImportResp_WithEmptyStrings(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AddConfigInvalid("", "")

	assert.Len(t, resp.ConfigInvalid, 1)
	assert.Empty(t, resp.ConfigInvalid[0].AgentKey)
	assert.Empty(t, resp.ConfigInvalid[0].AgentName)
}

func TestImportResp_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()
	resp.AddConfigInvalid("key-@#$%", "name-!@#$%")

	assert.Len(t, resp.ConfigInvalid, 1)
	assert.Equal(t, "key-@#$%", resp.ConfigInvalid[0].AgentKey)
	assert.Equal(t, "name-!@#$%", resp.ConfigInvalid[0].AgentName)
}

func TestImportResp_IsSuccessTrue(t *testing.T) {
	t.Parallel()

	resp := ImportResp{
		IsSuccess: true,
	}

	assert.True(t, resp.IsSuccess)
}

func TestImportResp_IsSuccessFalse(t *testing.T) {
	t.Parallel()

	resp := ImportResp{
		IsSuccess: false,
	}

	assert.False(t, resp.IsSuccess)
}

func TestImportResp_WithAllFailTypesMultipleItems(t *testing.T) {
	t.Parallel()

	resp := NewImportResp()

	// Add multiple items to each fail type
	for i := 1; i <= 3; i++ {
		resp.AddConfigInvalid("config-"+string(rune(i)), "name-"+string(rune(i)))
		resp.AddNoCreateSystemAgentPms("perm-"+string(rune(i)), "name-"+string(rune(i)))
		resp.AddAgentKeyConflict("conflict-"+string(rune(i)), "name-"+string(rune(i)))
		resp.AddBizDomainConflict("domain-"+string(rune(i)), "name-"+string(rune(i)))
	}

	assert.Len(t, resp.ConfigInvalid, 3)
	assert.Len(t, resp.NoCreateSystemAgentPms, 3)
	assert.Len(t, resp.AgentKeyConflict, 3)
	assert.Len(t, resp.BizDomainConflict, 3)
	assert.True(t, resp.HasFail())
}
